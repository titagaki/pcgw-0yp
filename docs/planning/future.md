# 今後の実装予定

元の pcgw から未移植の機能と、新規追加が望ましい機能のリスト。

## 未移植機能 (元の pcgw にあったもの)

### 優先度: 高

#### スクリーンショット機能
- **元の実装**: 配信中のストリームからスクリーンショットを取得・保存
- **必要なもの**:
  - `screen_shots` テーブル追加
  - `channel_infos.primary_screen_shot_id` カラム追加
  - スクリーンショット取得・保存ロジック (ffmpeg でストリームからキャプチャ)
  - ファイル配信 (ネストしたディレクトリ構造: `/ss/ab/cd/abcdef...`)
  - 古いファイルの定期削除 (1時間以上)
- **関連ファイル**: `internal/model/screen_shot.go`, `internal/handler/screenshot.go`
- **テンプレート**: スクリーンショットギャラリー, サムネイル表示

#### BBS (したらば) 連携
- **元の実装**: コンタクトURL がしたらば掲示板の場合、スレッド一覧・投稿を表示
- **必要なもの**:
  - `internal/bbs/reader.go` - したらばのHTML解析 (Board, Thread, Post)
  - `GET /channels/{id}/thread_list` ルート
  - `GET /bbs/latest-thread?board_url=` ルート
  - `GET /bbs/info?url=` ルート
  - ProgramDigest (スクリーンショットとBBS投稿の時刻同期表示)
- **注意**: したらばのHTML構造に依存するため、変更時にパーサーの修正が必要

#### Twitch OAuth 認証
- **元の実装**: omniauth-twitch による認証
- **必要なもの**:
  - `users.twitch_id` カラム追加
  - OAuth2 フロー実装 (Twitter と同様のパターン)
  - config.toml に `[twitch]` セクション追加
  - ログインページにTwitchボタン追加

### 優先度: 中

#### パスワード認証
- **元の実装**: SHA256+salt によるパスワードハッシュ
- **必要なもの**:
  - `passwords` テーブル追加
  - `internal/model/password.go` - bcrypt でハッシュ化 (SHA256 から移行推奨)
  - `POST /login` ルート (パスワードログイン)
  - `/account/change-password` ルート
  - Twitter/Twitch アカウント解除機能 (パスワード設定済みの場合のみ)

#### リピーター (RTMP リレー) 機能
- **元の実装**: FLV配信を外部サービス (Twitch, YouTube) にリレー
- **必要なもの**:
  - Yarr デーモンとの連携 (JSON-RPC localhost:8100) または代替実装
  - `GET /channels/{id}/create_repeater` - リピーター作成フォーム
  - `POST /channels/{id}/start_repeater` - リピーター開始
  - `GET /channels/{id}/stop_repeater` - リピーター停止
  - プロセス管理 (ffmpeg ベースが現実的)
- **代替案**: peercast-mi 側にリレー機能を追加する方が合理的かもしれない

#### アカウント管理画面
- **元の実装**: `/account` でパスワード変更、Twitter/Twitch アカウント連携解除
- **必要なもの**: パスワード認証の実装後に追加

#### IP アドレス日本語化
- **元の実装**: IPv4 を日本語音節にエンコード・デコード
- **必要なもの**:
  - `internal/nipponize/nipponize.go`
  - `GET /ip/info/{ip}`, `GET /ip/decode/{text}` ルート
  - 接続一覧での表示

### 優先度: 低

#### ドキュメントページ
- **元の実装**: OBS, WME, Expression Encoder 等の配信設定ガイド
- **必要なもの**: `/doc/{name}` ルート + テンプレート
- peercast-mi に合わせた OBS 設定ガイドを新規作成するのが良い

#### リレーツリー可視化 (GraphViz)
- **元の実装**: GraphViz を使ったリレーツリーの画像レンダリング
- **現在の実装**: テキストベースのツリー表示
- **改善案**: D3.js 等で JavaScript ベースの可視化に移行

---

## 新規追加が望ましい機能

### WebSocket リアルタイム更新
- チャンネルステータスの自動更新 (現在はページリロードが必要)
- `gorilla/websocket` または SSE (Server-Sent Events)
- チャンネル画面の接続状況・リスナー数をリアルタイム表示

### API 拡充
- 現在は `GET /api/1/channelStatus?name=` のみ
- RESTful API の追加:
  - `GET /api/1/channels` - チャンネル一覧
  - `GET /api/1/channels/{id}` - チャンネル詳細
  - `POST /api/1/channels` - 配信開始 (API経由)
  - `DELETE /api/1/channels/{id}` - 配信停止 (API経由)
- 認証: API キーまたは Bearer トークン

### テスト
- ユニットテスト: model 層の CRUD テスト
- 統合テスト: handler のHTTPテスト (`httptest`)
- peercast-mi クライアントのモック

### Docker 対応
- Dockerfile 作成
- docker-compose.yml (pcgw-0yp + peercast-mi)

### ログ改善
- 構造化ログ (slog) は実装済み
- アクセスログのファイル出力
- ログレベル設定の config.toml 対応

### セキュリティ改善
- HTTPS 対応 (Let's Encrypt / リバースプロキシ)
- Rate limiting
- セッションストアの Redis 移行 (マルチインスタンス対応)
- Content Security Policy ヘッダー

---

## コードレビュー残件 (低優先度, 2026-04-08)

### #15 PKCE が plain method
- `internal/handler/auth.go` TwitterLogin
- `code_challenge_method: "plain"` は PKCE の保護効果がほぼない
- `S256` (SHA-256) に変更すべき

### #16 ChannelStatusJSON の Content-Type 設定順序
- `internal/handler/channel.go` ChannelStatusJSON
- エラー時に `WriteHeader` してから JSON を書いているが Content-Type ヘッダーが未設定
- `WriteHeader` の前に `Content-Type` を設定する必要がある

### #17 Logout が GET メソッド
- `internal/server/router.go`
- CSRF 保護が効かないため、外部サイトの `<img src="/logout">` でログアウトさせられる
- POST に変更し、ログアウトボタンをフォームにする

### #18 /channels がログイン必須
- `internal/server/router.go`, `internal/middleware/auth.go`
- `/channels` が `publicPrefixes` に含まれないためログイン必須
- 意図的かどうか確認して、公開すべきなら prefix を追加する

---

## 実装時の注意事項

### テンプレートパターン
既存テンプレートは `{{define "content"}}...{{end}}{{template "layout" .}}` パターンを使用。
新しいテンプレートもこのパターンに従う。

### モデル層パターン
- 関数シグネチャ: `func GetXxx(db *sql.DB, ...) (*Xxx, error)`
- DB は引数で受け取る (グローバル変数にしない)
- エラーは呼び出し元に返す

### ハンドラーパターン
- `Handler` 構造体のメソッドとして定義
- `h.render(w, r, "template.html", data)` でレンダリング
- `h.flash(w, r, "メッセージ")` でフラッシュメッセージ
- `middleware.CurrentUser(r)` で現在のユーザー取得
- `h.peercastClient(servent)` で PeerCast クライアント取得

### peercast-mi API 呼び出しパターン
```go
servent, _ := model.GetServent(h.DB, serventID)
client := h.peercastClient(servent)
result, err := client.SomeMethod(args)
if err != nil {
    // Unavailable エラーの場合は接続失敗
    // rpcError の場合は API エラー
}
```
