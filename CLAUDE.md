# CLAUDE.md

## プロジェクト概要

pcgw-0yp は PeerCast 配信管理 Web アプリケーション。
Ruby/Sinatra 製の [pcgw](/home/megan/src/pcgw) を Go に移植したもの。
バックエンドの PeerCast ノードには [peercast-mi](/home/megan/src/go/peercast-mi) を使用する。

## ビルド・実行

```bash
go build -o pcgw-0yp .        # ビルド
go vet ./...                   # 静的解析
./pcgw-0yp                     # 起動 (config.toml を読み込み)
./pcgw-0yp /path/to/config.toml
```

CGO が必要 (go-sqlite3)。`CGO_ENABLED=1` を確認すること。

## 技術スタック

- Go 1.22+ / chi v5 / html/template / SQLite3 (mattn/go-sqlite3)
- 認証: Twitter OAuth2 (golang.org/x/oauth2)
- セッション: gorilla/sessions (cookie, 30日)
- PeerCast通信: peercast-mi JSON-RPC 2.0 (internal/peercast/client.go)
- CSS: Bulma 0.9.4 (CDN)

## ディレクトリ構成

- `internal/config/` - TOML設定読み込み
- `internal/db/` - DB接続・スキーマ (SQLite WAL, FK有効)
- `internal/model/` - データアクセス層 (関数ベース, `func XxxModel(db *sql.DB, ...) error`)
- `internal/peercast/` - peercast-mi JSON-RPCクライアント
- `internal/handler/` - HTTPハンドラー (`Handler` 構造体のメソッド)
- `internal/middleware/` - セッション・認証・CSRF
- `internal/server/` - ルーティング定義
- `internal/view/` - テンプレート関数
- `templates/` - HTMLテンプレート (layout + content パターン)
- `public/` - 静的ファイル
- `docs/` - 詳細ドキュメント

## コーディング規約

### ハンドラー
- `Handler` 構造体のメソッドとして定義
- `h.render(w, r, "template.html", data)` でレンダリング
- `h.flash(w, r, "msg")` でフラッシュメッセージ
- `middleware.CurrentUser(r)` で現在ユーザー取得

### モデル
- グローバル状態なし。`*sql.DB` を第1引数で受け取る
- エラーはそのまま返す (ハンドラー側で処理)

### テンプレート
- `{{define "content"}}...{{end}}{{template "layout" .}}` パターン
- データは `map[string]interface{}` で渡す
- `.User`, `.LoggedIn`, `.CSRFToken` は `render()` が自動注入

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
詳細は `docs/future.md` を参照。
