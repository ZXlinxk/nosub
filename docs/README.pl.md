# nosub

[English](../README.md) · [Português](README.pt.md) · [Español](README.es.md) · [Français](README.fr.md) · [Deutsch](README.de.md) · [Italiano](README.it.md) · [Polski](README.pl.md) · [Русский](README.ru.md) · [Türkçe](README.tr.md) · [Bahasa Indonesia](README.id.md) · [हिन्दी](README.hi.md) · [中文](README.zh.md) · [日本語](README.ja.md) · [한국어](README.ko.md)

Trzy małe lokalne narzędzia dev w jednym pliku binarnym 3 MB. Bez konta, bez telemetrii, bez subskrypcji. Działa lokalnie na twojej maszynie i nigdy samodzielnie nie otwiera połączeń sieciowych.

- `nosub env` chroni przed wyciekiem sekretów do gita
- `nosub schema` eksportuje schemat bazy do Markdowna pod prompt dla LLM
- `nosub svg` konwertuje pliki SVG na czyste komponenty JSX lub Tailwind

## Instalacja

Linux i macOS:

```sh
curl -fsSL https://raw.githubusercontent.com/ZXlinxk/nosub/main/install.sh | sh
```

Windows (PowerShell):

```powershell
irm https://raw.githubusercontent.com/ZXlinxk/nosub/main/install.ps1 | iex
```

Z użyciem Go:

```sh
go install github.com/ZXlinxk/nosub/cmd/nosub@latest
```

Albo pobierz gotowy plik binarny ze [strony wydań (releases)](https://github.com/ZXlinxk/nosub/releases). Nie trzeba instalować niczego więcej.

## nosub env

Przeskanuj projekt, commit przygotowywany do zatwierdzenia lub całą historię gita:

```sh
nosub env scan
nosub env scan --staged
nosub env scan --history
```

```
k.txt:1  [github-token]  ghp_****
1 possible secret(s) found.
```

Wykryte sekrety są maskowane w konsoli, dzięki czemu raport można bezpiecznie wkleić do logów CI. Kod wyjścia wynosi 0 gdy wszystko jest czyste, 1 gdy znaleziono sekrety i 2 przy błędach. Flaga `--json` zwraca format czytelny maszynowo.

Zablokuj niebezpieczne commity na stałe:

```sh
nosub env hook install
```

Hook odrzuca również pliki `.env` dodane do stagingu przez przypadek. Dwa dodatkowe polecenia pomocnicze:

```sh
nosub env example      # tworzy .env.example z .env, usuwając wartości
nosub env gitignore    # dodaje reguły .env do .gitignore
```

Rozpoznaje poświadczenia AWS, GitHub, GitLab, Slack, Stripe, OpenAI, Anthropic, Google, SendGrid, npm i PyPI, klucze prywatne, tokeny JWT, hasła w URL-ach oraz wartości o wysokiej entropii przypisane do zmiennych typu `API_SECRET`.

Fałszywy alarm (false positive)? Dodaj `nosub:allow` w danej linii lub wpisz ścieżkę do `.nosubignore`.

## nosub schema

```sh
nosub schema app.db                         # plik SQLite, bez zewnętrznych sterowników
nosub schema dump.sql                       # zrzut pg_dump, mysqldump, phpMyAdmin lub .schema
pg_dump -s mydb | nosub schema -
nosub schema postgres://user:pw@host/mydb   # wywołuje pg_dump
nosub schema mysql://user:pw@host/mydb      # wywołuje mysqldump
```

Format wyjściowy jest celowo zwięzły. Tabele ze znakami pipe i myślnikami marnują tokeny i nic nie wnoszą do zrozumienia modelu:

```
## orders
- id bigint PK
- user_id bigint -> users.id ON DELETE CASCADE
- total numeric(10,2) NOT NULL DEFAULT 0
Indexes: idx_orders_user (user_id)
```

W przypadku dużych baz flaga `--compact` generuje jedną linię na tabelę:

```
users(id PK, email, created_at)
orders(id PK, user_id -> users.id, total)
```

Filtruj tabele za pomocą `--only` i `--exclude` (obsługa dopasowań glob) i użyj `-o schema.md` do zapisu do pliku. Hasła z ciągów połączeń trafiają do narzędzi zrzutu przez zmienne środowiskowe, nigdy przez argumenty terminala.

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

Atrybuty zmieniają się w camelCase dla Reacta, właściwości `style` stają się obiektami, a zbędny kod po programach Inkscape, Sketch, Figma czy Illustrator zostaje usunięty. Kolory szesnastkowe pasujące do palety Tailwind stają się klasami utility.

Konwertuj cały katalog naraz z plikiem indeksu:

```sh
nosub svg ./icons -o ./src/icons --ts
```

Flaga `--current-color` zamienia kolory na `currentColor`, co sprawdza się idealnie w ikonach dopasowujących się do koloru tekstu.

## Ograniczenia

- `nosub env` opiera się na dopasowywaniu wzorców. Wychwytuje codzienne błędy, ale nie zastąpi natychmiastowej rotacji ujawnionego klucza.
- Bezpośrednie adresy URL baz wymagają obecności `pg_dump` lub `mysqldump` w zmiennej `PATH`. Zrzuty SQL i bazy SQLite działają bez żadnych narzędzi zewnętrznych.
- Baza SQLite w trybie WAL może posiadać ostatnie zmiany zapisane w pliku `-wal`. Przed eksportem schematu zamknij aplikację używającą bazy.
- Klasy `size-*` wymagają Tailwinda w wersji 3.4 lub nowszej.

## Języki

Komunikaty dostosowują się do języka systemu. Możesz wymusić język za pomocą `--lang pl` lub `NOSUB_LANG=pl`. Dostępne: en, pt, es, fr, de, it, pl, ru, tr, id, hi, zh, ja, ko.

## Wsparcie projektu

nosub jest darmowy, wydany na licencji MIT i taki pozostanie. Jeśli zaoszczędził ci opłat za subskrypcję, możesz wesprzeć projekt [tutaj](https://buymeacoffee.com/samaritansli).

- Lightning: `inborngrease4665@cake.cash`
- Bitcoin: `bc1qm2j0l9zy7npnhfctf80p3ukuj2ksffwlvum3c5`

## Licencja

[MIT](../LICENSE)
