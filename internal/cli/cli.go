// Package cli wires the command line to the tool packages.
package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"strings"

	"github.com/ZXlinxk/nosub/internal/i18n"
)

// Version is set at build time.
var Version = "dev"

const (
	exitOK       = 0
	exitFindings = 1
	exitError    = 2
)

var errHelp = errors.New("help")

type app struct {
	in       io.Reader
	out, err io.Writer
}

// Run executes the command line and returns the process exit code.
func Run(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	lang, args := extractLang(args)
	i18n.Init(lang)
	a := &app{stdin, stdout, stderr}

	if len(args) == 0 {
		a.usage()
		return exitOK
	}
	cmd, rest := args[0], args[1:]
	var code int
	var err error
	switch cmd {
	case "env":
		code, err = a.env(rest)
	case "schema":
		code, err = a.schema(rest)
	case "svg":
		code, err = a.svg(rest)
	case "version", "--version", "-v":
		fmt.Fprintln(stdout, Version)
	case "help", "--help", "-h":
		a.usage()
	default:
		fmt.Fprintln(stderr, i18n.T("err.unknown", cmd))
		return exitError
	}
	if err != nil {
		fmt.Fprintln(stderr, i18n.T("err.generic", err))
		return exitError
	}
	return code
}

func (a *app) usage() {
	fmt.Fprint(a.out, i18n.T("help.main", Version, strings.Join(i18n.Languages(), " ")))
}

func extractLang(args []string) (string, []string) {
	var lang string
	var rest []string
	for i := 0; i < len(args); i++ {
		switch v, ok := strings.CutPrefix(args[i], "--lang="); {
		case ok:
			lang = v
		case args[i] == "--lang" && i+1 < len(args):
			i++
			lang = args[i]
		default:
			rest = append(rest, args[i])
		}
	}
	return lang, rest
}

// parse lets flags appear after positional arguments, which flag.Parse alone refuses.
func (a *app) parse(fs *flag.FlagSet, helpKey string, args []string) ([]string, error) {
	fs.SetOutput(io.Discard)
	fs.Usage = func() {}
	var pos []string
	for {
		err := fs.Parse(args)
		if errors.Is(err, flag.ErrHelp) {
			fmt.Fprint(a.out, i18n.T(helpKey))
			return nil, errHelp
		}
		if err != nil {
			return nil, err
		}
		if fs.NArg() == 0 {
			return pos, nil
		}
		pos = append(pos, fs.Arg(0))
		args = fs.Args()[1:]
	}
}

// done turns the help sentinel into a clean exit.
func done(err error) (int, error) {
	if errors.Is(err, errHelp) {
		return exitOK, nil
	}
	return exitError, err
}

func splitList(s string) []string {
	if s == "" {
		return nil
	}
	return strings.Split(s, ",")
}
