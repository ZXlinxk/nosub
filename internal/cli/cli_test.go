package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func run(t *testing.T, stdin string, args ...string) (int, string, string) {
	t.Helper()
	var out, errb bytes.Buffer
	code := Run(args, strings.NewReader(stdin), &out, &errb)
	return code, out.String(), errb.String()
}

func TestHelpIsTranslated(t *testing.T) {
	_, en, _ := run(t, "", "--lang", "en")
	_, pt, _ := run(t, "", "--lang=pt")
	if en == pt || !strings.Contains(en, "nosub schema") || !strings.Contains(pt, "nosub schema") {
		t.Fatalf("en=%q pt=%q", en, pt)
	}
}

func TestUnknownCommand(t *testing.T) {
	if code, _, _ := run(t, "", "--lang=en", "nope"); code != exitError {
		t.Fatalf("code %d", code)
	}
}

func TestScanExitCodes(t *testing.T) {
	dir := t.TempDir()
	if code, out, _ := run(t, "", "--lang=en", "env", "scan", dir); code != exitOK || !strings.Contains(out, "No secrets") {
		t.Fatalf("%d %s", code, out)
	}
	key := "AKIA" + "Q7ZKD3LPX9MNB2WT"
	os.WriteFile(filepath.Join(dir, "a.txt"), []byte("k="+key), 0o644)
	code, out, _ := run(t, "", "--lang=en", "env", "scan", dir)
	if code != exitFindings || !strings.Contains(out, "a.txt:1") || strings.Contains(out, key) {
		t.Fatalf("%d %s", code, out)
	}
}

func TestScanMissingPathIsAnError(t *testing.T) {
	if code, _, _ := run(t, "", "env", "scan", filepath.Join(t.TempDir(), "nope")); code != exitError {
		t.Fatalf("code %d", code)
	}
}

func TestScanJSONAcceptsFlagAfterPath(t *testing.T) {
	code, out, _ := run(t, "", "env", "scan", t.TempDir(), "--json")
	var v []map[string]any
	if code != exitOK || json.Unmarshal([]byte(out), &v) != nil {
		t.Fatalf("%d %q", code, out)
	}
}

func TestSchemaFromStdin(t *testing.T) {
	code, out, _ := run(t, "CREATE TABLE a (id int PRIMARY KEY);", "schema", "-", "--compact")
	if code != exitOK || !strings.Contains(out, "a(id PK)") {
		t.Fatalf("%d %q", code, out)
	}
}

func TestSvgFolder(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "arrow-left.svg"), []byte(`<svg width="24" height="24"/>`), 0o644)
	out := filepath.Join(dir, "components")
	code, _, stderr := run(t, "", "svg", dir, "-o", out, "--ts", "--tailwind")
	if code != exitOK {
		t.Fatalf("%d %s", code, stderr)
	}
	b, err := os.ReadFile(filepath.Join(out, "ArrowLeft.tsx"))
	if err != nil || !strings.Contains(string(b), "size-6") {
		t.Fatalf("%s %v", b, err)
	}
	idx, _ := os.ReadFile(filepath.Join(out, "index.ts"))
	if !strings.Contains(string(idx), `export { default as ArrowLeft } from "./ArrowLeft";`) {
		t.Fatalf("%s", idx)
	}
}
