package media

import (
	"os"
	"strings"
	"testing"

	"github.com/GoreeCloud/goreecloud-home-security/internal/config"
)

func TestResolveWorkerDescriptorKeepsCredentialsOutOfURL(t *testing.T) {
	camera := config.Camera{ID: "front-door", Name: "Front Door", StreamURL: "rtsps://camera.local/live", UsernameEnv: "CAMERA_USER", PasswordEnv: "CAMERA_PASS", Enabled: true}
	values := map[string]string{"CAMERA_USER": "alice", "CAMERA_PASS": "top-secret"}
	d, err := ResolveWorkerDescriptor(camera, func(name string) (string, bool) { v, ok := values[name]; return v, ok })
	if err != nil {
		t.Fatal(err)
	}
	if d.Username != "alice" || d.Password != "top-secret" {
		t.Fatalf("descriptor = %#v", d)
	}
	if strings.Contains(d.StreamURL, "alice") || strings.Contains(d.StreamURL, "top-secret") {
		t.Fatalf("credential leaked into URL: %q", d.StreamURL)
	}
}

func TestWorkerDescriptorAnonymousPipeRoundTrip(t *testing.T) {
	d := WorkerDescriptor{Version: WorkerDescriptorVersion, CameraID: "garage", StreamURL: "rtsp://camera.local/live", Username: "user", Password: "secret"}
	reader, cleanup, err := NewWorkerDescriptorPipe(d)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	got, err := DecodeWorkerDescriptor(reader)
	if err != nil {
		t.Fatal(err)
	}
	if got != d {
		t.Fatalf("descriptor = %#v", got)
	}
}

func TestProtectedWorkerSurfaceHasNoCredentials(t *testing.T) {
	args := strings.Join(ProtectedWorkerArgs(), " ")
	env := strings.Join(SanitizedWorkerEnvironment([]string{"PATH=/usr/bin", "LANG=C", "CAMERA_USER=alice", "CAMERA_PASS=top-secret", "OTHER=private"}), " ")
	for _, secret := range []string{"alice", "top-secret", "CAMERA_USER", "CAMERA_PASS", "OTHER"} {
		if strings.Contains(args, secret) || strings.Contains(env, secret) {
			t.Fatalf("protected worker surface leaked %q: args=%q env=%q", secret, args, env)
		}
	}
	if !strings.Contains(args, "descriptor-fd=3") || !strings.Contains(env, "LANG=C") || strings.Contains(env, "PATH=") {
		t.Fatalf("surface = args %q env %q", args, env)
	}
}

func TestResolveWorkerDescriptorFailsClosedWhenSecretMissing(t *testing.T) {
	camera := config.Camera{ID: "front-door", Name: "Front Door", StreamURL: "rtsp://camera.local/live", UsernameEnv: "CAMERA_USER", PasswordEnv: "CAMERA_PASS", Enabled: true}
	_, err := ResolveWorkerDescriptor(camera, os.LookupEnv)
	if ErrorCode(err) != ReasonCredentialUnavailable {
		t.Fatalf("error = %v", err)
	}
}
