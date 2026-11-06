# nosub

[English](../README.md) · [Português](README.pt.md) · [Español](README.es.md) · [Français](README.fr.md) · [Deutsch](README.de.md) · [Italiano](README.it.md) · [Polski](README.pl.md) · [Русский](README.ru.md) · [Türkçe](README.tr.md) · [Bahasa Indonesia](README.id.md) · [हिन्दी](README.hi.md) · [中文](README.zh.md) · [日本語](README.ja.md) · [한국어](README.ko.md)

Tre piccoli strumenti di sviluppo locali in un singolo eseguibile da 3 MB. Nessun account, nessuna telemetria, nessun abbonamento. Gira sulla tua macchina e non apre mai connessioni di rete di sua iniziativa.

- `nosub env` impedisce ai segreti di finire nei repository git
- `nosub schema` esporta lo schema del database in formato Markdown ottimizzato per i prompt degli LLM
- `nosub svg` trasforma i file SVG in componenti JSX o Tailwind puliti

## Installazione

Linux e macOS:

```sh
curl -fsSL https://raw.githubusercontent.com/ZXlinxk/nosub/main/install.sh | sh
```

Windows (PowerShell):

```powershell
irm https://raw.githubusercontent.com/ZXlinxk/nosub/main/install.ps1 | iex
```

Con Go:

```sh
go install github.com/ZXlinxk/nosub/cmd/nosub@latest
```

Oppure scarica direttamente il binario dalla [pagina dei release](https://github.com/ZXlinxk/nosub/releases). Non c'è nient'altro da installare.

## nosub env

Scansiona un progetto, il commit che stai per creare o l'intera cronologia di git:

```sh
nosub env scan
nosub env scan --staged
nosub env scan --history
```

```
k.txt:1  [github-token]  ghp_****
1 possible secret(s) found.
```

I segreti vengono mascherati nell'output, rendendo il log sicuro da includere nei log della CI. Il codice di uscita è 0 se pulito, 1 se sono stati individuati segreti e 2 in caso di errore. L'opzione `--json` restituisce un output leggibile da script.

Blocca i commit pericolosi una volta per tutte:

```sh
nosub env hook install
```

L'hook blocca anche i file `.env` aggiunti allo staging per sbaglio. Altre due utili utility:

```sh
nosub env example      # genera .env.example da .env, rimuovendo i valori
nosub env gitignore    # aggiunge le regole per .env a .gitignore
```

Riconosce credenziali AWS, GitHub, GitLab, Slack, Stripe, OpenAI, Anthropic, Google, SendGrid, npm e PyPI, chiavi private, JWT, password incorporate negli URL e stringhe ad alta entropia assegnate a identificatori come `API_SECRET`.

Falso positivo? Inserisci `nosub:allow` sulla riga o aggiungi il percorso a `.nosubignore`.

## nosub schema

```sh
nosub schema app.db                         # file SQLite, nessun driver esterno
nosub schema dump.sql                       # output di pg_dump, mysqldump, phpMyAdmin o .schema
pg_dump -s mydb | nosub schema -
nosub schema postgres://user:pw@host/mydb   # richiama pg_dump
nosub schema mysql://user:pw@host/mydb      # richiama mysqldump
```

L'output è volutamente essenziale. Le tabelle piene di barre e trattini consumano token inutilmente senza aiutare il modello:

```
## orders
- id bigint PK
- user_id bigint -> users.id ON DELETE CASCADE
- total numeric(10,2) NOT NULL DEFAULT 0
Indexes: idx_orders_user (user_id)
```

Per database di grandi dimensioni, `--compact` mantiene una sola riga per tabella:

```
users(id PK, email, created_at)
orders(id PK, user_id -> users.id, total)
```

Filtra le tabelle con `--only` ed `--exclude` (supportano pattern glob) e usa `-o schema.md` per salvare su file. Le password passate nelle stringhe di connessione vengono inoltrate agli strumenti di dump tramite variabili d'ambiente, mai tramite argomenti da riga di comando.

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

Gli attributi vengono convertiti nel camelCase di React, gli stili inline diventano oggetti e tutti i residui lasciati da Inkscape, Sketch, Figma o Illustrator vengono eliminati. I codici colore esadecimali compatibili con la palette di Tailwind diventano classi utility.

Converti un'intera cartella generando un file indice:

```sh
nosub svg ./icons -o ./src/icons --ts
```

`--current-color` sostituisce i colori fissi con `currentColor`, la scelta naturale per icone che devono seguire il colore del testo.

## Limiti

- `nosub env` funziona tramite riconoscimento di pattern. Intercetta gli errori frequenti, ma non può sostituire la revoca e la rotazione di una chiave compromessa.
- Gli URL di database richiedono che `pg_dump` o `mysqldump` siano disponibili nel `PATH`. I file di dump e SQLite non richiedono strumenti esterni.
- Un database SQLite con journaling WAL potrebbe contenere modifiche recenti ancora confinate nel file `-wal`. Chiudi l'applicazione prima di ispezionare il database.
- Le classi `size-*` richiedono Tailwind 3.4 o successivo.

## Lingue

I messaggi seguono la lingua di sistema. Puoi forzarla con `--lang it` o `NOSUB_LANG=it`. Disponibili: en, pt, es, fr, de, it, pl, ru, tr, id, hi, zh, ja, ko.

## Sostieni il progetto

nosub è gratuito, rilasciato con licenza MIT e rimarrà tale. Se ti ha evitato di pagare un abbonamento, puoi offrire un contributo [qui](https://buymeacoffee.com/samaritansli).

- Lightning: `inborngrease4665@cake.cash`
- Bitcoin: `bc1qm2j0l9zy7npnhfctf80p3ukuj2ksffwlvum3c5`

## Licenza

[MIT](../LICENSE)
