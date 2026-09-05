package media

import (
	"strings"
	"testing"
	"time"

	"github.com/GoreeCloud/goreecloud-home-security/internal/config"
)

func TestResolveWorkerDescriptorV2CarriesTimeoutAndCredentials(t *testing.T) {
	camera := config.Camera{ID: "front-door", Name: "Front Door", StreamURL: "rtsp://camera.local/live", UsernameEnv: "CAMERA_USER", PasswordEnv: "CAMERA_PASS", Enabled: true}
	values := map[string]string{"CAMERA_USER": "alice", "CAMERA_PASS": "secret"}
	d, err := ResolveWorkerDescriptor(camera, func(name string) (string, bool) { v, ok := values[name]; return v, ok }, 17*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if d.Version != WorkerDescriptorVersion || d.RWTimeoutSeconds != 17 || d.Username != "alice" || d.Password != "secret" {
		t.Fatalf("descriptor = %#v", d)
	}
	if strings.Contains(d.StreamURL, "alice") || strings.Contains(d.StreamURL, "secret") {
		t.Fatalf("credential leaked into url %q", d.StreamURL)
	}
}

func TestProtectedWorkerArgsExposeOnlyDescriptors(t *testing.T) {
	joined := strings.Join(ProtectedWorkerArgs(), " ")
	if joined != "--descriptor-fd=3 --status-fd=4" {
		t.Fatalf("args = %q", joined)
	}
}

func TestDescriptorRejectsControlCharactersInCredentials(t *testing.T) {
	d := WorkerDescriptor{Version: WorkerDescriptorVersion, CameraID: "front-door", StreamURL: "rtsp://camera.local/live", Username: "alice\nInjected", Password: "secret", RWTimeoutSeconds: 15}
	if err := d.Validate(); err == nil {
		t.Fatal("expected validation failure")
	}
}
