# nosub

[English](../README.md) · [Português](README.pt.md) · [Español](README.es.md) · [Français](README.fr.md) · [Deutsch](README.de.md) · [Italiano](README.it.md) · [Polski](README.pl.md) · [Русский](README.ru.md) · [Türkçe](README.tr.md) · [Bahasa Indonesia](README.id.md) · [हिन्दी](README.hi.md) · [中文](README.zh.md) · [日本語](README.ja.md) · [한국어](README.ko.md)

एक ही 3 MB बाइनरी में तीन छोटे लोकल देव टूल्स। कोई अकाउंट नहीं, कोई टेलीमेट्री नहीं, कोई मासिक सब्सक्रिप्शन नहीं। यह आपकी मशीन पर चलता है और कभी भी खुद से कोई नेटवर्क कनेक्शन नहीं खोलता।

- `nosub env` सीक्रेट्स और कीज़ को git में जाने से रोकता है
- `nosub schema` डेटाबेस स्कीमा को Markdown में एक्सपोर्ट करता है जिसे आप सीधे LLM में पेस्ट कर सकते हैं
- `nosub svg` SVG फाइलों को साफ JSX या Tailwind कंपोनेंट्स में बदलता है

## इंस्टॉलेशन

Linux और macOS:

```sh
curl -fsSL https://raw.githubusercontent.com/ZXlinxk/nosub/main/install.sh | sh
```

Windows (PowerShell):

```powershell
irm https://raw.githubusercontent.com/ZXlinxk/nosub/main/install.ps1 | iex
```

Go के साथ:

```sh
go install github.com/ZXlinxk/nosub/cmd/nosub@latest
```

या सीधे [रिलीज़ पेज](https://github.com/ZXlinxk/nosub/releases) से बाइनरी डाउनलोड करें। इसके अलावा कुछ भी इंस्टॉल करने की ज़रूरत नहीं है।

## nosub env

किसी प्रोजेक्ट, होने वाले नए कमिट, या पूरे git इतिहास को स्कैन करें:

```sh
nosub env scan
nosub env scan --staged
nosub env scan --history
```

```
k.txt:1  [github-token]  ghp_****
1 possible secret(s) found.
```

आउटपुट में सीक्रेट्स को मास्क कर दिया जाता है, जिससे रिपोर्ट को CI लॉग्स में सुरक्षित रूप से पेस्ट किया जा सकता है। एग्जिट कोड: साफ़ होने पर 0, सीक्रेट मिलने पर 1, और एरर पर 2। `--json` मशीन-रीडेबल आउटपुट देता है।

गलत कमिट्स को हमेशा के लिए ब्लॉक करें:

```sh
nosub env hook install
```

यह हुक गलती से स्टेज की गई `.env` फ़ाइल को भी रिजेक्ट कर देता है। दो अतिरिक्त उपयोगी टूल्स:

```sh
nosub env example      # वैल्यूज़ हटाकर .env से .env.example बनाता है
nosub env gitignore    # .gitignore में .env के नियम जोड़ता है
```

यह AWS, GitHub, GitLab, Slack, Stripe, OpenAI, Anthropic, Google, SendGrid, npm और PyPI क्रेडेंशियल्स, प्राइवेट कीज़, JWTs, URLs में मौजूद पासवर्ड्स, और `API_SECRET` जैसे वेरिएबल्स को असाइन की गई हाई-एंट्रॉपी स्ट्रिंग्स को पहचानता है।

फॉल्स पॉजिटिव? लाइन में `nosub:allow` जोड़ें, या उस पाथ को `.nosubignore` में लिख दें।

## nosub schema

```sh
nosub schema app.db                         # SQLite फ़ाइल, किसी ड्राइवर की आवश्यकता नहीं
nosub schema dump.sql                       # pg_dump, mysqldump, phpMyAdmin या .schema आउटपुट
pg_dump -s mydb | nosub schema -
nosub schema postgres://user:pw@host/mydb   # pg_dump को कॉल करता है
nosub schema mysql://user:pw@host/mydb      # mysqldump को कॉल करता है
```

आउटपुट जानबूझकर छोटा रखा गया है। पाइप और डैश वाली बड़ी टेबल टोकन खाती हैं और मॉडल को कुछ नया नहीं सिखातीं:

```
## orders
- id bigint PK
- user_id bigint -> users.id ON DELETE CASCADE
- total numeric(10,2) NOT NULL DEFAULT 0
Indexes: idx_orders_user (user_id)
```

बड़े डेटाबेस के लिए, `--compact` प्रति टेबल एक लाइन रखता है:

```
users(id PK, email, created_at)
orders(id PK, user_id -> users.id, total)
```

टेबल चुनने के लिए globs के साथ `--only` और `--exclude` का उपयोग करें, और फ़ाइल में लिखने के लिए `-o schema.md` लगाएं। कनेक्शन स्ट्रिंग के पासवर्ड एनवायरनमेंट वेरिएबल्स के जरिए डंप टूल्स को भेजे जाते हैं, कमांड लाइन आर्गुमेंट्स में नहीं।

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

एट्रिब्यूट्स React camelCase में बदल जाते हैं, इनलाइन `style` ऑब्जेक्ट बन जाते हैं, और Inkscape, Sketch, Figma या Illustrator द्वारा छोड़े गए सभी कचरे को हटा दिया जाता है। Tailwind पैलेट से मेल खाने वाले Hex कलर्स अपने आप क्लास बन जाते हैं।

इंडेक्स फ़ाइल के साथ पूरे फ़ोल्डर को कनवर्ट करें:

```sh
nosub svg ./icons -o ./src/icons --ts
```

`--current-color` रंगों को `currentColor` से बदल देता है, जो उन आइकन्स के लिए सबसे अच्छा है जिन्हें टेक्स्ट के रंग का पालन करना होता है।

## सीमाएं

- `nosub env` पैटर्न मैचिंग पर काम करता है। यह सामान्य गलतियों को पकड़ता है, लेकिन यह लीक हुई कुंजी को रोटेट करने का विकल्प नहीं है।
- लाइव डेटाबेस URLs के लिए आपके `PATH` में `pg_dump` या `mysqldump` होना ज़रूरी है। डंप फाइलों और SQLite को किसी बाहरी टूल की आवश्यकता नहीं है।
- WAL मोड में SQLite डेटाबेस में हाल के बदलाव हो सकते हैं जो अभी भी `-wal` फ़ाइल में हैं। उपयोग करने से पहले उस ऐप को बंद करें।
- `size-*` क्लासेज के लिए Tailwind 3.4 या नए वर्जन की आवश्यकता होती है।

## भाषाएं

मैसेज आपके सिस्टम की भाषा का पालन करते हैं। इसे `--lang hi` या `NOSUB_LANG=hi` के साथ बदलें। उपलब्ध भाषाएं: en, pt, es, fr, de, it, pl, ru, tr, id, hi, zh, ja, ko.

## प्रोजेक्ट का समर्थन करें

nosub पूरी तरह मुफ़्त और MIT लाइसेंस प्राप्त है, और यह हमेशा ऐसा ही रहेगा। अगर इसने आपका सब्सक्रिप्शन खर्च बचाया है, तो आप [यहाँ](https://buymeacoffee.com/samaritansli) के माध्यम से समर्थन कर सकते हैं।

- Lightning: `inborngrease4665@cake.cash`
- Bitcoin: `bc1qm2j0l9zy7npnhfctf80p3ukuj2ksffwlvum3c5`

## लाइसेंस

[MIT](../LICENSE)
