# nosub

[English](../README.md) · [Português](README.pt.md) · [Español](README.es.md) · [Français](README.fr.md) · [Deutsch](README.de.md) · [Italiano](README.it.md) · [Polski](README.pl.md) · [Русский](README.ru.md) · [Türkçe](README.tr.md) · [Bahasa Indonesia](README.id.md) · [हिन्दी](README.hi.md) · [中文](README.zh.md) · [日本語](README.ja.md) · [한국어](README.ko.md)

Drei kleine lokale Dev-Tools in einem einzigen 4-MB-Binary. Ohne Account, ohne Telemetrie, ohne monatliches Abo. Läuft direkt auf deinem Rechner und baut niemals eigenständig Netzwerkverbindungen auf.

- `nosub env` verhindert, dass Secrets in git landen
- `nosub schema` exportiert Datenbankschemata als Markdown, optimiert für LLM-Prompts
- `nosub svg` wandelt SVG-Dateien in saubere JSX- oder Tailwind-Komponenten um

## Installation

Linux und macOS:

```sh
curl -fsSL https://raw.githubusercontent.com/ZXlinxk/nosub/main/install.sh | sh
```

Windows (PowerShell):

```powershell
irm https://raw.githubusercontent.com/ZXlinxk/nosub/main/install.ps1 | iex
```

Mit Go:

```sh
go install github.com/ZXlinxk/nosub/cmd/nosub@latest
```

Oder lade das fertige Binary direkt von der [Releases-Seite](https://github.com/ZXlinxk/nosub/releases) herunter. Es muss nichts weiter installiert werden.

## nosub env

Scanne ein Projekt, den bevorstehenden Commit oder die gesamte Git-Historie:

```sh
nosub env scan
nosub env scan --staged
nosub env scan --history
```

```
k.txt:1  [github-token]  ghp_****
1 possible secret(s) found.
```

Secrets werden in der Ausgabe automatisch maskiert, sodass CI-Logs gefahrlos geteilt werden können. Der Exit-Code ist 0 bei sauberem Code, 1 wenn Secrets gefunden wurden und 2 bei Fehlern. `--json` liefert maschinenlesbare JSON-Daten.

Verhindere riskante Commits dauerhaft:

```sh
nosub env hook install
```

Der Hook blockiert auch versehentlich im Git-Staging gelandete `.env`-Dateien. Zwei weitere nützliche Helfer:

```sh
nosub env example      # erzeugt .env.example aus .env ohne Werte
nosub env gitignore    # fügt .env-Regeln zu .gitignore hinzu
```

Erkennt Credentials von AWS, GitHub, GitLab, Slack, Stripe, OpenAI, Anthropic, Google, SendGrid, npm und PyPI, private Schlüssel, JWTs, Passwörter in Verbindungs-URLs sowie Zeichenketten mit hoher Entropie hinter Variablennamen wie `API_SECRET`.

False Positive? Füge `nosub:allow` in der Zeile ein oder trage den Pfad in `.nosubignore` ein.

## nosub schema

```sh
nosub schema app.db                         # SQLite-Datei, kein Treiber erforderlich
nosub schema dump.sql                       # pg_dump, mysqldump, phpMyAdmin oder .schema-Ausgabe
pg_dump -s mydb | nosub schema -
nosub schema postgres://user:pw@host/mydb   # ruft pg_dump auf
nosub schema mysql://user:pw@host/mydb      # ruft mysqldump auf
```

Die Ausgabe ist bewusst kompakt gehalten. Tabellen mit Pipe- und Bindestrich-Gittern verschwenden unnötig Tokens und helfen dem Modell nicht:

```
## orders
- id bigint PK
- user_id bigint -> users.id ON DELETE CASCADE
- total numeric(10,2) NOT NULL DEFAULT 0
Indexes: idx_orders_user (user_id)
```

Bei großen Datenbanken erzeugt `--compact` eine Zeile pro Tabelle:

```
users(id PK, email, created_at)
orders(id PK, user_id -> users.id, total)
```

Filtere Tabellen mit `--only` und `--exclude` (unterstützt Globs) und nutze `-o schema.md`, um in eine Datei zu schreiben. Passwörter in Connection-Strings werden über Umgebungsvariablen an die Dump-Tools übergeben, niemals über CLI-Argumente.

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

Attribute werden in React-CamelCase umgewandelt, `style`-Strings werden zu Objekten und Metadaten-Reste aus Inkscape, Sketch, Figma oder Illustrator werden bereinigt. Hex-Farbcodes, die zur Tailwind-Palette passen, werden automatisch zu Utility-Klassen.

Konvertiere einen ganzen Ordner samt Indexdatei:

```sh
nosub svg ./icons -o ./src/icons --ts
```

`--current-color` ersetzt feste Farben durch `currentColor` – ideal für Icons, die sich der Textfarbe anpassen sollen.

## Einschränkungen

- `nosub env` basiert auf Mustererkennung. Es fängt typische Flüchtigkeitsfehler ab, ersetzt aber keinesfalls die sofortige Rotation eines geleakten Keys.
- Direkte Datenbank-URLs setzen voraus, dass `pg_dump` bzw. `mysqldump` im `PATH` vorhanden sind. Dump-Dateien und SQLite benötigen keine externen Tools.
- SQLite-Datenbanken im WAL-Modus haben unter Umständen noch nicht festgeschriebene Änderungen in der `-wal`-Datei. Schließe die zugreifende Anwendung vor dem Ausführen.
- `size-*`-Klassen setzen Tailwind 3.4 oder neuer voraus.

## Sprachen

Meldungen orientieren sich an der Systemsprache. Du kannst sie mit `--lang de` oder `NOSUB_LANG=de` überschreiben. Verfügbar: en, pt, es, fr, de, it, pl, ru, tr, id, hi, zh, ja, ko.

## Projekt unterstützen

nosub ist kostenlos, steht unter der MIT-Lizenz und bleibt frei. Wenn es dir ein kostenpflichtiges Tool erspart hat, kannst du das Projekt [hier](https://buymeacoffee.com/samaritansli) unterstützen.

- Lightning: `inborngrease4665@cake.cash`
- Bitcoin: `bc1qm2j0l9zy7npnhfctf80p3ukuj2ksffwlvum3c5`

## Lizenz

[MIT](../LICENSE)
