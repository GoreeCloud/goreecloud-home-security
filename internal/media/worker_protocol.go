package media

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
)

const (
	WorkerEventVersion    = 1
	workerEventMaxBytes   = 4096
	WorkerEventMediaReady = "media_ready"
)

type WorkerEvent struct {
	Version int    `json:"version"`
	Type    string `json:"type"`
}

func (e WorkerEvent) Validate() error {
	if e.Version != WorkerEventVersion {
		return errors.New("unsupported worker event version")
	}
	if e.Type != WorkerEventMediaReady {
		return errors.New("unsupported worker event type")
	}
	return nil
}

func EncodeWorkerEvent(w io.Writer, e WorkerEvent) error {
	if w == nil {
		return errors.New("worker event writer must not be nil")
	}
	if err := e.Validate(); err != nil {
		return err
	}
	payload, err := json.Marshal(e)
	if err != nil {
		return fmt.Errorf("encode worker event: %w", err)
	}
	if len(payload) > workerEventMaxBytes {
		return errors.New("worker event exceeds size limit")
	}
	_, err = w.Write(append(payload, '\n'))
	return err
}

func DecodeWorkerEvent(payload []byte) (WorkerEvent, error) {
	if len(payload) == 0 || len(payload) > workerEventMaxBytes {
		return WorkerEvent{}, errors.New("worker event size is invalid")
	}
	dec := json.NewDecoder(strings.NewReader(string(payload)))
	dec.DisallowUnknownFields()
	var e WorkerEvent
	if err := dec.Decode(&e); err != nil {
		return WorkerEvent{}, err
	}
	var extra any
	if err := dec.Decode(&extra); !errors.Is(err, io.EOF) {
		return WorkerEvent{}, errors.New("worker event must contain one JSON value")
	}
	if err := e.Validate(); err != nil {
		return WorkerEvent{}, err
	}
	return e, nil
}
