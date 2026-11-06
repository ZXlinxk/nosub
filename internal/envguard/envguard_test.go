package envguard

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Fake credentials are split so repository secret scanners do not trip on this file.
var (
	awsKey    = "AKIA" + "Q7ZKD3LPX9MNB2WT"
	ghToken   = "ghp_" + "a1B2c3D4e5F6g7H8i9J0k1L2m3N4o5P6q7R8"
	stripeKey = "sk_live_" + "51Hx9aQ2kLmN7pRs4TuVw8Yz"
)

func rules(fs []Finding) string {
	var out []string
	for _, f := range fs {
		out = append(out, f.Rule)
	}
	return strings.Join(out, ",")
}

func TestScanTextDetectsProviders(t *testing.T) {
	cases := []struct{ name, line, rule string }{
		{"aws", `key = "` + awsKey + `"`, "aws-access-key"},
		{"github", `TOKEN=` + ghToken, "github-token"},
		{"stripe", `stripe.api_key = '` + stripeKey + `'`, "stripe-key"},
		{"pem", "-----BEGIN RSA " + "PRIVATE KEY-----", "private-key"},
		{"dburl", `DATABASE_URL=postgres://app:s3cr3tPw9x@db.internal:5432/main`, "url-password"},
		{"generic", `API_SECRET=Zk82hQmP0xLw7vTr`, "generic-secret"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := ScanText("config.txt", c.line)
			if len(got) == 0 || got[0].Rule != c.rule {
				t.Fatalf("want %s, got %q", c.rule, rules(got))
			}
			if got[0].Line != 1 || got[0].File != "config.txt" {
				t.Fatalf("bad location: %+v", got[0])
			}
		})
	}
}

func TestPreviewNeverContainsSecret(t *testing.T) {
	got := ScanText("a.env", "TOKEN="+ghToken)
	if len(got) != 1 {
		t.Fatalf("want one finding, got %d", len(got))
	}
	if strings.Contains(got[0].Preview, ghToken) || strings.Contains(got[0].Preview, ghToken[8:20]) {
		t.Fatalf("preview leaks the secret: %q", got[0].Preview)
	}
}

func TestPreviewMasksEverySecretOnTheLine(t *testing.T) {
	got := ScanText("a", awsKey+" "+ghToken)
	if len(got) != 2 {
		t.Fatalf("want two findings, got %d", len(got))
	}
	for _, f := range got {
		if strings.Contains(f.Preview, awsKey) || strings.Contains(f.Preview, ghToken) {
			t.Fatalf("leak: %q", f.Preview)
		}
	}
}

func TestTypedDeclaration(t *testing.T) {
	got := ScanText("A.java", `private String password = "Zk82hQmP0xLw7vTr";`)
	if len(got) != 1 || got[0].Rule != "generic-secret" {
		t.Fatalf("got %q", rules(got))
	}
}

func TestNoFalsePositives(t *testing.T) {
	lines := []string{
		`API_KEY=`,
		`API_KEY=changeme`,
		`API_KEY=your_api_key_here`,
		`PASSWORD=<password>`,
		`SECRET_KEY=${SECRET_KEY}`,
		`token = os.getenv("API_TOKEN")`,
		`const password = process.env.DB_PASSWORD;`,
		`DATABASE_URL=postgres://user:password@localhost:5432/db`,
		`PASSWORD=xxxxxxxxxxxxxxxx`,
		`AWS_ACCESS_KEY_ID=AKIAIOSFODNN7EXAMPLE`,
		`tokenizer_model = "gpt-tokenizer-small"`,
		`password = getPassword(userRecord)`,
		`token = request.headers.authorization`,
		`if token == nil {`,
		`"secret": "シークレットキー値です",`,
		`postgres://user:pw@host/db`,
		`DATABASE_URL=postgres://{user}:{password}@localhost/db`,
	}
	for _, l := range lines {
		if got := ScanText(".env.example", l); len(got) != 0 {
			t.Errorf("false positive on %q: %s", l, rules(got))
		}
	}
}

func TestInlineAllow(t *testing.T) {
	if got := ScanText("a.py", `k = "`+awsKey+`"  # nosub:allow`); len(got) != 0 {
		t.Fatalf("inline allow ignored: %s", rules(got))
	}
}

func TestLineNumbers(t *testing.T) {
	got := ScanText("a", "one\ntwo\nx="+awsKey+"\n")
	if len(got) != 1 || got[0].Line != 3 {
		t.Fatalf("bad: %+v", got)
	}
}

func TestIsEnvFile(t *testing.T) {
	yes := []string{".env", ".env.local", ".env.production", "config/.env", "app.env"}
	no := []string{".env.example", ".env.sample", ".env.template", ".env.dist", "environment.ts", "env.go"}
	for _, n := range yes {
		if !IsEnvFile(n) {
			t.Errorf("%s should be an env file", n)
		}
	}
	for _, n := range no {
		if IsEnvFile(n) {
			t.Errorf("%s should not be an env file", n)
		}
	}
}

func TestGenerateExample(t *testing.T) {
	in := "# Stripe\nexport STRIPE_KEY=abc123\nPORT=3000\nDEBUG=true\n\nDB_PASSWORD=\"hunter2\"\nAPI_URL=https://internal.example.net\n"
	want := "# Stripe\nexport STRIPE_KEY=\nPORT=3000\nDEBUG=true\n\nDB_PASSWORD=\nAPI_URL=\n"
	if got := GenerateExample(in); got != want {
		t.Fatalf("got:\n%s\nwant:\n%s", got, want)
	}
}

func TestEnsureGitignore(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, ".gitignore"), []byte("node_modules\n"), 0o644)
	changed, err := EnsureGitignore(dir)
	if err != nil || !changed {
		t.Fatalf("changed=%v err=%v", changed, err)
	}
	b, _ := os.ReadFile(filepath.Join(dir, ".gitignore"))
	for _, want := range []string{"node_modules", ".env\n", ".env.*\n", "!.env.example\n"} {
		if !strings.Contains(string(b), want) {
			t.Fatalf("missing %q in %q", want, b)
		}
	}
	changed, _ = EnsureGitignore(dir)
	if changed {
		t.Fatal("second run must not change the file")
	}
}

func git(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-c", "user.name=t", "-c", "user.email=t@t", "-c", "core.autocrlf=false"}, args...)...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

func repo(t *testing.T) string {
	dir := t.TempDir()
	git(t, dir, "init", "-q")
	return dir
}

func write(t *testing.T, dir, name, body string) {
	t.Helper()
	p := filepath.Join(dir, name)
	os.MkdirAll(filepath.Dir(p), 0o755)
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestScanDirSkipsIgnoredAndVendored(t *testing.T) {
	dir := repo(t)
	write(t, dir, ".gitignore", "secret.txt\n")
	write(t, dir, "secret.txt", "x="+awsKey)
	write(t, dir, "node_modules/lib/a.js", "x="+awsKey)
	write(t, dir, "src/app.py", "x="+awsKey)
	write(t, dir, "logo.bin", "\x00\x01"+awsKey)
	got, err := ScanDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || filepath.ToSlash(got[0].File) != "src/app.py" {
		t.Fatalf("got %+v", got)
	}
}

func TestScanDirWithoutGit(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "node_modules/a.js", "x="+awsKey)
	write(t, dir, "b.js", "x="+awsKey)
	got, err := ScanDir(dir)
	if err != nil || len(got) != 1 {
		t.Fatalf("got %+v err %v", got, err)
	}
}

func TestNosubignore(t *testing.T) {
	dir := repo(t)
	write(t, dir, ".nosubignore", "# fixtures\ntestdata/*\n")
	write(t, dir, "testdata/keys.txt", "x="+awsKey)
	got, err := ScanDir(dir)
	if err != nil || len(got) != 0 {
		t.Fatalf("got %+v err %v", got, err)
	}
}

func TestScanStaged(t *testing.T) {
	dir := repo(t)
	write(t, dir, "ok.txt", "hello\n")
	git(t, dir, "add", ".")
	git(t, dir, "commit", "-qm", "init")
	write(t, dir, "ok.txt", "unstaged="+awsKey+"\n")
	write(t, dir, ".env", "A=1\n")
	write(t, dir, "new.py", "t='"+ghToken+"'\n")
	git(t, dir, "add", ".env", "new.py")
	got, err := ScanStaged(dir)
	if err != nil {
		t.Fatal(err)
	}
	r := rules(got)
	if !strings.Contains(r, "env-file") || !strings.Contains(r, "github-token") || strings.Contains(r, "aws-access-key") {
		t.Fatalf("got %s", r)
	}
}

func TestScanHistory(t *testing.T) {
	dir := repo(t)
	write(t, dir, "a.txt", "x="+awsKey+"\n")
	git(t, dir, "add", ".")
	git(t, dir, "commit", "-qm", "oops")
	write(t, dir, "a.txt", "clean\n")
	git(t, dir, "add", ".")
	git(t, dir, "commit", "-qm", "fix")
	got, err := ScanHistory(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Rule != "aws-access-key" || got[0].Commit == "" || got[0].File != "a.txt" {
		t.Fatalf("got %+v", got)
	}
}

func TestHookInstall(t *testing.T) {
	dir := repo(t)
	if err := InstallHook(dir, false); err != nil {
		t.Fatal(err)
	}
	p := filepath.Join(dir, ".git", "hooks", "pre-commit")
	b, err := os.ReadFile(p)
	if err != nil || !strings.Contains(string(b), "nosub env scan --staged") {
		t.Fatalf("%s %v", b, err)
	}
	if err := InstallHook(dir, false); err != nil {
		t.Fatalf("reinstall of our own hook must succeed: %v", err)
	}
	os.WriteFile(p, []byte("#!/bin/sh\necho mine\n"), 0o755)
	if err := InstallHook(dir, false); err != ErrHookExists {
		t.Fatalf("want ErrHookExists, got %v", err)
	}
	if err := UninstallHook(dir); err != ErrHookExists {
		t.Fatalf("must not remove a foreign hook: %v", err)
	}
	if err := InstallHook(dir, true); err != nil {
		t.Fatal(err)
	}
	if err := UninstallHook(dir); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(p); !os.IsNotExist(err) {
		t.Fatal("hook still there")
	}
}
