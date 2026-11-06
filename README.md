# nosub

[English](README.md) · [Português](docs/README.pt.md) · [Español](docs/README.es.md) · [Français](docs/README.fr.md) · [Deutsch](docs/README.de.md) · [Italiano](docs/README.it.md) · [Polski](docs/README.pl.md) · [Русский](docs/README.ru.md) · [Türkçe](docs/README.tr.md) · [Bahasa Indonesia](docs/README.id.md) · [हिन्दी](docs/README.hi.md) · [中文](docs/README.zh.md) · [日本語](docs/README.ja.md) · [한국어](docs/README.ko.md)

Three small dev tools in one 4 MB binary. No account, no telemetry, no monthly plan. It runs on your machine and never opens a network connection of its own.

- `nosub env` keeps secrets out of git
- `nosub schema` exports a database schema to Markdown you can paste into an LLM
- `nosub svg` turns SVG files into clean JSX or Tailwind components

## Install

Linux and macOS:

```sh
curl -fsSL https://raw.githubusercontent.com/ZXlinxk/nosub/main/install.sh | sh
```

Windows (PowerShell):

```powershell
irm https://raw.githubusercontent.com/ZXlinxk/nosub/main/install.ps1 | iex
```

With Go:

```sh
go install github.com/ZXlinxk/nosub/cmd/nosub@latest
```

Or grab a binary from the [releases page](https://github.com/ZXlinxk/nosub/releases). There is nothing else to install.

## nosub env

Scan a project, a commit that is about to happen, or the whole git history:

```sh
nosub env scan
nosub env scan --staged
nosub env scan --history
```

```
k.txt:1  [github-token]  ghp_****
1 possible secret(s) found.
```

Secrets are masked in the output, so the report is safe to paste in a CI log. The exit code is 0 when clean, 1 when something was found and 2 on errors. `--json` gives you machine readable output.

Block bad commits for good:

```sh
nosub env hook install
```

The hook also refuses a `.env` file that was staged by accident. Two more helpers:

```sh
nosub env example      # writes .env.example from .env, without the values
nosub env gitignore    # adds .env rules to .gitignore
```

It knows AWS, GitHub, GitLab, Slack, Stripe, OpenAI, Anthropic, Google, SendGrid, npm and PyPI credentials, private keys, JWTs, passwords inside URLs, and high entropy values assigned to names like `API_SECRET`.

False positive? Put `nosub:allow` on the line, or list the path in `.nosubignore`.

## nosub schema

```sh
nosub schema app.db                         # SQLite file, no driver needed
nosub schema dump.sql                       # pg_dump, mysqldump, phpMyAdmin or .schema output
pg_dump -s mydb | nosub schema -
nosub schema postgres://user:pw@host/mydb   # calls pg_dump
nosub schema mysql://user:pw@host/mydb      # calls mysqldump
```

Output is short on purpose. Tables of pipes and dashes eat tokens and tell the model nothing:

```
## orders
- id bigint PK
- user_id bigint -> users.id ON DELETE CASCADE
- total numeric(10,2) NOT NULL DEFAULT 0
Indexes: idx_orders_user (user_id)
```

For big databases, `--compact` keeps one line per table:

```
users(id PK, email, created_at)
orders(id PK, user_id -> users.id, total)
```

Use `--only` and `--exclude` with globs to pick tables, and `-o schema.md` to write a file. Passwords in connection strings go to the dump tools through environment variables, not arguments.

## nosub svg

```sh
nosub svg arrow-right.svg --ts --tailwind
```

```tsx
import type { SVGProps } from "react";

export default function ArrowRight(props: SVGProps<SVGSVGElement>) {
  return (
    <svg className="size-6 fill-none" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" {...props}>
      <path className="stroke-blue-500 stroke-2" d="M5 12h14" strokeLinecap="round" />
    </svg>
  );
}
```

Attributes become their React names, `style` strings become objects, and everything Inkscape, Sketch, Figma or Illustrator left behind is removed. Hex colors that match the Tailwind palette become palette classes.

Convert a whole folder, with an index file:

```sh
nosub svg ./icons -o ./src/icons --ts
```

`--current-color` swaps colors for `currentColor`, which is what you want for icons that follow the text color.

## Limits

- `nosub env` matches patterns. It catches the usual mistakes and it will never catch everything, so it does not replace rotating a leaked key.
- Live database URLs need `pg_dump` or `mysqldump` in your `PATH`. Dump files and SQLite need nothing.
- A SQLite database in WAL mode may have recent schema changes that are still in the `-wal` file. Close the app using it first.
- `size-*` classes need Tailwind 3.4 or newer.

## Languages

Messages follow your system language. Override with `--lang pt` or `NOSUB_LANG=pt`. Available: en, pt, es, fr, de, it, pl, ru, tr, id, hi, zh, ja, ko.

## Support the project

nosub is free and MIT licensed, and it will stay that way. If it saved you a subscription, [you can chip in here](https://buymeacoffee.com/samaritansli).

- Lightning: `inborngrease4665@cake.cash`
- Bitcoin: `bc1qm2j0l9zy7npnhfctf80p3ukuj2ksffwlvum3c5`

## License

[MIT](LICENSE)
