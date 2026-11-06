# nosub

[English](../README.md) · [Português](README.pt.md) · [Español](README.es.md) · [Français](README.fr.md) · [Deutsch](README.de.md) · [Italiano](README.it.md) · [Polski](README.pl.md) · [Русский](README.ru.md) · [Türkçe](README.tr.md) · [Bahasa Indonesia](README.id.md) · [हिन्दी](README.hi.md) · [中文](README.zh.md) · [日本語](README.ja.md) · [한국어](README.ko.md)

Trois petits outils dev locaux dans un binaire de 4 Mo. Sans compte, sans télémétrie, sans abonnement. Tourne sur votre machine et n'ouvre jamais de connexion réseau de lui-même.

- `nosub env` empêche les secrets de fuiter dans git
- `nosub schema` exporte un schéma de base de données en Markdown prêt à copier-coller dans un LLM
- `nosub svg` convertit vos fichiers SVG en composants JSX ou Tailwind propres

## Installation

Linux et macOS :

```sh
curl -fsSL https://raw.githubusercontent.com/ZXlinxk/nosub/main/install.sh | sh
```

Windows (PowerShell) :

```powershell
irm https://raw.githubusercontent.com/ZXlinxk/nosub/main/install.ps1 | iex
```

Avec Go :

```sh
go install github.com/ZXlinxk/nosub/cmd/nosub@latest
```

Ou téléchargez directement un binaire sur la [page des releases](https://github.com/ZXlinxk/nosub/releases). Rien d'autre à installer.

## nosub env

Analysez un projet, un commit imminent ou l'ensemble de l'historique git :

```sh
nosub env scan
nosub env scan --staged
nosub env scan --history
```

```
k.txt:1  [github-token]  ghp_****
1 possible secret(s) found.
```

Les secrets sont masqués à l'écran, ce qui permet de coller les logs CI sans risque. Code de sortie : 0 si propre, 1 si des secrets sont détectés, 2 en cas d'erreur. Utilisez `--json` pour un format exploitable par machine.

Bloquez définitivement les commits à risque :

```sh
nosub env hook install
```

Le hook refuse également tout fichier `.env` indexé par inadvertance. Deux commandes pratiques supplémentaires :

```sh
nosub env example      # génère .env.example depuis .env, sans les valeurs
nosub env gitignore    # ajoute les règles .env au .gitignore
```

Reconnaît les identifiants AWS, GitHub, GitLab, Slack, Stripe, OpenAI, Anthropic, Google, SendGrid, npm et PyPI, les clés privées, les JWTs, les mots de passe dans les URLs et les chaînes à haute entropie affectées à des variables comme `API_SECRET`.

Faux positif ? Ajoutez `nosub:allow` sur la ligne ou listez le chemin dans `.nosubignore`.

## nosub schema

```sh
nosub schema app.db                         # fichier SQLite, aucun driver requis
nosub schema dump.sql                       # sortie pg_dump, mysqldump, phpMyAdmin ou .schema
pg_dump -s mydb | nosub schema -
nosub schema postgres://user:pw@host/mydb   # appelle pg_dump
nosub schema mysql://user:pw@host/mydb      # appelle mysqldump
```

La sortie est volontairement concise. Les tableaux pleins de pipes et de tirets gaspillent des tokens sans rien apporter au modèle :

```
## orders
- id bigint PK
- user_id bigint -> users.id ON DELETE CASCADE
- total numeric(10,2) NOT NULL DEFAULT 0
Indexes: idx_orders_user (user_id)
```

Pour les bases volumineuses, `--compact` résume chaque table sur une seule ligne :

```
users(id PK, email, created_at)
orders(id PK, user_id -> users.id, total)
```

Filtrez les tables avec `--only` et `--exclude` (support des globs) et utilisez `-o schema.md` pour sauvegarder dans un fichier. Les mots de passe des chaînes de connexion sont transmis aux outils de dump via les variables d'environnement, jamais dans les arguments de ligne de commande.

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

Les attributs passent au format React, les styles inline deviennent des objets et toutes les métadonnées inutiles laissées par Inkscape, Sketch, Figma ou Illustrator sont purgées. Les couleurs hexadécimales correspondant à la palette Tailwind sont converties en classes utilitaires.

Convertissez un dossier complet avec un fichier d'index :

```sh
nosub svg ./icons -o ./src/icons --ts
```

`--current-color` remplace les couleurs par `currentColor`, ce qui est idéal pour les icônes devant hériter de la couleur du texte.

## Limites

- `nosub env` repose sur la détection de motifs. Il attrape les erreurs courantes mais ne peut pas tout détecter : il ne remplace pas la révocation et la rotation immédiate d'une clé compromise.
- Les URLs de bases de données distantes nécessitent `pg_dump` ou `mysqldump` dans votre `PATH`. Les dumps et fichiers SQLite n'ont aucune dépendance.
- Une base SQLite en mode WAL peut contenir des modifications récentes encore dans le fichier `-wal`. Fermez l'application avant l'analyse.
- Les classes `size-*` nécessitent Tailwind 3.4 ou supérieur.

## Langues

Les messages suivent la langue de votre système. Forcez avec `--lang fr` ou `NOSUB_LANG=fr`. Disponibles : en, pt, es, fr, de, it, pl, ru, tr, id, hi, zh, ja, ko.

## Soutenir le projet

nosub est gratuit, sous licence MIT, et le restera. S'il vous a évité un abonnement mensuel, vous pouvez soutenir le projet [ici](https://buymeacoffee.com/samaritansli).

- Lightning: `inborngrease4665@cake.cash`
- Bitcoin: `bc1qm2j0l9zy7npnhfctf80p3ukuj2ksffwlvum3c5`

## Licence

[MIT](../LICENSE)
