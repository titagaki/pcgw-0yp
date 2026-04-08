# CLAUDE.md

## プロジェクト概要

pcgw-0yp は PeerCast 配信管理 Web アプリケーション。
Ruby/Sinatra 製の [pcgw](/home/megan/src/pcgw) を Go に移植したもの。
バックエンドの PeerCast ノードには [peercast-mi](/home/megan/src/go/peercast-mi) を使用する。

## ビルド・実行

```bash
docker compose up -d           # MySQL起動
go build -o pcgw-0yp .        # ビルド
go vet ./...                   # 静的解析
./pcgw-0yp                     # 起動 (config.toml を読み込み)
./pcgw-0yp /path/to/config.toml
```

## 技術スタック

- Go 1.22+ / chi v5 / html/template / MySQL 8.0 (go-sql-driver/mysql)
- 認証: Twitter OAuth2 (golang.org/x/oauth2)
- セッション: gorilla/sessions (cookie, 30日)
- PeerCast通信: peercast-mi JSON-RPC 2.0 (internal/peercast/client.go)
- CSS: Bulma 0.9.4 (CDN)

## ディレクトリ構成

- `internal/config/` - TOML設定読み込み
- `internal/db/` - DB接続・スキーマ (MySQL, InnoDB)
- `internal/domain/` - エンティティ (構造体定義のみ、依存なし)
- `internal/repository/` - データアクセス層 (関数ベース, `func Xxx(db *sql.DB, ...) error`)
- `internal/usecase/` - ビジネスロジック (配信開始・停止・クリーンアップ等)
- `internal/peercast/` - peercast-mi JSON-RPCクライアント
- `internal/handler/` - HTTPハンドラー (`Handler` 構造体のメソッド)
- `internal/middleware/` - セッション・認証・CSRF
- `internal/server/` - ルーティング定義
- `internal/view/` - テンプレート (templ)
- `templates/` - HTMLテンプレート (layout + content パターン)
- `public/` - 静的ファイル
- `docs/` - 詳細ドキュメント
  - `docs/design/` - システム設計 (アーキテクチャ, DB, ルーティング)
  - `docs/guides/` - 運用・移行ガイド (設定, 移行)
  - `docs/reference/` - 外部APIリファレンス (PeerCast API)
  - `docs/planning/` - 今後の計画・未実装機能

## コーディング規約

### ハンドラー
- `Handler` 構造体のメソッドとして定義
- `h.renderTempl(w, r, component)` でレンダリング
- `h.flash(w, r, "msg")` でフラッシュメッセージ
- `middleware.CurrentUser(r)` で現在ユーザー取得
- 単純なCRUDは `repository` を直接呼ぶ。複雑なオーケストレーションは `usecase` 経由

### ドメイン (domain)
- 構造体定義のみ。外部依存なし

### リポジトリ (repository)
- グローバル状態なし。`*sql.DB` を第1引数で受け取る
- エラーはそのまま返す (ハンドラー側で処理)

### ユースケース (usecase)
- 複数の repository/peercast 呼び出しを組み合わせたビジネスロジック
- 配信開始 (`StartBroadcast`)、停止 (`StopChannel`)、クリーンアップ (`RunCleanup`) 等

### テンプレート
- templ コンポーネントベース
- `view.PageData` に `.User`, `.LoggedIn`, `.CSRFToken` を `pageData()` が自動注入

### PeerCast API 呼び出し
- `h.peercastClient(servent)` でクライアント取得
- エラー型: `*peercast.Unavailable` (接続不可), `*peercast.rpcError` (API エラー)

## 配信フロー (peercast-mi 固有)

1. `issueStreamKey(accountName, streamKey)` - キー発行
2. `broadcastChannel({streamKey, info, track})` - チャンネル作成
3. ユーザーが `rtmp://host:1935/live/{streamKey}` にプッシュ
4. `stopChannel(channelId)` + `revokeStreamKey(accountName)` - 停止

## 未実装機能

スクリーンショット, BBS連携, Twitch OAuth, パスワード認証, リピーター, IP日本語化。
詳細は `docs/planning/future.md` を参照。
