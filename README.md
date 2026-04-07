# pcgw-0yp

PeerCast 配信管理 Web アプリケーション。

## 必要なもの

- Go 1.22+
- Docker / Docker Compose (MySQL用)

## 起動方法

```bash
# MySQL起動
docker compose up -d

# ビルド&起動
go build -o pcgw-0yp .
./pcgw-0yp
```

設定ファイルを指定する場合:

```bash
./pcgw-0yp /path/to/config.toml
```

デフォルトではカレントディレクトリの `config.toml` を読み込みます。

## 設定

`config.toml` でサーバーポート、DB接続情報、Twitter OAuth を設定します。

```toml
[server]
port = 8080
session_secret = "change-me-to-a-random-string"

[db]
host = "127.0.0.1"
port = 3306
user = "pcgw"
passwd = "pcgw"
dbname = "pcgw"

[twitter]
client_id = ""
client_secret = ""
redirect_url = "http://localhost:8080/auth/twitter/callback"
```
