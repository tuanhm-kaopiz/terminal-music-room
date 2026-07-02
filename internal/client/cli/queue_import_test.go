package cli

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestQueueImportDryRun(t *testing.T) {
	resetCLIGlobals()
	defer resetCLIGlobals()

	dir := t.TempDir()
	csvPath := filepath.Join(dir, "playlist.csv")
	if err := os.WriteFile(csvPath, []byte(`url
https://www.youtube.com/watch?v=dQw4w9WgXcQ
https://youtu.be/abc123xyz01
`), 0o644); err != nil {
		t.Fatal(err)
	}

	RootCmd.SetArgs([]string{"queue", "import", csvPath, "--dry-run"})
	var out strings.Builder
	RootCmd.SetOut(&out)
	RootCmd.SetErr(io.Discard)
	if err := RootCmd.ExecuteContext(context.Background()); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "dry-run: 2 url(s)") {
		t.Fatalf("output: %s", out.String())
	}
}

func TestQueueImportRequiresRoom(t *testing.T) {
	_, ts, cfgPath := testHub(t)
	defer ts.Close()
	resetCLIGlobals()
	defer resetCLIGlobals()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := Login(ctx, io.Discard, cfgPath, "import-user", ts.URL); err != nil {
		t.Fatal(err)
	}
	configPath = cfgPath

	dir := t.TempDir()
	csvPath := filepath.Join(dir, "playlist.csv")
	if err := os.WriteFile(csvPath, []byte("url\nhttps://www.youtube.com/watch?v=dQw4w9WgXcQ\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	RootCmd.SetArgs([]string{"queue", "import", csvPath})
	RootCmd.SetOut(io.Discard)
	RootCmd.SetErr(io.Discard)
	err := RootCmd.ExecuteContext(ctx)
	if err == nil || !strings.Contains(err.Error(), "not in a room") {
		t.Fatalf("err = %v", err)
	}
}
