package cli

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/ZXlinxk/nosub/internal/envguard"
	"github.com/ZXlinxk/nosub/internal/i18n"
)

func (a *app) env(args []string) (int, error) {
	if len(args) == 0 || args[0] == "--help" || args[0] == "-h" {
		fmt.Fprint(a.out, i18n.T("help.env"))
		return exitOK, nil
	}
	switch args[0] {
	case "scan":
		return a.envScan(args[1:])
	case "hook":
		return a.envHook(args[1:])
	case "example":
		return a.envExample(args[1:])
	case "gitignore":
		return a.envGitignore()
	}
	return exitError, errors.New(i18n.T("err.unknown", args[0]))
}

func (a *app) envScan(args []string) (int, error) {
	fs := flag.NewFlagSet("scan", flag.ContinueOnError)
	staged := fs.Bool("staged", false, "")
	history := fs.Bool("history", false, "")
	asJSON := fs.Bool("json", false, "")
	pos, err := a.parse(fs, "help.env", args)
	if err != nil {
		return done(err)
	}
	root := "."
	if len(pos) > 0 {
		root = pos[0]
	}

	// A mistyped path must fail, not report a clean project.
	if _, err := os.Stat(root); err != nil {
		return exitError, err
	}

	var found []envguard.Finding
	switch {
	case *staged:
		found, err = envguard.ScanStaged(root)
	case *history:
		found, err = envguard.ScanHistory(root)
	default:
		found, err = envguard.ScanDir(root)
	}
	if err != nil {
		return exitError, err
	}

	if *asJSON {
		if found == nil {
			found = []envguard.Finding{}
		}
		enc := json.NewEncoder(a.out)
		enc.SetIndent("", "  ")
		if err := enc.Encode(found); err != nil {
			return exitError, err
		}
	} else {
		a.printFindings(found)
	}
	if len(found) > 0 {
		return exitFindings, nil
	}
	return exitOK, nil
}

func (a *app) printFindings(found []envguard.Finding) {
	if len(found) == 0 {
		fmt.Fprintln(a.out, i18n.T("scan.clean"))
		return
	}
	for _, f := range found {
		loc := f.File
		if f.Line > 0 {
			loc = fmt.Sprintf("%s:%d", f.File, f.Line)
		}
		preview := f.Preview
		if f.Rule == "env-file" {
			preview = i18n.T("scan.envfile")
		}
		fmt.Fprintf(a.out, "%s  [%s]  %s", loc, f.Rule, preview)
		if f.Commit != "" {
			fmt.Fprintf(a.out, "  (%s)", i18n.T("scan.commit", f.Commit))
		}
		fmt.Fprintln(a.out)
	}
	fmt.Fprintln(a.err, i18n.T("scan.found", len(found)))
	fmt.Fprintln(a.err, i18n.T("scan.tip"))
}

func (a *app) envHook(args []string) (int, error) {
	fs := flag.NewFlagSet("hook", flag.ContinueOnError)
	force := fs.Bool("force", false, "")
	pos, err := a.parse(fs, "help.env", args)
	if err != nil {
		return done(err)
	}
	if len(pos) != 1 {
		return exitError, errors.New(i18n.T("err.args"))
	}
	var key string
	switch pos[0] {
	case "install":
		err, key = envguard.InstallHook(".", *force), "hook.installed"
	case "uninstall":
		err, key = envguard.UninstallHook("."), "hook.removed"
	default:
		return exitError, errors.New(i18n.T("err.unknown", pos[0]))
	}
	if errors.Is(err, envguard.ErrHookExists) {
		return exitError, errors.New(i18n.T("err.hook_exists"))
	}
	if err != nil {
		return exitError, err
	}
	fmt.Fprintln(a.out, i18n.T(key))
	return exitOK, nil
}

func (a *app) envExample(args []string) (int, error) {
	fs := flag.NewFlagSet("example", flag.ContinueOnError)
	in := fs.String("in", ".env", "")
	out := fs.String("out", ".env.example", "")
	if _, err := a.parse(fs, "help.env", args); err != nil {
		return done(err)
	}
	b, err := os.ReadFile(*in)
	if err != nil {
		return exitError, err
	}
	if err := os.WriteFile(*out, []byte(envguard.GenerateExample(string(b))), 0o644); err != nil {
		return exitError, err
	}
	fmt.Fprintln(a.out, i18n.T("example.written", *out))
	return exitOK, nil
}

func (a *app) envGitignore() (int, error) {
	changed, err := envguard.EnsureGitignore(".")
	if err != nil {
		return exitError, err
	}
	if changed {
		fmt.Fprintln(a.out, i18n.T("gitignore.updated"))
	} else {
		fmt.Fprintln(a.out, i18n.T("gitignore.ok"))
	}
	return exitOK, nil
}
