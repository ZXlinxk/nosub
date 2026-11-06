# nosub

[English](../README.md) · [Português](README.pt.md) · [Español](README.es.md) · [Français](README.fr.md) · [Deutsch](README.de.md) · [Italiano](README.it.md) · [Polski](README.pl.md) · [Русский](README.ru.md) · [Türkçe](README.tr.md) · [Bahasa Indonesia](README.id.md) · [हिन्दी](README.hi.md) · [中文](README.zh.md) · [日本語](README.ja.md) · [한국어](README.ko.md)

Tres herramientas de desarrollo locales en un binario de 3 MB. Sin cuentas, sin telemetría, sin suscripciones. Corre en tu máquina y nunca abre conexiones de red por su cuenta.

- `nosub env` evita que se filtren secretos a git
- `nosub schema` exporta el esquema de la base de datos a Markdown listo para pasar a un LLM
- `nosub svg` convierte archivos SVG en componentes limpios JSX o Tailwind

## Instalación

Linux y macOS:

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

O descarga el binario directamente desde la [página de releases](https://github.com/ZXlinxk/nosub/releases). No hay nada más que instalar.

## nosub env

Analiza un proyecto, el commit a punto de realizarse o todo el historial de git:

```sh
nosub env scan
nosub env scan --staged
nosub env scan --history
```

```
k.txt:1  [github-token]  ghp_****
1 possible secret(s) found.
```

Los secretos se enmascaran en la salida para poder compartir el log en CI sin riesgos. El código de salida es 0 si está limpio, 1 si se detectaron secretos y 2 en caso de error. `--json` proporciona salida estructurada.

Bloquea commits indebidos de forma definitiva:

```sh
nosub env hook install
```

El hook también rechaza archivos `.env` agregados al stage por descuido. Dos utilidades más:

```sh
nosub env example      # genera .env.example desde .env, sin los valores
nosub env gitignore    # agrega reglas de .env al .gitignore
```

Reconoce credenciales de AWS, GitHub, GitLab, Slack, Stripe, OpenAI, Anthropic, Google, SendGrid, npm y PyPI, claves privadas, JWTs, contraseñas en URLs y valores de alta entropía asignados a variables como `API_SECRET`.

¿Falso positivo? Añade `nosub:allow` en la línea o incluye la ruta en `.nosubignore`.

## nosub schema

```sh
nosub schema app.db                         # archivo SQLite, sin drivers externos
nosub schema dump.sql                       # salida de pg_dump, mysqldump, phpMyAdmin o .schema
pg_dump -s mydb | nosub schema -
nosub schema postgres://user:pw@host/mydb   # llama a pg_dump
nosub schema mysql://user:pw@host/mydb      # llama a mysqldump
```

La salida es intencionadamente compacta. Las tablas con barras y guiones devoran tokens sin aportar contexto al modelo:

```
## orders
- id bigint PK
- user_id bigint -> users.id ON DELETE CASCADE
- total numeric(10,2) NOT NULL DEFAULT 0
Indexes: idx_orders_user (user_id)
```

Para bases de datos grandes, `--compact` mantiene una línea por tabla:

```
users(id PK, email, created_at)
orders(id PK, user_id -> users.id, total)
```

Usa `--only` y `--exclude` con globs para filtrar tablas, y `-o schema.md` para escribir a un archivo. Las contraseñas en URLs de conexión se pasan a las herramientas de volcado vía variables de entorno, nunca por argumentos de terminal.

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

Los atributos pasan a camelCase para React, los estilos en string se transforman en objetos y se elimina cualquier rastro dejado por Inkscape, Sketch, Figma o Illustrator. Los colores hexadecimales que coinciden con la paleta de Tailwind se convierten en clases utilitarias.

Convierte una carpeta entera generando un archivo índice:

```sh
nosub svg ./icons -o ./src/icons --ts
```

`--current-color` sustituye los colores por `currentColor`, ideal para iconos que deben heredar el color del texto.

## Límites

- `nosub env` funciona mediante coincidencia de patrones. Detecta los errores comunes, pero no reemplaza la rotación obligatoria de una clave filtrada.
- Las URLs de bases de datos remotas requieren `pg_dump` o `mysqldump` en el `PATH`. Los archivos de volcado y SQLite no requieren nada.
- Una base de datos SQLite en modo WAL puede contener cambios recientes aún retenidos en el archivo `-wal`. Conviene cerrar la app antes de analizarla.
- Las clases `size-*` requieren Tailwind 3.4 o superior.

## Idiomas

Los mensajes se adaptan al idioma de tu sistema. Puedes forzarlo con `--lang es` o `NOSUB_LANG=es`. Disponibles: en, pt, es, fr, de, it, pl, ru, tr, id, hi, zh, ja, ko.

## Apoya el proyecto

nosub es gratuito y con licencia MIT, y siempre lo será. Si te ahorró pagar una suscripción, puedes colaborar [aquí](https://buymeacoffee.com/samaritansli).

- Lightning: `inborngrease4665@cake.cash`
- Bitcoin: `bc1qm2j0l9zy7npnhfctf80p3ukuj2ksffwlvum3c5`

## Licencia

[MIT](../LICENSE)
