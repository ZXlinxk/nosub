# nosub

[English](../README.md) · [Português](README.pt.md) · [Español](README.es.md) · [Français](README.fr.md) · [Deutsch](README.de.md) · [Italiano](README.it.md) · [Polski](README.pl.md) · [Русский](README.ru.md) · [Türkçe](README.tr.md) · [Bahasa Indonesia](README.id.md) · [हिन्दी](README.hi.md) · [中文](README.zh.md) · [日本語](README.ja.md) · [한국어](README.ko.md)

仅 4 MB 的单文件本地开发工具包，集成三大实用功能。无需注册，无遥测，无月费。完全在你的本地机器运行，从不主动发起网络连接。

- `nosub env` 防止密钥意外提交到 git
- `nosub schema` 将数据库结构导出为适合直接粘贴给 LLM 的 Markdown
- `nosub svg` 将 SVG 文件转换为干净的 JSX 或 Tailwind 组件

## 安装

Linux 与 macOS：

```sh
curl -fsSL https://raw.githubusercontent.com/ZXlinxk/nosub/main/install.sh | sh
```

Windows (PowerShell)：

```powershell
irm https://raw.githubusercontent.com/ZXlinxk/nosub/main/install.ps1 | iex
```

使用 Go 安装：

```sh
go install github.com/ZXlinxk/nosub/cmd/nosub@latest
```

或直接从 [Releases 发布页](https://github.com/ZXlinxk/nosub/releases) 下载编译好的二进制文件。无需安装任何额外依赖。

## nosub env

扫描整个项目、即将提交的暂存内容或完整的 git 历史：

```sh
nosub env scan
nosub env scan --staged
nosub env scan --history
```

```
k.txt:1  [github-token]  ghp_****
1 possible secret(s) found.
```

输出中的密钥均会自动打码遮蔽，可安全粘贴至 CI 构建日志中。退出码：未发现为 0，发现密钥为 1，发生错误为 2。使用 `--json` 输出结构化数据。

彻底拦截违规提交：

```sh
nosub env hook install
```

此 hook 还会拦截被误加入暂存区的 `.env` 文件。另外两个实用辅助工具：

```sh
nosub env example      # 从 .env 生成剔除敏感值的 .env.example
nosub env gitignore    # 将 .env 规则添加到 .gitignore
```

支持识别 AWS、GitHub、GitLab、Slack、Stripe、OpenAI、Anthropic、Google、SendGrid、npm 与 PyPI 凭证、私钥、JWT、URL 中的内嵌密码，以及赋值给类似 `API_SECRET` 的高熵字符串。

遇到误报？在行尾添加 `nosub:allow`，或者在 `.nosubignore` 中列出文件路径。

## nosub schema

```sh
nosub schema app.db                         # SQLite 文件，无需驱动
nosub schema dump.sql                       # pg_dump、mysqldump、phpMyAdmin 或 .schema 输出
pg_dump -s mydb | nosub schema -
nosub schema postgres://user:pw@host/mydb   # 自动调用 pg_dump
nosub schema mysql://user:pw@host/mydb      # 自动调用 mysqldump
```

输出格式特意保持极简。由管道线与破折号构成的复杂表格既浪费 token，又对模型理解毫无助益：

```
## orders
- id bigint PK
- user_id bigint -> users.id ON DELETE CASCADE
- total numeric(10,2) NOT NULL DEFAULT 0
Indexes: idx_orders_user (user_id)
```

对于大型数据库，`--compact` 参数可将每张表压缩为一行：

```
users(id PK, email, created_at)
orders(id PK, user_id -> users.id, total)
```

使用带有通配符的 `--only` 和 `--exclude` 筛选表，使用 `-o schema.md` 将结果保存到文件。连接串中的密码通过环境变量安全传递给 dump 工具，绝不会暴露在命令行参数中。

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

SVG 属性自动转换为 React 的 camelCase，内联 `style` 转换为对象，并彻底清除 Inkscape、Sketch、Figma 或 Illustrator 留下的冗余元数据。与 Tailwind 调色板匹配的十六进制颜色会自动转为 Tailwind 类名。

批量转换整个目录并生成索引文件：

```sh
nosub svg ./icons -o ./src/icons --ts
```

`--current-color` 参数可将颜色替换为 `currentColor`，非常适合需要跟随文本颜色的图标组件。

## 局限与说明

- `nosub env` 基于模式匹配工作。它能捕获日常粗心犯下的错误，但绝不能替代对已泄露密钥的立即轮换。
- 连接实时数据库需要系统 `PATH` 中具备 `pg_dump` 或 `mysqldump`。SQL dump 导出文件与 SQLite 无需外部依赖。
- WAL 模式下的 SQLite 数据库可能将最新更改暂存在 `-wal` 文件中，分析前请先关闭正在连接的应用程序。
- `size-*` 类名需要 Tailwind 3.4 或更高版本。

## 多语言支持

提示信息默认跟随操作系统语言。可通过 `--lang zh` 或环境变量 `NOSUB_LANG=zh` 覆盖。支持语言：en, pt, es, fr, de, it, pl, ru, tr, id, hi, zh, ja, ko。

## 支持本项目

nosub 完全免费且采用 MIT 协议开源，并且永远保持免费。如果它为你省去了一笔付费订阅，欢迎[在这里](https://buymeacoffee.com/samaritansli)赞助。

- Lightning: `inborngrease4665@cake.cash`
- Bitcoin: `bc1qm2j0l9zy7npnhfctf80p3ukuj2ksffwlvum3c5`

## 开源协议

[MIT](../LICENSE)
