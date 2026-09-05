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
	if err := event.Validate(); err != nil {
		return err
	}
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("encode event: %w", err)
	}

	j.mu.Lock()
	defer j.mu.Unlock()

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

func (j *Journal) List(limit int) ([]Event, error) {
	if limit <= 0 || limit > 1000 {
		return nil, errors.New("limit must be between 1 and 1000")
	}

	j.mu.Lock()
	defer j.mu.Unlock()

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

	if len(all) > limit {
		all = all[len(all)-limit:]
	}
	return all, nil
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
