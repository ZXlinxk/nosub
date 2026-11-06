# nosub

[English](../README.md) · [Português](README.pt.md) · [Español](README.es.md) · [Français](README.fr.md) · [Deutsch](README.de.md) · [Italiano](README.it.md) · [Polski](README.pl.md) · [Русский](README.ru.md) · [Türkçe](README.tr.md) · [Bahasa Indonesia](README.id.md) · [हिन्दी](README.hi.md) · [中文](README.zh.md) · [日本語](README.ja.md) · [한국어](README.ko.md)

Три компактных локальных dev-инструмента в одном бинарнике на 4 МБ. Без регистрации, без телеметрии, без подписок. Работает локально и никогда не выходит в сеть самостоятельно.

- `nosub env` предотвращает утечку секретов и ключей в git
- `nosub schema` экспортирует схему БД в Markdown для промптов LLM
- `nosub svg` превращает SVG в чистые компоненты JSX или Tailwind

## Установка

Linux и macOS:

```sh
curl -fsSL https://raw.githubusercontent.com/ZXlinxk/nosub/main/install.sh | sh
```

Windows (PowerShell):

```powershell
irm https://raw.githubusercontent.com/ZXlinxk/nosub/main/install.ps1 | iex
```

Через Go:

```sh
go install github.com/ZXlinxk/nosub/cmd/nosub@latest
```

Либо скачайте готовый бинарник со [страницы релизов](https://github.com/ZXlinxk/nosub/releases). Никаких дополнительных зависимостей.

## nosub env

Просканируйте проект, подготовленный коммит или всю историю репозитория:

```sh
nosub env scan
nosub env scan --staged
nosub env scan --history
```

```
k.txt:1  [github-token]  ghp_****
1 possible secret(s) found.
```

Секреты маскируются в выводе, поэтому отчет можно безопасно вставлять в логи CI. Код завершения: 0 — чисто, 1 — найдены секреты, 2 — ошибка. Флаг `--json` выводит машиночитаемый JSON.

Заблокируйте опасные коммиты навсегда:

```sh
nosub env hook install
```

Хук также не позволит случайно закоммитить `.env` файл. Еще две полезные команды:

```sh
nosub env example      # создает .env.example из .env без значений
nosub env gitignore    # добавляет правила для .env в .gitignore
```

Распознает ключи и токены AWS, GitHub, GitLab, Slack, Stripe, OpenAI, Anthropic, Google, SendGrid, npm и PyPI, приватные ключи, JWT, пароли в строках подключения (URL) и строки с высокой энтропией рядом с именами вроде `API_SECRET`.

Ложное срабатывание? Добавьте `nosub:allow` в строку или укажите путь в `.nosubignore`.

## nosub schema

```sh
nosub schema app.db                         # файл SQLite, драйвер не требуется
nosub schema dump.sql                       # вывод pg_dump, mysqldump, phpMyAdmin или .schema
pg_dump -s mydb | nosub schema -
nosub schema postgres://user:pw@host/mydb   # вызывает pg_dump
nosub schema mysql://user:pw@host/mydb      # вызывает mysqldump
```

Вывод намеренно лаконичен. Громоздкие псевдотаблицы с рамками съедают токены контекста и не несут ценности для нейросети:

```
## orders
- id bigint PK
- user_id bigint -> users.id ON DELETE CASCADE
- total numeric(10,2) NOT NULL DEFAULT 0
Indexes: idx_orders_user (user_id)
```

Для больших баз флаг `--compact` сжимает описание таблицы в одну строку:

```
users(id PK, email, created_at)
orders(id PK, user_id -> users.id, total)
```

Выбирайте нужные таблицы с помощью `--only` и `--exclude` (поддерживаются маски) и сохраняйте результат в файл через `-o schema.md`. Пароли из connection string передаются утилитам дампа через переменные окружения, не светясь в аргументах процессов.

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

Атрибуты переводятся в React-стиль (camelCase), строковые `style` становятся объектами, а мусор от Inkscape, Sketch, Figma и Illustrator полностью вычищается. HEX-цвета, совпадающие с палитрой Tailwind, автоматически становятся классами.

Пакетная конвертация папки с генерацией индексного файла:

```sh
nosub svg ./icons -o ./src/icons --ts
```

Флаг `--current-color` заменяет цвета на `currentColor` — стандарт для иконок, наследующих цвет текста.

## Ограничения

- `nosub env` работает по паттернам. Он ловит типовые оплошности, но не заменит отзыв и смену скомпрометированного ключа.
- Подключение к БД по живому URL требует наличия `pg_dump` или `mysqldump` в переменной `PATH`. Дампы и файлы SQLite работают автономно.
- База SQLite в режиме WAL может содержать свежие изменения в файле `-wal`. Закройте использующее ее приложение перед вызовом.
- Классы `size-*` требуют Tailwind 3.4 или новее.

## Языки

Язык сообщений совпадает с системным. Его можно переопределить через `--lang ru` или `NOSUB_LANG=ru`. Поддерживаются: en, pt, es, fr, de, it, pl, ru, tr, id, hi, zh, ja, ko.

## Поддержка проекта

nosub бесплатен, распространяется под лицензией MIT и останется таковым. Если он сэкономил вам деньги на подписках, вы можете поддержать проект [здесь](https://buymeacoffee.com/samaritansli).

- Lightning: `inborngrease4665@cake.cash`
- Bitcoin: `bc1qm2j0l9zy7npnhfctf80p3ukuj2ksffwlvum3c5`

## Лицензия

[MIT](../LICENSE)
