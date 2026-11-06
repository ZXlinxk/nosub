# nosub

[English](../README.md) · [Português](README.pt.md) · [Español](README.es.md) · [Français](README.fr.md) · [Deutsch](README.de.md) · [Italiano](README.it.md) · [Polski](README.pl.md) · [Русский](README.ru.md) · [Türkçe](README.tr.md) · [Bahasa Indonesia](README.id.md) · [हिन्दी](README.hi.md) · [中文](README.zh.md) · [日本語](README.ja.md) · [한국어](README.ko.md)

Três ferramentas locais em um binário de 3 MB. Sem conta, sem telemetria, sem mensalidade. Roda direto na sua máquina e nunca abre conexões de rede por conta própria.

- `nosub env` evita que segredos cheguem ao git
- `nosub schema` exporta o schema do banco para Markdown legível por LLMs
- `nosub svg` converte arquivos SVG em componentes limpos JSX ou Tailwind

## Instalação

Linux e macOS:

```sh
curl -fsSL https://raw.githubusercontent.com/ZXlinxk/nosub/main/install.sh | sh
```

Windows (PowerShell):

```powershell
irm https://raw.githubusercontent.com/ZXlinxk/nosub/main/install.ps1 | iex
```

Com Go:

```sh
go install github.com/ZXlinxk/nosub/cmd/nosub@latest
```

Ou baixe o binário pronto na [página de releases](https://github.com/ZXlinxk/nosub/releases). Não precisa instalar mais nada.

## nosub env

Varra um projeto, o commit prestes a acontecer ou todo o histórico do git:

```sh
nosub env scan
nosub env scan --staged
nosub env scan --history
```

```
k.txt:1  [github-token]  ghp_****
1 possible secret(s) found.
```

Os segredos aparecem mascarados no terminal, permitindo colar o log em CI com segurança. O código de saída é 0 se estiver limpo, 1 se encontrou segredos e 2 em erros. Use `--json` para integração estruturada.

Bloqueie commits perigosos de forma definitiva:

```sh
nosub env hook install
```

O hook também impede commits acidentais de arquivos `.env`. Mais dois utilitários:

```sh
nosub env example      # cria .env.example a partir de .env, sem os valores
nosub env gitignore    # adiciona regras de .env ao .gitignore
```

Reconhece credenciais de AWS, GitHub, GitLab, Slack, Stripe, OpenAI, Anthropic, Google, SendGrid, npm e PyPI, chaves privadas, JWTs, senhas em URLs e strings de alta entropia atribuídas a variáveis como `API_SECRET`.

Falso positivo? Coloque `nosub:allow` na linha ou adicione o caminho ao `.nosubignore`.

## nosub schema

```sh
nosub schema app.db                         # arquivo SQLite, sem driver externo
nosub schema dump.sql                       # pg_dump, mysqldump, phpMyAdmin ou .schema
pg_dump -s mydb | nosub schema -
nosub schema postgres://user:pw@host/mydb   # chama o pg_dump
nosub schema mysql://user:pw@host/mydb      # chama o mysqldump
```

A saída é enxuta de propósito. Tabelas cheias de barras e traços devoram tokens e não ajudam o modelo:

```
## orders
- id bigint PK
- user_id bigint -> users.id ON DELETE CASCADE
- total numeric(10,2) NOT NULL DEFAULT 0
Indexes: idx_orders_user (user_id)
```

Para bancos extensos, `--compact` mantém uma linha por tabela:

```
users(id PK, email, created_at)
orders(id PK, user_id -> users.id, total)
```

Filtre tabelas com `--only` e `--exclude` (com suporte a globs) e use `-o schema.md` para salvar em arquivo. Senhas passadas na string de conexão vão para as ferramentas de dump por variáveis de ambiente, sem expor em argumentos.

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

Atributos viram camelCase para o React, strings `style` viram objetos e sobras de editores como Inkscape, Sketch, Figma ou Illustrator são removidas. Cores hexadecimais correspondentes à paleta do Tailwind viram classes utilitárias.

Converta uma pasta inteira com arquivo de índice:

```sh
nosub svg ./icons -o ./src/icons --ts
```

`--current-color` substitui as cores por `currentColor`, ideal para ícones que acompanham a cor do texto.

## Limitações

- `nosub env` usa casamento de padrões. Pega os descuidos do dia a dia, mas não substitui a rotação de uma chave que já vazou.
- Conexões diretas por URL exigem `pg_dump` ou `mysqldump` no `PATH`. Dumps e arquivos SQLite não exigem dependências.
- Bancos SQLite em modo WAL podem ter alterações recentes ainda retidas no arquivo `-wal`. Feche o aplicativo antes de inspecionar.
- Classes `size-*` exigem Tailwind 3.4 ou mais recente.

## Idiomas

As mensagens respeitam o idioma do sistema. Altere com `--lang pt` ou `NOSUB_LANG=pt`. Disponíveis: en, pt, es, fr, de, it, pl, ru, tr, id, hi, zh, ja, ko.

## Apoie o projeto

O nosub é gratuito, open source (licença MIT) e continuará assim. Se economizou sua assinatura mensal, considere apoiar [aqui](https://buymeacoffee.com/samaritansli).

- Lightning: `inborngrease4665@cake.cash`
- Bitcoin: `bc1qm2j0l9zy7npnhfctf80p3ukuj2ksffwlvum3c5`

## Licença

[MIT](../LICENSE)
