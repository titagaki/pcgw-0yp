# pcgw (Ruby) からの移行ガイド

元の pcgw と pcgw-0yp の対応関係のリファレンス。

## ファイル対応表

### エントリーポイント・設定

| pcgw (Ruby) | pcgw-0yp (Go) |
|-------------|---------------|
| `jimson.rb` (メインアプリ) | `main.go` + `internal/server/router.go` |
| `init.rb` | `internal/config/config.go` |
| `config/config.yml` | `config.toml` |
| `Schemafile` (Ridgepole) | `internal/db/schema.go` |
| `Gemfile` | `go.mod` |

### モデル

| pcgw (Ruby) | pcgw-0yp (Go) |
|-------------|---------------|
| `models/user.rb` | `internal/model/user.go` |
| `models/channel.rb` | `internal/model/channel.go` |
| `models/channel_info.rb` | `internal/model/channel_info.go` |
| `models/servent.rb` | `internal/model/servent.go` |
| `models/source.rb` | `internal/model/source.go` |
| `models/notice.rb` | `internal/model/notice.go` |
| `models/password.rb` | 未実装 (OAuth のみ) |
| `models/screen_shot.rb` | 未実装 |
| `models/yellow_page.rb` | 不要 (peercast-mi から動的取得) |
| `models/genre.rb` | 未実装 (今後必要に応じて) |
| `models/connection.rb` | `internal/peercast/client.go` の `Connection` 型 |
| `models/relay_tree.rb` | `internal/peercast/client.go` の `RelayTreeNode` 型 |
| `models/digest.rb` | 未実装 (BBS連携なし) |

### ルート

| pcgw (Ruby) | pcgw-0yp (Go) |
|-------------|---------------|
| `routes/main.rb` | `internal/handler/home.go` + `stats.go` |
| `routes/login.rb` | `internal/handler/auth.go` |
| `routes/broadcast.rb` | `internal/handler/broadcast.go` |
| `routes/channels.rb` | `internal/handler/channel.go` |
| `routes/account.rb` | 未実装 |
| `routes/profile.rb` | `internal/handler/profile.go` |
| `routes/programs.rb` | `internal/handler/program.go` |
| `routes/servents.rb` | `internal/handler/servent.go` |
| `routes/admin.rb` | `internal/handler/admin.go` |
| `routes/notices.rb` | `internal/handler/notice.go` |
| `routes/sources.rb` | `internal/handler/source.go` |
| `routes/api.rb` | `internal/handler/api.go` |
| `routes/bbs.rb` | 未実装 |
| `routes/tests.rb` | 不要 |

### ライブラリ

| pcgw (Ruby) | pcgw-0yp (Go) |
|-------------|---------------|
| `lib/peercast.rb` (Jimson RPC) | `internal/peercast/client.go` |
| `lib/logging.rb` | `log/slog` (標準ライブラリ) |
| `lib/nipponize.rb` | 未実装 |
| `lib/yarr_client.rb` | 未実装 |
| `lib/bbs_reader.rb` | 未実装 |
| `lib/core_ext.rb` | 不要 (Go標準機能で代替) |

### ヘルパー

| pcgw (Ruby) | pcgw-0yp (Go) |
|-------------|---------------|
| `helpers/helpers.rb` | `internal/middleware/auth.go` + `internal/view/helpers.go` |
| `helpers/view_helpers.rb` | `internal/view/helpers.go` |
| `helpers/time_util.rb` | `internal/view/helpers.go` の時刻フォーマット関数 |
| `helpers/relay_tree_renderer.rb` | テンプレートで直接表示 |
| `helpers/graphviz.rb` | 未実装 |

### ビュー

| pcgw (Ruby/Slim) | pcgw-0yp (Go/html) |
|-------------------|---------------------|
| `views/layout.slim` | `templates/layout.html` |
| `views/navbar.slim` | `templates/partials/navbar.html` |
| `views/top.slim` | `templates/top.html` |
| `views/home.slim` | `templates/home.html` |
| `views/login.slim` | `templates/login.html` |
| `views/create.erb` | `templates/create.html` |
| `views/status.slim` | `templates/channel.html` |
| `views/edit.erb` | `templates/channel_edit.html` |
| `views/profile.slim` | `templates/profile.html` |
| `views/profile_edit.slim` | `templates/profile_edit.html` |
| `views/active_users.slim` | `templates/profiles.html` |
| `views/programs.slim` | `templates/programs.html` |
| `views/program.slim` | `templates/program.html` |
| `views/stats.slim` | `templates/stats.html` |
| `views/servent_index.slim` | `templates/servents.html` |
| `views/servent.slim` | `templates/servent.html` |
| `views/users.erb` | `templates/users.html` |
| `views/user_edit.erb` | `templates/user_edit.html` |
| `views/notice_index.slim` | `templates/notices.html` |
| `views/source_index.slim` | `templates/sources.html` |

## 配信フローの違い

### pcgw (PeerCast YT)

```
1. POST /broadcast
2. servent.api.fetch(url: "rtmp://mirror/live/key", name: ..., genre: ..., ...)
   → PeerCast が URL から PULL してチャンネル作成
3. Channel レコード作成 (push_uri, stream_key 保存)
4. ユーザーが push_uri に RTMP/HTTP push
```

### pcgw-0yp (peercast-mi)

```
1. POST /broadcast
2. client.IssueStreamKey("user_123", "sk_xxx...")
3. client.BroadcastChannel({streamKey: "sk_xxx...", info: ..., track: ...})
   → peercast-mi がチャンネル作成、RTMP 待ち受け開始
4. Channel レコード作成 (stream_key 保存)
5. ユーザーが rtmp://host:1935/live/sk_xxx... に RTMP push
```

**主な違い:**
- pcgw: PeerCast が外部URLから PULL → ミラーサーバー (WM_MIRROR) 経由
- pcgw-0yp: ユーザーが peercast-mi に直接 RTMP push → ミラーサーバー不要
- pcgw: ストリームキーはアプリ側で管理
- pcgw-0yp: ストリームキーは peercast-mi 側で管理 (issueStreamKey API)

## DB マイグレーション

pcgw のデータを pcgw-0yp に移行する場合:

```sql
-- users: twitter_id の型変更 (INTEGER → TEXT)
INSERT INTO users_new (id, name, image, twitter_id, admin, suspended, bio, notice_checked_at, logged_on_at, created_at)
SELECT id, name, image, CAST(twitter_id AS TEXT), admin, suspended, bio, notice_checked_at, logged_on_at, created_at
FROM users_old;

-- channel_infos: desc → description カラム名変更
INSERT INTO channel_infos_new (..., description, ...)
SELECT ..., desc, ...
FROM channel_infos_old;

-- servents: desc → description カラム名変更
INSERT INTO servents_new (..., description, ...)
SELECT ..., desc, ...
FROM servents_old;

-- channels: push_uri 削除、hide_screenshots 削除
-- sources, notices: そのまま移行可能
```
