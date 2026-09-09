package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/zhuangzhuangzhao921/kdl-agent-cli-test/internal/config"
)

func TestSaveLoadRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, config.DefaultFilename)
	if err := config.Save(path, config.File{
		GatewayURL: "http://127.0.0.1:8080",
		Token:      "kdl_ag_test.selector",
	}); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("want 0600, got %o", info.Mode().Perm())
	}

	t.Setenv("KDL_AGENT_CONFIG", path)
	t.Setenv("KDL_AGENT_GATEWAY_URL", "")
	t.Setenv("KDL_AGENT_TOKEN", "")

	got, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	if got.GatewayURL != "http://127.0.0.1:8080" {
		t.Fatalf("gateway url: %q", got.GatewayURL)
	}
	if got.Token != "kdl_ag_test.selector" {
		t.Fatalf("token mismatch")
	}
}

func TestValidateForAPIMissing(t *testing.T) {
	err := config.ValidateForAPI(config.Resolved{ConfigPath: "/tmp/kdl-agent.toml"})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestEnvOverridesFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, config.DefaultFilename)
	if err := config.Save(path, config.File{
		GatewayURL: "http://old",
		Token:      "old",
	}); err != nil {
		t.Fatal(err)
	}
	t.Setenv("KDL_AGENT_CONFIG", path)
	t.Setenv("KDL_AGENT_GATEWAY_URL", "http://new")
	t.Setenv("KDL_AGENT_TOKEN", "new-token")

	got, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	if got.GatewayURL != "http://new" || got.Token != "new-token" {
		t.Fatalf("env override failed: %+v", got)
	}
}

func TestPathOverrideBeatsEnv(t *testing.T) {
	dir := t.TempDir()
	flagPath := filepath.Join(dir, "from-flag.toml")
	envPath := filepath.Join(dir, "from-env.toml")
	for _, p := range []string{flagPath, envPath} {
		if err := config.Save(p, config.File{
			GatewayURL: "http://" + filepath.Base(p),
			Token:      "tok",
		}); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("KDL_AGENT_CONFIG", envPath)
	t.Setenv("KDL_AGENT_GATEWAY_URL", "")
	t.Setenv("KDL_AGENT_TOKEN", "")

	config.SetPathOverride(flagPath)
	t.Cleanup(func() { config.SetPathOverride("") })

	got, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	if got.ConfigPath != flagPath {
		want, err := filepath.Abs(flagPath)
		if err != nil {
			t.Fatal(err)
		}
		if got.ConfigPath != want {
			t.Fatalf("config path: got %q want %q", got.ConfigPath, want)
		}
	}
	if got.GatewayURL != "http://from-flag.toml" {
		t.Fatalf("flag path file should win: %+v", got)
	}
}

func TestResolvePathDefaultCWD(t *testing.T) {
	dir := t.TempDir()
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		config.SetPathOverride("")
		_ = os.Chdir(cwd)
	})
	t.Setenv("KDL_AGENT_CONFIG", "")
	config.SetPathOverride("")

	got, err := config.ResolvePath()
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(dir, config.DefaultFilename)
	if !sameResolvedPath(got, want) {
		t.Fatalf("ResolvePath: got %q want %q", got, want)
	}
}

func sameResolvedPath(a, b string) bool {
	return normalizeExistingParent(a) == normalizeExistingParent(b)
}

func normalizeExistingParent(p string) string {
	abs, err := filepath.Abs(p)
	if err != nil {
		return p
	}
	dir := filepath.Dir(abs)
	base := filepath.Base(abs)
	if evalDir, err := filepath.EvalSymlinks(dir); err == nil {
		return filepath.Join(evalDir, base)
	}
	return abs
}
