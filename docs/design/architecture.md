# アーキテクチャ

## 概要

pcgw-0yp は PeerCast 配信を管理する Web アプリケーション。
Ruby/Sinatra 製の pcgw を Go に移植したもので、バックエンドに peercast-mi を使用する。

## 技術スタック

| 項目 | 技術 |
|------|------|
| 言語 | Go 1.22+ |
| ルーター | go-chi/chi v5 |
| テンプレート | html/template |
| DB | SQLite3 (github.com/mattn/go-sqlite3) |
| ORM | なし (database/sql 直接) |
| 認証 | Twitter OAuth2 (golang.org/x/oauth2) |
| セッション | gorilla/sessions (cookie-based, 30日有効) |
| CSS | Bulma 0.9.4 (CDN) |
| PeerCast通信 | JSON-RPC 2.0 (自前クライアント) |

## ディレクトリ構成

```
pcgw-0yp/
├── main.go                     # エントリーポイント
├── config.toml                 # 設定ファイル
├── go.mod / go.sum
├── internal/
│   ├── config/config.go        # TOML設定読み込み
│   ├── db/
│   │   ├── db.go               # DB接続 (WALモード, FK有効)
│   │   └── schema.go           # DDL (自動マイグレーション)
│   ├── model/                  # データアクセス層
│   │   ├── user.go
│   │   ├── channel.go
│   │   ├── channel_info.go
│   │   ├── servent.go
│   │   ├── source.go
│   │   └── notice.go
│   ├── peercast/client.go      # peercast-mi JSON-RPCクライアント
│   ├── handler/                # HTTPハンドラー
│   │   ├── handler.go          # 共通基盤 (テンプレート, flash, render)
│   │   ├── auth.go             # Twitter OAuth2
│   │   ├── home.go             # ダッシュボード
│   │   ├── broadcast.go        # 配信作成
│   │   ├── channel.go          # チャンネル管理
│   │   ├── profile.go          # プロフィール
│   │   ├── program.go          # 配信履歴
│   │   ├── stats.go            # 統計
│   │   ├── servent.go          # サーバント管理 (admin)
│   │   ├── admin.go            # ユーザー管理 (admin)
│   │   ├── notice.go           # お知らせ
│   │   ├── source.go           # ソース管理
│   │   ├── api.go              # JSON API
│   │   └── cleanup.go          # 定期クリーンアップ
│   ├── middleware/
│   │   ├── session.go          # セッション管理
│   │   ├── auth.go             # 認証チェック + ユーザー注入
│   │   └── csrf.go             # CSRFトークン検証
│   ├── server/router.go        # ルーティング定義
│   └── view/helpers.go         # テンプレート関数
├── templates/                  # HTMLテンプレート
│   ├── partials/               # 共通パーツ (navbar, flash)
│   └── *.html                  # 各画面
├── public/                     # 静的ファイル
│   ├── css/style.css
│   └── js/
├── db/pcgw.db                  # SQLiteデータベース (実行時生成)
└── docs/                       # ドキュメント
```

## リクエスト処理フロー

```
HTTP Request
  │
  ├─ chi/middleware.Logger      ログ出力
  ├─ chi/middleware.Recoverer   パニック回復
  ├─ middleware.Session         セッション読み込み → context に格納
  ├─ middleware.CSRF            POST時にトークン検証
  ├─ middleware.Auth            ユーザー認証チェック
  │   ├─ 公開パス → スキップ
  │   ├─ 未ログイン → /login にリダイレクト
  │   ├─ 凍結ユーザー → 403
  │   └─ 認証OK → User を context に格納
  │
  └─ handler.*                  ビジネスロジック
      ├─ model.*                DB操作
      ├─ peercast.Client        peercast-mi API呼び出し
      └─ render()               テンプレートレンダリング
```

## データフロー: 配信作成

```
ユーザー (ブラウザ)
  │
  │  POST /broadcast
  │  (channel_name, genre, desc, comment, url, bitrate, yp, servent_id)
  │
  ▼
handler.Broadcast()
  │
  ├─ model.RequestServentWithVacancy()   空きサーバント検索
  │
  ├─ peercast.Client.IssueStreamKey()    ストリームキー発行
  │   → peercast-mi: issueStreamKey(accountName, streamKey)
  │
  ├─ peercast.Client.BroadcastChannel()  チャンネル作成
  │   → peercast-mi: broadcastChannel({streamKey, info, track})
  │   ← channelId
  │
  ├─ model.CreateChannel()               DBにチャンネルレコード作成
  ├─ model.CreateChannelInfo()           DBに配信メタデータ作成
  │
  └─ Redirect → /channels/{id}
                    │
                    ▼
              チャンネル画面に RTMP push URL を表示
              ユーザーが OBS 等から rtmp://host:1935/live/{streamKey} にプッシュ
```

## 定期クリーンアップ

`handler.StartCleanup()` がゴルーチンで30秒間隔で実行:

1. 全チャンネルの `getChannelStatus` を確認
2. 受信中 → `last_active_at` を更新
3. 15分以上非アクティブ → `stopChannel` + DB削除
4. 作成後15分でデータ未受信 → 同上
