package cli

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/ZXlinxk/nosub/internal/i18n"
	"github.com/ZXlinxk/nosub/internal/schema"
)

func (a *app) schema(args []string) (int, error) {
	fs := flag.NewFlagSet("schema", flag.ContinueOnError)
	out := fs.String("o", "", "")
	compact := fs.Bool("compact", false, "")
	only := fs.String("only", "", "")
	exclude := fs.String("exclude", "", "")
	pos, err := a.parse(fs, "help.schema", args)
	if err != nil {
		return done(err)
	}
	if len(pos) == 0 {
		fmt.Fprint(a.out, i18n.T("help.schema"))
		return exitOK, nil
	}
	if len(pos) != 1 {
		return exitError, errors.New(i18n.T("err.args"))
	}

	var s *schema.Schema
	if pos[0] == "-" {
		s, err = schema.LoadReader(a.in)
	} else {
		s, err = schema.Load(pos[0])
	}
	if err != nil {
		return exitError, err
	}
	md := schema.Markdown(s, schema.MarkdownOptions{
		Compact: *compact,
		Only:    splitList(*only),
		Exclude: splitList(*exclude),
	})
	if *out == "" {
		fmt.Fprint(a.out, md)
		return exitOK, nil
	}
	if err := os.WriteFile(*out, []byte(md), 0o644); err != nil {
		return exitError, err
	}
	// Four characters per token is the usual rough estimate for English text.
	fmt.Fprintln(a.out, i18n.T("schema.written", *out, len(s.Tables), len(md)/4))
	return exitOK, nil
}
