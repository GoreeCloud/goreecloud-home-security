package media

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"regexp"
	"strings"

	"github.com/GoreeCloud/goreecloud-home-security/internal/config"
)

const (
	WorkerDescriptorVersion  = 1
	workerDescriptorMaxBytes = 64 << 10
	WorkerDescriptorFD       = 3
)

var workerCameraIDPattern = regexp.MustCompile(`^[a-z][a-z0-9-]{0,62}$`)

type SecretLookup func(name string) (string, bool)

type WorkerDescriptor struct {
	Version   int    `json:"version"`
	CameraID  string `json:"camera_id"`
	StreamURL string `json:"stream_url"`
	Username  string `json:"username,omitempty"`
	Password  string `json:"password,omitempty"`
}

func ResolveWorkerDescriptor(camera config.Camera, lookup SecretLookup) (WorkerDescriptor, error) {
	if err := camera.Validate(); err != nil {
		return WorkerDescriptor{}, err
	}
	d := WorkerDescriptor{Version: WorkerDescriptorVersion, CameraID: camera.ID, StreamURL: camera.StreamURL}
	if camera.UsernameEnv == "" {
		if err := d.Validate(); err != nil {
			return WorkerDescriptor{}, err
		}
		return d, nil
	}
	if lookup == nil {
		return WorkerDescriptor{}, &ProbeError{Code: ReasonCredentialUnavailable}
	}
	username, okUser := lookup(camera.UsernameEnv)
	password, okPass := lookup(camera.PasswordEnv)
	if !okUser || !okPass || username == "" || password == "" {
		return WorkerDescriptor{}, &ProbeError{Code: ReasonCredentialUnavailable}
	}
	d.Username, d.Password = username, password
	if err := d.Validate(); err != nil {
		return WorkerDescriptor{}, err
	}
	return d, nil
}

func (d WorkerDescriptor) Validate() error {
	if d.Version != WorkerDescriptorVersion {
		return errors.New("unsupported worker descriptor version")
	}
	if !workerCameraIDPattern.MatchString(d.CameraID) {
		return errors.New("worker camera id is invalid")
	}
	if len(d.StreamURL) == 0 || len(d.StreamURL) > 8192 {
		return errors.New("worker stream url length is invalid")
	}
	u, err := url.Parse(d.StreamURL)
	if err != nil {
		return errors.New("worker stream url is invalid")
	}
	if (u.Scheme != "rtsp" && u.Scheme != "rtsps") || u.Host == "" || u.User != nil {
		return errors.New("worker stream url must be credential-free rtsp/rtsps")
	}
	if (d.Username == "") != (d.Password == "") {
		return errors.New("worker username and password must be supplied together")
	}
	if len(d.Username) > 1024 || len(d.Password) > 4096 {
		return errors.New("worker credential length is invalid")
	}
	return nil
}

func EncodeWorkerDescriptor(w io.Writer, descriptor WorkerDescriptor) error {
	if w == nil {
		return errors.New("worker descriptor writer must not be nil")
	}
	if err := descriptor.Validate(); err != nil {
		return err
	}
	payload, err := json.Marshal(descriptor)
	if err != nil {
		return fmt.Errorf("encode worker descriptor: %w", err)
	}
	if len(payload) > workerDescriptorMaxBytes {
		return errors.New("worker descriptor exceeds size limit")
	}
	if _, err := w.Write(append(payload, '\n')); err != nil {
		return fmt.Errorf("write worker descriptor: %w", err)
	}
	return nil
}

func DecodeWorkerDescriptor(r io.Reader) (WorkerDescriptor, error) {
	if r == nil {
		return WorkerDescriptor{}, errors.New("worker descriptor reader must not be nil")
	}
	limited := io.LimitReader(r, workerDescriptorMaxBytes+1)
	payload, err := io.ReadAll(limited)
	if err != nil {
		return WorkerDescriptor{}, fmt.Errorf("read worker descriptor: %w", err)
	}
	if len(payload) > workerDescriptorMaxBytes {
		return WorkerDescriptor{}, errors.New("worker descriptor exceeds size limit")
	}
	dec := json.NewDecoder(strings.NewReader(string(payload)))
	dec.DisallowUnknownFields()
	var d WorkerDescriptor
	if err := dec.Decode(&d); err != nil {
		return WorkerDescriptor{}, fmt.Errorf("decode worker descriptor: %w", err)
	}
	var extra any
	if err := dec.Decode(&extra); !errors.Is(err, io.EOF) {
		return WorkerDescriptor{}, errors.New("worker descriptor must contain one JSON value")
	}
	if err := d.Validate(); err != nil {
		return WorkerDescriptor{}, err
	}
	return d, nil
}

func NewWorkerDescriptorPipe(descriptor WorkerDescriptor) (*os.File, func(), error) {
	if err := descriptor.Validate(); err != nil {
		return nil, nil, err
	}
	readEnd, writeEnd, err := os.Pipe()
	if err != nil {
		return nil, nil, fmt.Errorf("create worker descriptor pipe: %w", err)
	}
	done := make(chan struct{})
	go func() { defer close(done); defer writeEnd.Close(); _ = EncodeWorkerDescriptor(writeEnd, descriptor) }()
	cleanup := func() { _ = readEnd.Close(); <-done }
	return readEnd, cleanup, nil
}

func ProtectedWorkerArgs() []string {
	return []string{fmt.Sprintf("--descriptor-fd=%d", WorkerDescriptorFD)}
}

func SanitizedWorkerEnvironment(_ []string) []string {
	return []string{"LANG=C", "LC_ALL=C", "TZ=UTC"}
}
