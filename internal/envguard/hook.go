package envguard

import (
	"os"
	"path/filepath"
	"strings"
)

const (
	hookSignature = "# managed by nosub"
	hookScript    = "#!/bin/sh\n# managed by nosub\nexec nosub env scan --staged\n"
)

// InstallHook installs the pre-commit hook in a git repository.
func InstallHook(root string, force bool) error {
	p, err := resolveHookPath(root)
	if err != nil {
		return err
	}

	b, err := os.ReadFile(p)
	if err == nil {
		isOurs := strings.Contains(string(b), hookSignature)
		if !isOurs && !force {
			return ErrHookExists
		}
	} else if !os.IsNotExist(err) {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	return os.WriteFile(p, []byte(hookScript), 0o755)
}

// UninstallHook removes the pre-commit hook if managed by nosub.
func UninstallHook(root string) error {
	p, err := resolveHookPath(root)
	if err != nil {
		return err
	}

	b, err := os.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	if !strings.Contains(string(b), hookSignature) {
		return ErrHookExists
	}

	return os.Remove(p)
}

func resolveHookPath(root string) (string, error) {
	out, err := gitCmd(root, "rev-parse", "--git-path", "hooks/pre-commit")
	if err != nil {
		return "", err
	}
	p := strings.TrimSpace(out)
	if !filepath.IsAbs(p) {
		p = filepath.Join(root, p)
	}
	return p, nil
}
