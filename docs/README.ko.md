# nosub

[English](../README.md) · [Português](README.pt.md) · [Español](README.es.md) · [Français](README.fr.md) · [Deutsch](README.de.md) · [Italiano](README.it.md) · [Polski](README.pl.md) · [Русский](README.ru.md) · [Türkçe](README.tr.md) · [Bahasa Indonesia](README.id.md) · [हिन्दी](README.hi.md) · [中文](README.zh.md) · [日本語](README.ja.md) · [한국어](README.ko.md)

4 MB 단일 바이너리에 담긴 3가지 로컬 개발 도구. 계정 불필요, 텔레메트리 없음, 월간 구독 없음. 사용자 로컬 머신에서만 작동하며 자체적인 네트워크 연결을 절대 열지 않습니다.

- `nosub env` 비밀 정보 및 API 키가 git에 커밋되는 것을 방지
- `nosub schema` LLM에 바로 붙여넣기 좋은 깔끔한 Markdown으로 DB 스키마 추출
- `nosub svg` SVG 파일을 군더더기 없는 JSX 또는 Tailwind 컴포넌트로 변환

## 설치

Linux 및 macOS:

```sh
curl -fsSL https://raw.githubusercontent.com/ZXlinxk/nosub/main/install.sh | sh
```

Windows (PowerShell):

```powershell
irm https://raw.githubusercontent.com/ZXlinxk/nosub/main/install.ps1 | iex
```

Go를 사용하는 경우:

```sh
go install github.com/ZXlinxk/nosub/cmd/nosub@latest
```

또는 [릴리스 페이지](https://github.com/ZXlinxk/nosub/releases)에서 사전 빌드된 바이너리를 직접 다운로드하세요. 추가 설치가 필요한 의존성은 없습니다.

## nosub env

프로젝트 전체, 커밋 대기 중인 파일, 혹은 전체 git 커밋 히스토리를 스캔:

```sh
nosub env scan
nosub env scan --staged
nosub env scan --history
```

```
k.txt:1  [github-token]  ghp_****
1 possible secret(s) found.
```

출력 시 비밀 정보는 마스킹 처리되므로 CI 빌드 로그에 안심하고 붙여넣을 수 있습니다. 종료 코드는 이상 없을 때 0, 비밀 정보 발견 시 1, 에러 발생 시 2입니다. `--json` 플래그로 파싱하기 쉬운 JSON 출력을 제공합니다.

위험한 커밋을 영구 차단:

```sh
nosub env hook install
```

이 훅은 실수로 스테이징된 `.env` 파일의 커밋도 자동으로 거부합니다. 추가 편의 도구:

```sh
nosub env example      # .env에서 실제 값을 지우고 .env.example 생성
nosub env gitignore    # .gitignore에 .env 제외 규칙 추가
```

AWS, GitHub, GitLab, Slack, Stripe, OpenAI, Anthropic, Google, SendGrid, npm, PyPI 인증 정보, 개인 키, JWT, URL 내부의 비밀번호 및 `API_SECRET` 등과 같은 변수에 할당된 고엔트로피 문자열을 감지합니다.

오탐(False positive)인가요? 해당 라인에 `nosub:allow`를 적거나 `.nosubignore`에 파일 경로를 등록하세요.

## nosub schema

```sh
nosub schema app.db                         # SQLite 파일, 드라이버 불필요
nosub schema dump.sql                       # pg_dump, mysqldump, phpMyAdmin 또는 .schema 출력
pg_dump -s mydb | nosub schema -
nosub schema postgres://user:pw@host/mydb   # pg_dump 호출
nosub schema mysql://user:pw@host/mydb      # mysqldump 호출
```

출력 결과는 의도적으로 매우 간결하게 작성됩니다. 파이프와 대시가 가득한 복잡한 표는 LLM의 토큰만 낭비할 뿐 모델의 맥락 이해에 도움이 되지 않습니다:

```
## orders
- id bigint PK
- user_id bigint -> users.id ON DELETE CASCADE
- total numeric(10,2) NOT NULL DEFAULT 0
Indexes: idx_orders_user (user_id)
```

대형 데이터베이스의 경우, `--compact` 플래그로 테이블당 한 줄 요약을 생성할 수 있습니다:

```
users(id PK, email, created_at)
orders(id PK, user_id -> users.id, total)
```

glob 패턴을 지원하는 `--only`와 `--exclude`로 테이블을 필터링하고, `-o schema.md`로 파일에 저장할 수 있습니다. 연결 문자열의 비밀번호는 CLI 인자가 아니라 환경 변수를 통해 덤프 도구로 안전하게 전달됩니다.

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

속성은 React camelCase로 바뀌고, 인라인 `style`은 객체로 변환되며, Inkscape, Sketch, Figma, Illustrator가 남긴 불필요한 메타데이터는 깨끗이 제거됩니다. Tailwind 팔레트와 일치하는 Hex 색상은 자동으로 유틸리티 클래스로 변환됩니다.

폴더 전체를 인덱스 파일과 함께 한 번에 변환:

```sh
nosub svg ./icons -o ./src/icons --ts
```

`--current-color` 플래그는 색상을 `currentColor`로 치환하여 텍스트 색상과 연동되어야 하는 아이콘 컴포넌트에 유용합니다.

## 제약 사항

- `nosub env`는 패턴 매칭 기반으로 동작합니다. 일상적인 실수를 효과적으로 방지하지만, 유출된 실제 키를 즉시 재발급(rotate)하는 것을 대체할 수는 없습니다.
- 라이브 데이터베이스 URL을 분석하려면 시스템 `PATH`에 `pg_dump` 또는 `mysqldump`가 있어야 합니다. 덤프 파일과 SQLite는 외부 도구 없이 즉시 동작합니다.
- WAL 모드의 SQLite 데이터베이스는 최근 스키마 변경 사항이 `-wal` 파일에 남아 있을 수 있습니다. 분석 전에 연결된 애플리케이션을 닫으세요.
- `size-*` 클래스는 Tailwind 3.4 이상이 필요합니다.

## 언어 설정

출력 메시지는 시스템 기본 언어를 따릅니다. `--lang ko` 또는 `NOSUB_LANG=ko`로 변경할 수 있습니다. 지원 언어: en, pt, es, fr, de, it, pl, ru, tr, id, hi, zh, ja, ko.

## 프로젝트 후원

nosub은 MIT 라이선스를 따르는 완전 무료 오픈소스이며 앞으로도 계속 무료로 제공됩니다. 만약 유료 구독 비용을 절약하는 데 도움이 되었다면, [여기](https://buymeacoffee.com/samaritansli)에서 후원하실 수 있습니다.

- Lightning: `inborngrease4665@cake.cash`
- Bitcoin: `bc1qm2j0l9zy7npnhfctf80p3ukuj2ksffwlvum3c5`

## 라이선스

[MIT](../LICENSE)
