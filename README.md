# CardGame-Admin

テキサスホールデム教材プロジェクトの **管理者ページ** (Go)。
`cardgame` データベースに**直接アクセス**して、プレイヤー・チップ・お知らせ・メンテナンスを管理する。

> **提供物であって生徒課題ではない**。生徒が触るゲームサーバー ([CardGame-Server](../CardGame-Server))
> とは別リポジトリ・別コンテナで、同じ MySQL を共有する。スキーマの正は
> [CardGame-Server/Docs/DB.md](../CardGame-Server/Docs/DB.md)。管理ページはマイグレーションしない
> (テーブル作成はゲームサーバーの責務)。

## 機能

- プレイヤー一覧・検索 (id / device_id / name)
- プレイヤー詳細 + チップ台帳の閲覧
- **チップ調整**: `players.chips` の更新と `chip_transactions` への台帳記録を**同一トランザクション**で行う
  (残高と台帳を常に一致させる。裸の UPDATE をしない)
- お知らせ (`notices`) の追加・削除
- メンテナンス予定 (`maintenance_windows`) の追加・削除

## 認証

全管理ルートに **Basic 認証**。環境変数で設定する。

| 変数 | 用途 |
|---|---|
| `ADMIN_USER` | Basic 認証ユーザー名 |
| `ADMIN_PASS` | Basic 認証パスワード |
| `DATABASE_DSN` | 接続先 (必須) |
| `ADMIN_ADDR` | 待ち受けアドレス (既定 `:9090`) |

`ADMIN_USER`/`ADMIN_PASS` を両方空にすると認証なしで起動する (ローカル開発用。本番では必ず設定)。
`/healthz` は認証不要 (コンテナ healthcheck 用)。

## クイックスタート

### ローカル (Go 直接)

ゲームサーバーの MySQL が `127.0.0.1:3306` で動いている前提:

```powershell
$env:DATABASE_DSN="root:cardgame@tcp(127.0.0.1:3306)/cardgame?parseTime=true&loc=UTC"
$env:ADMIN_USER="admin"; $env:ADMIN_PASS="cardgame"
go run ./cmd/admin
# → http://localhost:9090 (admin / cardgame)
```

### Docker (単体スタック: admin + mysql)

```bash
docker compose up --build
# → http://localhost:9090
```

### 既存の MySQL に相乗りさせる (推奨: コンテナとして乗せる)

CardGame-Server 側で MySQL が起動済み (`docker compose up -d mysql`) なら、
同梱の `docker-compose.shared.yml` で admin だけをそのネットワークに相乗りさせる:

```bash
docker compose -f docker-compose.shared.yml up -d --build
# → http://localhost:9090 (既存 MySQL / 既存データに接続)
```

`docker-compose.shared.yml` は既存ネットワーク `cardgame-server_default` を外部参照し、
サービス名 `mysql` に接続する (自前で MySQL を立てない = データを共有する)。

### CardGame-Server の docker-compose.yml に直接足す場合

次の service を足しても同じ MySQL を共有して起動できる:

```yaml
  admin:
    build: ../CardGame-Admin        # パスは配置に合わせる
    ports: ["9090:9090"]
    environment:
      DATABASE_DSN: "root:cardgame@tcp(mysql:3306)/cardgame?parseTime=true&loc=UTC"
      ADMIN_USER: "admin"
      ADMIN_PASS: "cardgame"
    depends_on:
      mysql:
        condition: service_healthy
```

## 構成

```
CardGame-Admin/
├── docker-compose.yml    admin + mysql
├── Dockerfile
├── cmd/admin/            エントリポイント・ルーティング・Basic認証の配線
└── internal/
    ├── store/            DB 直アクセス (players / チップ台帳 / notices / maintenance)
    └── web/              HTTP ハンドラ + html/template (画面)
```

## 注意

- 管理ページは強い権限 (残高書き換え等) を持つ。公開ネットワークに晒さない。Basic 認証を必ず設定する
- チップ調整は台帳付き。残高だけを直接書き換える運用はしない (調査可能性を保つ)
