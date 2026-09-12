package events

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const maxEventLineBytes = 1 << 20

type Event struct {
	ID        string     `json:"id"`
	CameraID  string     `json:"camera_id"`
	Type      string     `json:"type"`
	Label     string     `json:"label,omitempty"`
	Score     *float64   `json:"score,omitempty"`
	Zones     []string   `json:"zones,omitempty"`
	StartedAt time.Time  `json:"started_at"`
	EndedAt   *time.Time `json:"ended_at,omitempty"`
}

type Journal struct {
	mu   sync.Mutex
	path string
}

func NewJournal(path string) (*Journal, error) {
	if path == "" {
		return nil, errors.New("journal path must not be empty")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return nil, fmt.Errorf("create journal directory: %w", err)
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return nil, fmt.Errorf("open journal: %w", err)
	}
	if err := f.Close(); err != nil {
		return nil, fmt.Errorf("close journal: %w", err)
	}
	if err := os.Chmod(path, 0o600); err != nil {
		return nil, fmt.Errorf("restrict journal permissions: %w", err)
	}
	return &Journal{path: path}, nil
}

func (j *Journal) Append(event Event) error {
	payload, err := encodeEvent(event)
	if err != nil {
		return err
	}

	j.mu.Lock()
	defer j.mu.Unlock()
	return j.appendEncodedLocked(payload)
}

// AppendUnique appends an event only when its durable ID does not already
// exist anywhere in the journal. The complete journal is scanned under the same
// mutex as the append, so restart/replay of the same deterministic event ID is
// idempotent within one Journal authority. Malformed historical data fails
// closed instead of being skipped.
func (j *Journal) AppendUnique(event Event) (bool, error) {
	payload, err := encodeEvent(event)
	if err != nil {
		return false, err
	}

	j.mu.Lock()
	defer j.mu.Unlock()

	all, err := j.loadAllLocked()
	if err != nil {
		return false, err
	}
	for _, existing := range all {
		if existing.ID == event.ID {
			return false, nil
		}
	}
	if err := j.appendEncodedLocked(payload); err != nil {
		return false, err
	}
	return true, nil
}

func (j *Journal) List(limit int) ([]Event, error) {
	if limit <= 0 || limit > 1000 {
		return nil, errors.New("limit must be between 1 and 1000")
	}

	j.mu.Lock()
	defer j.mu.Unlock()
	all, err := j.loadAllLocked()
	if err != nil {
		return nil, err
	}
	if len(all) > limit {
		all = all[len(all)-limit:]
	}
	return all, nil
}

// LastForCamera returns the newest durable event matching a camera and event
// type while scanning the complete journal with bounded line sizes. It does not
// expose raw media state.
func (j *Journal) LastForCamera(cameraID, eventType string) (*Event, error) {
	if cameraID == "" || eventType == "" {
		return nil, errors.New("camera id and event type are required")
	}
	j.mu.Lock()
	defer j.mu.Unlock()

	all, err := j.loadAllLocked()
	if err != nil {
		return nil, err
	}
	for index := len(all) - 1; index >= 0; index-- {
		if all[index].CameraID == cameraID && all[index].Type == eventType {
			copy := all[index]
			return &copy, nil
		}
	}
	return nil, nil
}

// PruneBefore removes durable events whose retention timestamp is strictly
// older than cutoff. EndedAt is used when present; otherwise StartedAt is the
// retention timestamp. The complete journal must decode and validate before any
// replacement occurs. A same-directory temporary file is fsynced, atomically
// renamed over the journal, and the containing directory is then fsynced.
func (j *Journal) PruneBefore(cutoff time.Time) (int, error) {
	if cutoff.IsZero() {
		return 0, errors.New("retention cutoff must not be zero")
	}
	cutoff = cutoff.UTC()

	j.mu.Lock()
	defer j.mu.Unlock()

	all, err := j.loadAllLocked()
	if err != nil {
		return 0, err
	}
	kept := make([]Event, 0, len(all))
	removed := 0
	for _, event := range all {
		retentionAt := event.StartedAt
		if event.EndedAt != nil {
			retentionAt = *event.EndedAt
		}
		if retentionAt.Before(cutoff) {
			removed++
			continue
		}
		kept = append(kept, event)
	}
	if removed == 0 {
		return 0, nil
	}
	if err := j.replaceAllLocked(kept); err != nil {
		return 0, err
	}
	return removed, nil
}

func (j *Journal) appendEncodedLocked(payload []byte) error {
	f, err := os.OpenFile(j.path, os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return fmt.Errorf("open journal: %w", err)
	}
	defer f.Close()
	if _, err := f.Write(append(payload, '\n')); err != nil {
		return fmt.Errorf("append event: %w", err)
	}
	if err := f.Sync(); err != nil {
		return fmt.Errorf("sync journal: %w", err)
	}
	return nil
}

func (j *Journal) replaceAllLocked(events []Event) error {
	dir := filepath.Dir(j.path)
	temp, err := os.CreateTemp(dir, ".events-retention-*")
	if err != nil {
		return fmt.Errorf("create retention journal: %w", err)
	}
	tempPath := temp.Name()
	committed := false
	defer func() {
		if !committed {
			_ = temp.Close()
			_ = os.Remove(tempPath)
		}
	}()
	if err := temp.Chmod(0o600); err != nil {
		return fmt.Errorf("restrict retention journal permissions: %w", err)
	}
	writer := bufio.NewWriter(temp)
	for _, event := range events {
		payload, err := encodeEvent(event)
		if err != nil {
			return fmt.Errorf("encode retained event: %w", err)
		}
		if _, err := writer.Write(append(payload, '\n')); err != nil {
			return fmt.Errorf("write retention journal: %w", err)
		}
	}
	if err := writer.Flush(); err != nil {
		return fmt.Errorf("flush retention journal: %w", err)
	}
	if err := temp.Sync(); err != nil {
		return fmt.Errorf("sync retention journal: %w", err)
	}
	if err := temp.Close(); err != nil {
		return fmt.Errorf("close retention journal: %w", err)
	}
	if err := os.Rename(tempPath, j.path); err != nil {
		return fmt.Errorf("replace retained journal: %w", err)
	}
	directory, err := os.Open(dir)
	if err != nil {
		return fmt.Errorf("open journal directory for sync: %w", err)
	}
	if err := directory.Sync(); err != nil {
		_ = directory.Close()
		return fmt.Errorf("sync journal directory: %w", err)
	}
	if err := directory.Close(); err != nil {
		return fmt.Errorf("close journal directory: %w", err)
	}
	committed = true
	return nil
}

func (j *Journal) loadAllLocked() ([]Event, error) {
	f, err := os.Open(j.path)
	if err != nil {
		return nil, fmt.Errorf("open journal: %w", err)
	}
	defer f.Close()

	all := make([]Event, 0)
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 64*1024), maxEventLineBytes)
	line := 0
	for scanner.Scan() {
		line++
		var event Event
		if err := json.Unmarshal(scanner.Bytes(), &event); err != nil {
			return nil, fmt.Errorf("decode journal line %d: %w", line, err)
		}
		if err := event.Validate(); err != nil {
			return nil, fmt.Errorf("validate journal line %d: %w", line, err)
		}
		all = append(all, event)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan journal: %w", err)
	}
	return all, nil
}

func encodeEvent(event Event) ([]byte, error) {
	if err := event.Validate(); err != nil {
		return nil, err
	}
	payload, err := json.Marshal(event)
	if err != nil {
		return nil, fmt.Errorf("encode event: %w", err)
	}
	return payload, nil
}

func (e Event) Validate() error {
	if e.ID == "" {
		return errors.New("event id must not be empty")
	}
	if e.CameraID == "" {
		return errors.New("camera id must not be empty")
	}
	switch e.Type {
	case "motion", "object", "camera":
	default:
		return fmt.Errorf("unsupported event type %q", e.Type)
	}
	if e.StartedAt.IsZero() {
		return errors.New("started_at must not be zero")
	}
	if e.Score != nil && (*e.Score < 0 || *e.Score > 1) {
		return errors.New("score must be between 0 and 1")
	}
	if e.EndedAt != nil && e.EndedAt.Before(e.StartedAt) {
		return errors.New("ended_at must not precede started_at")
	}
	return nil
}
