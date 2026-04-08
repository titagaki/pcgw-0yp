# 設定ファイル

## config.toml

```toml
[server]
port = 8080                                    # HTTPサーバーポート
session_secret = "change-me-to-a-random-string" # セッション暗号化キー

[db]
path = "db/pcgw.db"                            # SQLiteデータベースパス

[twitter]
client_id = ""                                 # Twitter OAuth2 Client ID
client_secret = ""                             # Twitter OAuth2 Client Secret
redirect_url = "http://localhost:8080/auth/twitter/callback"
```

## 環境構築手順

### 1. Twitter OAuth2 アプリ登録

1. https://developer.twitter.com/en/portal/dashboard でアプリ作成
2. OAuth 2.0 を有効化
3. Redirect URL に `http://{host}:{port}/auth/twitter/callback` を設定
4. Scopes: `tweet.read`, `users.read`
5. Client ID と Client Secret を config.toml に記入

### 2. peercast-mi 設定

peercast-mi の `config.toml` で以下を確認:
- `peercast_port`: JSON-RPC ポート (デフォルト 7144)
- `rtmp_port`: RTMP ポート (デフォルト 1935)
- `admin_user` / `admin_pass`: リモート接続時の認証情報

### 3. サーバント登録

pcgw-0yp 起動後、管理者ユーザーで `/servents` から peercast-mi を登録:
- ホスト名: peercast-mi のアドレス
- ポート: JSON-RPC ポート (7144)
- 認証ID/パスワード: peercast-mi の admin_user/admin_pass (localhost なら不要)
- 「有効」にチェック
- 「サーバー情報を取得」でエージェント名とYPを取得

### 4. 最初の管理者ユーザー

初回起動時は管理者がいないため、以下のいずれかで設定:

```sql
-- SQLite で直接設定
sqlite3 db/pcgw.db "UPDATE users SET admin = 1 WHERE id = 1;"
```

### 5. 起動

```bash
# ビルド
go build -o pcgw-0yp .

# 起動
./pcgw-0yp                    # config.toml を読み込み
./pcgw-0yp /path/to/config.toml  # 設定ファイルを指定
```
