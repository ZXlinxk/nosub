# nosub

[English](../README.md) · [Português](README.pt.md) · [Español](README.es.md) · [Français](README.fr.md) · [Deutsch](README.de.md) · [Italiano](README.it.md) · [Polski](README.pl.md) · [Русский](README.ru.md) · [Türkçe](README.tr.md) · [Bahasa Indonesia](README.id.md) · [हिन्दी](README.hi.md) · [中文](README.zh.md) · [日本語](README.ja.md) · [한국어](README.ko.md)

4 MB の単一バイナリに凝縮された3つのローカル開発ツール。アカウント登録なし、テレメトリなし、月額課金なし。ローカルマシン上で直接動作し、自身で勝手に外部ネットワークへ接続することもありません。

- `nosub env` シークレットや機密情報の git への混入を防止
- `nosub schema` DB スキーマを LLM にそのまま渡せる Markdown 形式でエクスポート
- `nosub svg` SVG ファイルをクリーンな JSX や Tailwind コンポーネントに変換

## インストール

Linux および macOS:

```sh
curl -fsSL https://raw.githubusercontent.com/ZXlinxk/nosub/main/install.sh | sh
```

Windows (PowerShell):

```powershell
irm https://raw.githubusercontent.com/ZXlinxk/nosub/main/install.ps1 | iex
```

Go を使用する場合:

```sh
go install github.com/ZXlinxk/nosub/cmd/nosub@latest
```

または [リリースページ](https://github.com/ZXlinxk/nosub/releases) からビルド済みバイナリを直接取得してください。追加でインストールするものは一切ありません。

## nosub env

プロジェクト全体、直前のコミット対象、あるいは git の全履歴をスキャン:

```sh
nosub env scan
nosub env scan --staged
nosub env scan --history
```

```
k.txt:1  [github-token]  ghp_****
1 possible secret(s) found.
```

検出されたシークレットは出力時に自動でマスキングされるため、CI のログにそのまま貼っても安全です。終了コードはクリーンな状態で 0、シークレット検出時が 1、エラー時が 2 です。`--json` を指定すると機械可読な JSON を出力します。

危険なコミットを恒久的にブロック:

```sh
nosub env hook install
```

この hook は、誤ってステージングされた `.env` ファイルも自動でコミット拒否します。便利な追加コマンド:

```sh
nosub env example      # 値を除去した .env.example を .env から生成
nosub env gitignore    # .gitignore に .env 用の除外ルールを追加
```

AWS、GitHub、GitLab、Slack、Stripe、OpenAI、Anthropic、Google、SendGrid、npm、PyPI の認証情報、秘密鍵、JWT、接続 URL 内のパスワード、`API_SECRET` などの変数名に代入された高エントロピーな文字列を検知します。

誤検知が発生した場合: 該当する行に `nosub:allow` を追記するか、`.nosubignore` にファイルパスを記述してください。

## nosub schema

```sh
nosub schema app.db                         # SQLite ファイル、ドライバ不要
nosub schema dump.sql                       # pg_dump、mysqldump、phpMyAdmin、.schema の出力
pg_dump -s mydb | nosub schema -
nosub schema postgres://user:pw@host/mydb   # pg_dump を呼び出し
nosub schema mysql://user:pw@host/mydb      # mysqldump を呼び出し
```

出力は意図的に簡潔に保たれています。パイプやハイフンが並ぶ装飾テーブルは LLM のトークンを無駄に消費するだけで、モデルの理解にも寄与しません:

```
## orders
- id bigint PK
- user_id bigint -> users.id ON DELETE CASCADE
- total numeric(10,2) NOT NULL DEFAULT 0
Indexes: idx_orders_user (user_id)
```

大規模なデータベース向けに、`--compact` オプションでテーブルごとに1行にまとめることも可能です:

```
users(id PK, email, created_at)
orders(id PK, user_id -> users.id, total)
```

glob 対応の `--only` や `--exclude` でテーブルを絞り込み、`-o schema.md` でファイルへ保存できます。接続文字列のパスワードは環境変数経由でダンプツールに安全に渡され、プロセスの引数に露出することはありません。

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

属性名は React 用の camelCase に置換され、インライン `style` はオブジェクトに変換されます。Inkscape、Sketch、Figma、Illustrator などが残した不要なメタデータはすべて削除されます。Tailwind のカラーパレットに一致する HEX カラーは自動的にユーティリティクラスへ変換されます。

フォルダ内の全 SVG をインデックスファイル付きで一括変換:

```sh
nosub svg ./icons -o ./src/icons --ts
```

`--current-color` フラグは固定色を `currentColor` に置換します。文字色と連動させたいアイコンに最適です。

## 注意事項・制約

- `nosub env` はパターンマッチングで動作します。一般的なミスを未然に防ぎますが、一度漏洩したキーを直ちにローテーション（失効・再発行）する作業の代わりにはなりません。
- ライブ接続 URL からのエクスポートには、システム環境変数 `PATH` に `pg_dump` または `mysqldump` が含まれている必要があります。ダンプファイルや SQLite の処理には外部ツールは不要です。
- WAL モードの SQLite データベースでは、最新のスキーマ変更が `-wal` ファイル内に残っている場合があります。実行前に接続元のアプリを終了してください。
- `size-*` クラスの利用には Tailwind 3.4 以降が必要です。

## 多言語対応

メッセージは OS のシステム言語に自動追従します。`--lang ja` や `NOSUB_LANG=ja` で上書き可能です。対応言語: en, pt, es, fr, de, it, pl, ru, tr, id, hi, zh, ja, ko。

## プロジェクトの支援

nosub は MIT ライセンスの完全無料オープンソースです。今後も有料化することはありません。もし開発費やツールの月額サブスクを浮かせることができたなら、[こちら](https://buymeacoffee.com/samaritansli) からのサポートを歓迎します。

- Lightning: `inborngrease4665@cake.cash`
- Bitcoin: `bc1qm2j0l9zy7npnhfctf80p3ukuj2ksffwlvum3c5`

## ライセンス

[MIT](../LICENSE)
