package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/ZXlinxk/nosub/internal/i18n"
	"github.com/ZXlinxk/nosub/internal/svgjsx"
)

func (a *app) svg(args []string) (int, error) {
	fs := flag.NewFlagSet("svg", flag.ContinueOnError)
	out := fs.String("o", "", "")
	name := fs.String("name", "", "")
	ts := fs.Bool("ts", false, "")
	tailwind := fs.Bool("tailwind", false, "")
	current := fs.Bool("current-color", false, "")
	pos, err := a.parse(fs, "help.svg", args)
	if err != nil {
		return done(err)
	}
	if len(pos) == 0 {
		fmt.Fprint(a.out, i18n.T("help.svg"))
		return exitOK, nil
	}
	if len(pos) != 1 {
		return exitError, errors.New(i18n.T("err.args"))
	}
	opts := svgjsx.Options{Name: *name, TypeScript: *ts, Tailwind: *tailwind, CurrentColor: *current}

	src := pos[0]
	if src == "-" {
		b, err := io.ReadAll(a.in)
		if err != nil {
			return exitError, err
		}
		if opts.Name == "" {
			opts.Name = "Icon"
		}
		return a.writeComponent(b, opts, *out)
	}
	if info, err := os.Stat(src); err != nil {
		return exitError, err
	} else if info.IsDir() {
		return a.svgDir(src, *out, opts)
	}
	b, err := os.ReadFile(src)
	if err != nil {
		return exitError, err
	}
	if opts.Name == "" {
		opts.Name = svgjsx.ComponentName(src)
	}
	return a.writeComponent(b, opts, *out)
}

func (a *app) writeComponent(svg []byte, opts svgjsx.Options, out string) (int, error) {
	code, err := svgjsx.Convert(svg, opts)
	if err != nil {
		return exitError, err
	}
	if out == "" {
		fmt.Fprint(a.out, code)
		return exitOK, nil
	}
	if err := os.WriteFile(out, []byte(code), 0o644); err != nil {
		return exitError, err
	}
	fmt.Fprintln(a.out, i18n.T("svg.written", 1, out))
	return exitOK, nil
}

// svgDir converts every .svg in dir and writes a barrel file next to the components.
func (a *app) svgDir(dir, out string, opts svgjsx.Options) (int, error) {
	files, err := filepath.Glob(filepath.Join(dir, "*.svg"))
	if err != nil {
		return exitError, err
	}
	if len(files) == 0 {
		return exitError, errors.New(i18n.T("err.no_svg", dir))
	}
	if out == "" {
		out = dir
	}
	if err := os.MkdirAll(out, 0o755); err != nil {
		return exitError, err
	}
	ext, indexName := ".jsx", "index.js"
	if opts.TypeScript {
		ext, indexName = ".tsx", "index.ts"
	}
	var index strings.Builder
	for _, f := range files {
		b, err := os.ReadFile(f)
		if err != nil {
			return exitError, err
		}
		opts.Name = svgjsx.ComponentName(f)
		code, err := svgjsx.Convert(b, opts)
		if err != nil {
			return exitError, fmt.Errorf("%s: %w", f, err)
		}
		if err := os.WriteFile(filepath.Join(out, opts.Name+ext), []byte(code), 0o644); err != nil {
			return exitError, err
		}
		fmt.Fprintf(&index, "export { default as %s } from \"./%s\";\n", opts.Name, opts.Name)
	}
	if err := os.WriteFile(filepath.Join(out, indexName), []byte(index.String()), 0o644); err != nil {
		return exitError, err
	}
	fmt.Fprintln(a.out, i18n.T("svg.written", len(files), out))
	return exitOK, nil
}
