# PeerCast API (peercast-mi JSON-RPC)

pcgw-0yp は peercast-mi の JSON-RPC 2.0 API を `internal/peercast/client.go` 経由で呼び出す。

## 接続方法

- エンドポイント: `POST http://{hostname}:{port}/api/1`
- localhost からの接続: 認証不要
- リモートからの接続: HTTP Basic Auth 必要 (`auth_id` / `passwd`)
- ヘッダー: `Content-Type: application/json`, `X-Requested-With: XMLHttpRequest`
- タイムアウト: 5秒

## 使用する API メソッド

### システム情報

| メソッド | パラメータ | 戻り値 | 用途 |
|----------|-----------|--------|------|
| `getVersionInfo` | なし | `{agentName}` | サーバント情報取得 |
| `getSettings` | なし | `{serverPort, rtmpPort}` | ポート情報取得 |
| `getYellowPages` | なし | `[{yellowPageId, name, uri, announceUri, channelCount}]` | YP一覧取得 |

### チャンネル管理

| メソッド | パラメータ | 戻り値 | 用途 |
|----------|-----------|--------|------|
| `getChannels` | なし | `[{channelId, status, info, track}]` | 全チャンネル一覧 |
| `getChannelInfo` | `[channelId]` | `{info, track}` | チャンネル情報取得 |
| `getChannelStatus` | `[channelId]` | `{status, source, uptime, localRelays, localDirects, ...}` | ステータス取得 |
| `setChannelInfo` | `[channelId, info, track]` | `null` | メタデータ更新 |
| `stopChannel` | `[channelId]` | `null` | 配信停止 |
| `bumpChannel` | `[channelId]` | `null` | YPへ即時通知 |

### 接続管理

| メソッド | パラメータ | 戻り値 | 用途 |
|----------|-----------|--------|------|
| `getChannelConnections` | `[channelId]` | `[{connectionId, type, status, sendRate, ...}]` | 接続一覧 |
| `stopChannelConnection` | `[channelId, connectionId]` | `boolean` | relay接続切断 |
| `getChannelRelayTree` | `[channelId]` | `[{sessionId, address, port, children, ...}]` | リレーツリー |

### ストリームキー管理 (peercast-mi 固有)

| メソッド | パラメータ | 戻り値 | 用途 |
|----------|-----------|--------|------|
| `issueStreamKey` | `[accountName, streamKey]` | `null` | ストリームキー発行 |
| `revokeStreamKey` | `[accountName]` | `null` | ストリームキー失効 |
| `listStreamKeys` | なし | `[{accountName, streamKey}]` | 一覧取得 |

### 配信開始 (peercast-mi 固有)

| メソッド | パラメータ | 戻り値 | 用途 |
|----------|-----------|--------|------|
| `broadcastChannel` | `[{streamKey, info, track}]` | `{channelId}` | 配信チャンネル作成 |

## 配信フロー

```
1. issueStreamKey("user_123", "sk_xxxx...")
   → peercast-mi がストリームキーを登録

2. broadcastChannel({
     streamKey: "sk_xxxx...",
     info: { name: "配信名", genre: "ゲーム", ... },
     track: { creator: "ユーザー名" }
   })
   → peercast-mi がチャンネルを作成し channelId を返す
   → RTMP: rtmp://{hostname}:1935/live/sk_xxxx... で待ち受け開始

3. ユーザーが OBS 等から RTMP push

4. setChannelInfo(channelId, info, track)
   → メタデータ変更時

5. stopChannel(channelId)
   → 配信停止

6. revokeStreamKey("user_123")
   → ストリームキー無効化
```

## 元の pcgw (PeerCast YT) との違い

| pcgw (PeerCast YT) | pcgw-0yp (peercast-mi) |
|---------------------|------------------------|
| `fetch(url, name, ...)` でチャンネル作成 | `issueStreamKey` + `broadcastChannel` |
| PeerCast が外部 URL から PULL | ユーザーが RTMP で PUSH |
| WMV/MKV/FLV 対応 | FLV のみ |
| `restartChannelConnection` 対応 | 未対応 |
| YP はアプリ内ハードコード | `getYellowPages` で動的取得 |

## エラーハンドリング

- `peercast.Unavailable` エラー: peercast-mi に接続できない場合
- JSON-RPC エラーコード:
  - `-32700`: パースエラー
  - `-32601`: メソッド未対応
  - `-32602`: パラメータ不正
  - `-32603`: 内部エラー
