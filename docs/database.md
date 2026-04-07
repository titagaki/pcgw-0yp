# データベース設計

SQLite3 を使用。WALモード、busy_timeout=10000ms、外部キー制約有効。

## テーブル一覧

### users

ユーザーアカウント。

| カラム | 型 | 説明 |
|--------|------|------|
| id | INTEGER PK | 自動採番 |
| name | TEXT | 表示名 |
| image | TEXT | プロフィール画像URL |
| twitter_id | TEXT UNIQUE | Twitter ID (NULL可) |
| admin | INTEGER | 管理者フラグ (0/1) |
| suspended | INTEGER | 凍結フラグ (0/1) |
| bio | TEXT | 自己紹介 (160文字以内) |
| notice_checked_at | DATETIME | お知らせ既読時刻 |
| logged_on_at | DATETIME | 最終ログイン |
| created_at | DATETIME | 作成日時 |

### servents

peercast-mi インスタンスの接続情報。

| カラム | 型 | 説明 |
|--------|------|------|
| id | INTEGER PK | 自動採番 |
| name | TEXT UNIQUE | 表示名 |
| description | TEXT | 説明 |
| hostname | TEXT | ホスト名 |
| port | INTEGER | JSON-RPCポート (通常7144) |
| auth_id | TEXT | HTTP Basic Auth ユーザー名 |
| passwd | TEXT | HTTP Basic Auth パスワード |
| priority | INTEGER | 優先度 (小さい方が優先) |
| max_channels | INTEGER | 最大チャンネル数 (0=無制限) |
| enabled | INTEGER | 有効フラグ |
| agent | TEXT | エージェント名 (getVersionInfo で取得) |
| yellow_pages | TEXT | YP名のスペース区切り (getYellowPages で取得) |

### channels

現在アクティブな配信チャンネル。配信停止時に削除される。

| カラム | 型 | 説明 |
|--------|------|------|
| id | INTEGER PK | 自動採番 |
| gnu_id | TEXT | PeerCast チャンネルID (32文字hex) |
| user_id | INTEGER FK→users | 配信者 |
| servent_id | INTEGER FK→servents | 使用サーバント |
| stream_key | TEXT | RTMP ストリームキー |
| last_active_at | DATETIME | 最後にデータ受信した時刻 |
| created_at | DATETIME | 配信開始時刻 |

### channel_infos

配信メタデータの履歴。配信終了後も残る。

| カラム | 型 | 説明 |
|--------|------|------|
| id | INTEGER PK | 自動採番 |
| user_id | INTEGER FK→users | 配信者 |
| channel | TEXT | チャンネル名 |
| genre | TEXT | ジャンル |
| description | TEXT | 説明 |
| comment | TEXT | コメント |
| url | TEXT | コンタクトURL |
| stream_type | TEXT | ストリーム種別 (現在は "FLV" のみ) |
| yp | TEXT | Yellow Page 名 |
| channel_id | INTEGER FK→channels | 紐づくアクティブチャンネル (NULL=終了済み) |
| servent_id | INTEGER FK→servents | 使用サーバント |
| source_name | TEXT | ソース名 |
| terminated_at | DATETIME | 配信終了時刻 (NULL=配信中) |
| created_at | DATETIME | 作成日時 |
| updated_at | DATETIME | 更新日時 |

### sources

ユーザーごとの配信ソース設定 (最大3つ)。

| カラム | 型 | 説明 |
|--------|------|------|
| id | INTEGER PK | 自動採番 |
| user_id | INTEGER FK→users | 所有者 |
| name | TEXT | ソース名 |
| key | TEXT | ランダムキー (8文字hex) |

### notices

管理者が作成するお知らせ。

| カラム | 型 | 説明 |
|--------|------|------|
| id | INTEGER PK | 自動採番 |
| title | TEXT | タイトル (60文字以内) |
| body | TEXT | 本文 (1000文字以内) |
| created_at | DATETIME | 作成日時 |
| updated_at | DATETIME | 更新日時 |

## ER図 (関連)

```
users 1──* channels
users 1──* channel_infos
users 1──* sources
servents 1──* channels
channels 1──1 channel_infos (channel_id)
servents 1──* channel_infos (servent_id)
```

## 元の pcgw との差分

| pcgw (Ruby) | pcgw-0yp (Go) | 理由 |
|-------------|---------------|------|
| passwords テーブル | 削除 | Twitter OAuth のみのため |
| screen_shots テーブル | 削除 | スクリーンショット機能は未実装 |
| users.twitch_id | 削除 | Twitch OAuth は未実装 |
| channels.hide_screenshots | 削除 | スクリーンショット機能は未実装 |
| channels.push_uri | 削除 | peercast-mi では不要 (RTMP URL は stream_key から構築) |
| channel_infos.primary_screen_shot_id | 削除 | スクリーンショット機能は未実装 |
| channel_infos.hide_screenshots | 削除 | 同上 |
