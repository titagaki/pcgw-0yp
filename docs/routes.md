# ルーティング一覧

## 公開ページ (認証不要)

| メソッド | パス | ハンドラー | 説明 |
|----------|------|-----------|------|
| GET | `/` | TopPage | トップページ (ログイン済みは /home へリダイレクト) |
| GET | `/login` | LoginPage | ログインページ |
| GET | `/auth/twitter` | TwitterLogin | Twitter OAuth2 認証開始 |
| GET | `/auth/twitter/callback` | TwitterCallback | OAuth2 コールバック |
| GET | `/logout` | Logout | ログアウト |
| GET | `/stats` | Stats | 統計情報 (全サーバントのチャンネル一覧) |
| GET | `/profile` | ProfileList | アクティブユーザー一覧 / 検索 |
| GET | `/profile/{id}` | ProfileShow | ユーザープロフィール |
| GET | `/programs` | ProgramIndex | 番組表トップ |
| GET | `/programs/recent` | ProgramRecent | 最近の配信一覧 |
| GET | `/programs/by-date/{year}/{month}` | ProgramByMonth | 月別配信一覧 |
| GET | `/programs/{id}` | ProgramShow | 配信詳細 |
| GET | `/api/1/channelStatus?name=` | APIChannelStatus | JSON API (CORS有効) |
| GET | `/public/*` | FileServer | 静的ファイル |

## 認証済みページ

| メソッド | パス | ハンドラー | 説明 |
|----------|------|-----------|------|
| GET | `/home` | HomePage | ダッシュボード |
| GET | `/create` | CreatePage | 配信作成フォーム |
| POST | `/broadcast` | Broadcast | 配信開始 |
| GET | `/channels` | ChannelList | チャンネル一覧 |
| GET | `/channels/{id}` | ChannelShow | チャンネル詳細 |
| GET | `/channels/{id}/edit` | ChannelEdit | チャンネル編集フォーム |
| POST | `/channels/{id}` | ChannelUpdate | チャンネル情報更新 |
| POST | `/channels/{id}/stop` | ChannelStop | 配信停止 |
| GET | `/channels/{id}/relay_tree` | ChannelRelayTree | リレーツリー |
| GET | `/channels/{id}/connections` | ChannelConnections | 接続一覧 |
| POST | `/channels/{id}/connections/{connID}/disconnect` | ChannelDisconnect | 接続切断 |
| GET | `/channels/{id}/status.json` | ChannelStatusJSON | ステータスJSON |
| GET | `/profile/edit` | ProfileEdit | プロフィール編集フォーム |
| POST | `/profile/edit` | ProfileUpdate | プロフィール更新 |
| GET | `/notices` | NoticeIndex | お知らせ一覧 |
| GET | `/notices/{id}` | NoticeShow | お知らせ詳細 |
| GET | `/sources` | SourceIndex | ソース一覧 |
| GET | `/sources/add?name=` | SourceAdd | ソース追加 |
| GET | `/sources/del?id=` | SourceDelete | ソース削除 |
| GET | `/sources/regen?id=` | SourceRegen | ソースキー再生成 |
| POST | `/programs/{id}/delete` | ProgramDelete | 配信履歴削除 |

## 管理者専用ページ

| メソッド | パス | ハンドラー | 説明 |
|----------|------|-----------|------|
| GET | `/admin` | AdminIndex | 管理画面トップ |
| GET | `/users` | UserList | ユーザー一覧 |
| GET | `/users/{id}` | UserShow | ユーザー詳細 |
| GET | `/users/{id}/edit` | UserEdit | ユーザー編集フォーム |
| POST | `/users/{id}` | UserUpdate | ユーザー更新 |
| POST | `/users/{id}/delete` | UserDelete | ユーザー削除 |
| GET | `/servents` | ServentIndex | サーバント一覧 |
| POST | `/servents` | ServentCreate | サーバント追加 |
| GET | `/servents/{id}` | ServentShow | サーバント詳細 |
| POST | `/servents/{id}` | ServentUpdate | サーバント更新 |
| POST | `/servents/{id}/delete` | ServentDelete | サーバント削除 |
| POST | `/servents/{id}/refresh` | ServentRefresh | サーバント情報再取得 |
| GET | `/notices/new` | NoticeNew | お知らせ作成フォーム |
| POST | `/notices` | NoticeCreate | お知らせ作成 |
| GET | `/notices/{id}/edit` | NoticeEdit | お知らせ編集フォーム |
| POST | `/notices/{id}/update` | NoticeUpdate | お知らせ更新 |
| POST | `/notices/{id}/delete` | NoticeDelete | お知らせ削除 |

## 元の pcgw との差分

### 削除されたルート

| パス | 理由 |
|------|------|
| `/auth/twitch/*` | Twitch OAuth 未実装 |
| `/account/*` | パスワード管理不要 (OAuth のみ) |
| `/channels/{id}/create_repeater` | リピーター機能未実装 |
| `/channels/{id}/start_repeater` | 同上 |
| `/channels/{id}/stop_repeater` | 同上 |
| `/channels/{id}/thread_list` | BBS連携未実装 |
| `/programs/{id}/screen_shots` | スクリーンショット未実装 |
| `/programs/{id}/digest` | BBS連携未実装 |
| `/bbs/*` | BBS連携未実装 |
| `/ip/*` | IP日本語化未実装 |
| `/doc/*` | ドキュメントページ未実装 |

### 変更されたルート

| pcgw | pcgw-0yp | 変更点 |
|------|----------|--------|
| `DELETE /channels/{id}/connections/{id}` | `POST .../disconnect` | HTMLフォームから直接使えるようPOSTに変更 |
| `DELETE /programs/{id}` | `POST /programs/{id}/delete` | 同上 |
| `DELETE /users/{id}` | `POST /users/{id}/delete` | 同上 |
| `PATCH /servents/{id}` | `POST /servents/{id}` | 同上 |
