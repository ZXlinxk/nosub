# nosub

[English](../README.md) · [Português](README.pt.md) · [Español](README.es.md) · [Français](README.fr.md) · [Deutsch](README.de.md) · [Italiano](README.it.md) · [Polski](README.pl.md) · [Русский](README.ru.md) · [Türkçe](README.tr.md) · [Bahasa Indonesia](README.id.md) · [हिन्दी](README.hi.md) · [中文](README.zh.md) · [日本語](README.ja.md) · [한국어](README.ko.md)

3 MB'lık tek bir binary içinde üç küçük yerel geliştirici aracı. Hesap yok, telemetri yok, aylık abonelik yok. Doğrudan kendi makinenizde çalışır ve asla kendi başına ağ bağlantısı açmaz.

- `nosub env` gizli anahtarların git'e sızmasını önler
- `nosub schema` veritabanı şemasını LLM'e yapıştırmaya uygun Markdown olarak dışa aktarır
- `nosub svg` SVG dosyalarını temiz JSX veya Tailwind bileşenlerine dönüştürür

## Kurulum

Linux ve macOS:

```sh
curl -fsSL https://raw.githubusercontent.com/ZXlinxk/nosub/main/install.sh | sh
```

Windows (PowerShell):

```powershell
irm https://raw.githubusercontent.com/ZXlinxk/nosub/main/install.ps1 | iex
```

Go ile:

```sh
go install github.com/ZXlinxk/nosub/cmd/nosub@latest
```

Veya doğrudan [sürümler sayfasından](https://github.com/ZXlinxk/nosub/releases) binary dosyasını indirin. Başka hiçbir şey kurmanız gerekmez.

## nosub env

Bir projeyi, yapılmak üzere olan commit'i veya tüm git geçmişini tarayın:

```sh
nosub env scan
nosub env scan --staged
nosub env scan --history
```

```
k.txt:1  [github-token]  ghp_****
1 possible secret(s) found.
```

Gizli anahtarlar çıktıda maskelenir; böylece raporu CI loglarına güvenle yapıştırabilirsiniz. Çıkış kodu temiz olduğunda 0, bir anahtar bulunduğunda 1 ve hata durumunda 2'dir. `--json` makine tarafından okunabilir çıktı verir.

Hatalı commit'leri kalıcı olarak engelleyin:

```sh
nosub env hook install
```

Hook ayrıca yanlışlıkla stage'e eklenen `.env` dosyalarını da reddeder. İki yardımcı komut daha:

```sh
nosub env example      # değerleri içermeyen .env.example dosyasını .env'den üretir
nosub env gitignore    # .gitignore dosyasına .env kurallarını ekler
```

AWS, GitHub, GitLab, Slack, Stripe, OpenAI, Anthropic, Google, SendGrid, npm ve PyPI kimlik bilgilerini, özel anahtarları (private keys), JWT'leri, URL içi şifreleri ve `API_SECRET` gibi değişkenlere atanan yüksek entropili dizgileri tanır.

Yanlış alarm mı? Satıra `nosub:allow` ekleyin veya yolu `.nosubignore` dosyasına yazın.

## nosub schema

```sh
nosub schema app.db                         # SQLite dosyası, harici sürücü gerekmez
nosub schema dump.sql                       # pg_dump, mysqldump, phpMyAdmin veya .schema çıktısı
pg_dump -s mydb | nosub schema -
nosub schema postgres://user:pw@host/mydb   # pg_dump çağırır
nosub schema mysql://user:pw@host/mydb      # mysqldump çağırır
```

Çıktı bilerek sade ve kısadır. Çizgiler ve tirelerle dolu tablolar LLM token'larını tüketir ve modele hiçbir fayda sağlamaz:

```
## orders
- id bigint PK
- user_id bigint -> users.id ON DELETE CASCADE
- total numeric(10,2) NOT NULL DEFAULT 0
Indexes: idx_orders_user (user_id)
```

Büyük veritabanları için `--compact` tablo başına tek satır üretir:

```
users(id PK, email, created_at)
orders(id PK, user_id -> users.id, total)
```

Tabloları filtrelemek için glob destekli `--only` ve `--exclude` bayraklarını, dosyaya kaydetmek için ise `-o schema.md` bayrağını kullanın. Bağlantı dizesindeki parolalar dump araçlarına komut satırı argümanı olarak değil, ortam değişkeni üzerinden iletilir.

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

Öznitelikler React camelCase formatına döner, `style` dizgileri nesneye dönüştürülür ve Inkscape, Sketch, Figma veya Illustrator'dan kalan tüm gereksiz kodlar temizlenir. Tailwind paletiyle eşleşen hex renkler otomatik olarak class'lara dönüştürülür.

Dizindeki tüm simgeleri bir index dosyasıyla birlikte dönüştürün:

```sh
nosub svg ./icons -o ./src/icons --ts
```

`--current-color` bayrağı renkleri `currentColor` ile değiştirir; metin rengini devralması gereken simgeler için tam olarak aradığınız çözümdür.

## Sınırlar

- `nosub env` desen eşleme (pattern matching) yöntemiyle çalışır. Yaygın dikkatsizlikleri yakalar ancak sızdırılmış bir anahtarı hemen geçersiz kılmanın (rotate) yerini tutmaz.
- Canlı veritabanı URL'leri `PATH` içinde `pg_dump` veya `mysqldump` bulunmasını gerektirir. Dump dosyaları ve SQLite hiçbir harici bağımlılık gerektirmez.
- WAL modundaki bir SQLite veritabanında son değişiklikler henüz `-wal` dosyasında bulunuyor olabilir. Analiz etmeden önce veritabanını kullanan uygulamayı kapatın.
- `size-*` sınıfları Tailwind 3.4 veya üzerini gerektirir.

## Diller

Mesajlar sistem dilinize uyar. `--lang tr` veya `NOSUB_LANG=tr` ile değiştirebilirsiniz. Desteklenen diller: en, pt, es, fr, de, it, pl, ru, tr, id, hi, zh, ja, ko.

## Projeyi destekleyin

nosub ücretsizdir, MIT lisanslıdır ve öyle kalacaktır. Size bir abonelik ücreti tasarruf ettirdiyse, [buradan](https://buymeacoffee.com/samaritansli) destek olabilirsiniz.

- Lightning: `inborngrease4665@cake.cash`
- Bitcoin: `bc1qm2j0l9zy7npnhfctf80p3ukuj2ksffwlvum3c5`

## Lisans

[MIT](../LICENSE)
