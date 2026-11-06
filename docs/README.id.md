# nosub

[English](../README.md) · [Português](README.pt.md) · [Español](README.es.md) · [Français](README.fr.md) · [Deutsch](README.de.md) · [Italiano](README.it.md) · [Polski](README.pl.md) · [Русский](README.ru.md) · [Türkçe](README.tr.md) · [Bahasa Indonesia](README.id.md) · [हिन्दी](README.hi.md) · [中文](README.zh.md) · [日本語](README.ja.md) · [한국어](README.ko.md)

Tiga perkakas dev lokal dalam satu binary berukuran 4 MB. Tanpa akun, tanpa telemetri, tanpa langganan bulanan. Berjalan langsung di komputermu dan tidak pernah membuka koneksi jaringan sendiri.

- `nosub env` mencegah secret dan kunci kredensial bocor ke git
- `nosub schema` mengekspor skema database ke Markdown yang siap disalin ke LLM
- `nosub svg` mengubah file SVG menjadi komponen JSX atau Tailwind yang bersih

## Instalasi

Linux dan macOS:

```sh
curl -fsSL https://raw.githubusercontent.com/ZXlinxk/nosub/main/install.sh | sh
```

Windows (PowerShell):

```powershell
irm https://raw.githubusercontent.com/ZXlinxk/nosub/main/install.ps1 | iex
```

Dengan Go:

```sh
go install github.com/ZXlinxk/nosub/cmd/nosub@latest
```

Atau unduh langsung binary dari [halaman rilis (releases)](https://github.com/ZXlinxk/nosub/releases). Tidak ada dependensi lain yang perlu dipasang.

## nosub env

Pindai proyek, commit yang akan dibuat, atau seluruh riwayat git:

```sh
nosub env scan
nosub env scan --staged
nosub env scan --history
```

```
k.txt:1  [github-token]  ghp_****
1 possible secret(s) found.
```

Secret otomatis disensor (masking) pada output, sehingga aman disalin ke log CI. Kode keluar (exit code) bernilai 0 jika bersih, 1 jika ditemukan secret, dan 2 bila terjadi error. Opsi `--json` menyediakan format JSON yang mudah diproses skrip.

Cegah commit berbahaya untuk selamanya:

```sh
nosub env hook install
```

Hook ini juga otomatis memblokir file `.env` yang tidak sengaja masuk ke staging. Dua utilitas tambahan:

```sh
nosub env example      # buat .env.example dari .env, tanpa menyertakan nilainya
nosub env gitignore    # tambahkan aturan .env ke .gitignore
```

Mendeteksi kredensial AWS, GitHub, GitLab, Slack, Stripe, OpenAI, Anthropic, Google, SendGrid, npm, dan PyPI, private key, JWT, password di dalam URL, serta string dengan entropi tinggi yang ditetapkan ke variabel seperti `API_SECRET`.

False positive? Tambahkan `nosub:allow` pada baris terkait, atau daftarkan path pada `.nosubignore`.

## nosub schema

```sh
nosub schema app.db                         # file SQLite, tanpa driver eksternal
nosub schema dump.sql                       # output pg_dump, mysqldump, phpMyAdmin, atau .schema
pg_dump -s mydb | nosub schema -
nosub schema postgres://user:pw@host/mydb   # memanggil pg_dump
nosub schema mysql://user:pw@host/mydb      # memanggil mysqldump
```

Format output sengaja dibuat ringkas. Tabel dengan border garis dan tanda minus menghabiskan token LLM tanpa memberi nilai tambah:

```
## orders
- id bigint PK
- user_id bigint -> users.id ON DELETE CASCADE
- total numeric(10,2) NOT NULL DEFAULT 0
Indexes: idx_orders_user (user_id)
```

Untuk database berukuran besar, opsi `--compact` membuat satu baris per tabel:

```
users(id PK, email, created_at)
orders(id PK, user_id -> users.id, total)
```

Gunakan `--only` dan `--exclude` dengan pola glob untuk memilih tabel tertentu, dan gunakan `-o schema.md` untuk menyimpannya ke file. Password pada connection string diteruskan ke alat dump lewat variabel lingkungan (environment variable), bukan lewat argumen perintah terminal.

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

Atribut SVG otomatis diubah ke camelCase ala React, style string inline diubah menjadi objek, dan sisa metadata dari Inkscape, Sketch, Figma, maupun Illustrator dibersihkan total. Warna hex yang cocok dengan palet Tailwind otomatis diubah menjadi utility class.

Konversi seluruh folder sekaligus dengan file indeks:

```sh
nosub svg ./icons -o ./src/icons --ts
```

Opsi `--current-color` mengganti semua warna menjadi `currentColor`, sangat pas untuk ikon yang warnanya mengikuti warna teks.

## Batasan

- `nosub env` bekerja menggunakan pencocokan pola. Alat ini menangkap kelalaian umum sehari-hari, namun tidak menggantikan rotasi wajib atas kunci yang sempat bocor.
- Koneksi langsung ke URL database membutuhkan `pg_dump` atau `mysqldump` terpasang di `PATH`. File dump SQL dan SQLite tidak memerlukan dependensi apapun.
- Database SQLite dalam mode WAL mungkin menyimpan perubahan terbaru di dalam file `-wal`. Tutup aplikasi yang sedang memakainya sebelum mengekspor.
- Utility class `size-*` membutuhkan Tailwind 3.4 atau yang lebih baru.

## Bahasa

Pesan mengikuti bahasa sistem komputermu. Kamu bisa menimpanya dengan `--lang id` atau `NOSUB_LANG=id`. Bahasa yang tersedia: en, pt, es, fr, de, it, pl, ru, tr, id, hi, zh, ja, ko.

## Dukung proyek ini

nosub sepenuhnya gratis, berlisensi MIT, dan akan selalu seperti ini. Jika nosub membantumu menghemat biaya langganan, dukung proyek ini [di sini](https://buymeacoffee.com/samaritansli).

- Lightning: `inborngrease4665@cake.cash`
- Bitcoin: `bc1qm2j0l9zy7npnhfctf80p3ukuj2ksffwlvum3c5`

## Lisensi

[MIT](../LICENSE)
