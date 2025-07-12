# 飲酒管理Slackアプリ

## プロジェクト概要
毎日の飲酒量を管理するSlackアプリを開発する。飲み過ぎを防ぐため、簡単に記録できて適量を超えた場合にアラートを出す機能を実装する。

## 技術スタック

### バックエンド
- **言語**: Go (開発者の得意言語)
- **フレームワーク**: Echo
- **データベース**: SQLite (初期) → PostgreSQL (本格運用時)
- **ORM**: Bob (SQLボイラープレート生成、型安全性重視)
- **Slack SDK**: `github.com/slack-go/slack`

### アーキテクチャ
```
Slack App (UI)
    ↓ (HTTP Request/Webhook)
Go API Server (Echo)
    ↓
SQLite Database (Bob ORM)
```

## 機能要件

### 1. 飲酒記録機能
- **スラッシュコマンド**: `/drink [種類] [量]`
  - `/drink beer` → ビール1本（350ml）
  - `/drink beer 500ml` → ビール500ml
  - `/drink wine 150ml` → ワイン150ml
  - `/drink sake 1go` → 日本酒1合

### 2. 過去データ閲覧機能
- `/drink stats` → 今日の飲酒量
- `/drink stats week` → 今週の飲酒量
- `/drink stats month` → 今月の飲酒量
- `/drink history` → 過去3日分の記録

### 3. 編集・削除機能
- `/drink edit` → 今日の記録を修正
- `/drink delete last` → 最後の記録を削除

### 4. アラート機能
- 適量超過時の警告メッセージ
- 日次サマリー通知（オプション）

## データベース設計

### テーブル構成
```sql
-- ユーザーテーブル
users (
    id INTEGER PRIMARY KEY,
    slack_user_id TEXT UNIQUE,
    slack_team_id TEXT,
    daily_limit_ml INTEGER DEFAULT 40000,  -- 適量設定（ml）
    created_at DATETIME,
    updated_at DATETIME
)

-- 飲酒記録テーブル
drink_records (
    id INTEGER PRIMARY KEY,
    user_id INTEGER,
    drink_type TEXT,           -- beer, wine, sake, etc.
    amount_ml INTEGER,         -- 飲酒量（ml）
    alcohol_percentage REAL,   -- アルコール度数
    recorded_at DATETIME,
    created_at DATETIME,
    FOREIGN KEY (user_id) REFERENCES users(id)
)

-- 飲み物マスタテーブル
drink_types (
    id INTEGER PRIMARY KEY,
    name TEXT,                 -- beer, wine, sake
    typical_amount_ml INTEGER, -- 標準的な量
    alcohol_percentage REAL    -- 標準的なアルコール度数
)
```

## API仕様

### Slack認証
- **リクエスト検証**: Signing Secretを使用したHMAC検証
- **Botトークン**: `xoxb-...` でSlackへのレスポンス送信

### エンドポイント
- `POST /slack/command` - スラッシュコマンド受信
- `POST /slack/events` - Slackイベント受信（将来拡張用）

## 適量の基準
- **適量**: 1日あたり純アルコール20g以下
- **計算式**: `飲酒量(ml) × アルコール度数(%) × 0.8 = 純アルコール量(g)`
- **例**: ビール350ml (5%) = 350 × 0.05 × 0.8 = 14g

## レスポンス例

### 記録成功時
```
✅ ビール350mlを記録しました
📊 今日の飲酒量: 350ml (純アルコール14g)
😊 適量内です
```

### 適量超過時
```
⚠️ ビール700mlを記録しました
📊 今日の飲酒量: 700ml (純アルコール28g)
🚨 適量を超えています。お気をつけください。
```

### 統計表示例
```
📊 今週の飲酒量 (7/6-7/12)
月: ビール350ml 🍺
火: お休み 😊
水: ワイン200ml + ビール350ml 🍷🍺
木: ビール350ml 🍺
金: 日本酒1合 + ビール350ml 🍶🍺
土: ビール700ml 🍺🍺
日: ビール350ml 🍺

週間合計: 2,350ml (純アルコール188g)
週間平均: 27g/日
```

## 開発・リポジトリ管理

### GitHub Public Repository
- オープンソースとして公開開発
- 健康管理系アプリとして他の開発者の参考になる
- 将来的な一般公開時の信頼性向上

### セキュリティ管理
- **機密情報は絶対にコミットしない**:
  - Bot User OAuth Token (`xoxb-...`)
  - Signing Secret
- **環境変数管理**:
  - `.env` ファイルで管理
  - `.gitignore` に `.env` を追加
  - `config.example.env` でサンプル設定を提供

### 推奨ディレクトリ構成
```
drink-tracker/
├── cmd/
│   └── server/
│       └── main.go
├── internal/
│   ├── handlers/
│   ├── models/
│   └── services/
├── migrations/
├── .env
├── .gitignore
├── config.example.env
├── go.mod
├── go.sum
└── README.md
```

## 開発フェーズ
1. Go環境構築
2. Slack App設定（Slash Command）
3. 基本的な記録機能 (`/drink beer`)
4. SQLiteデータベース設定
5. 適量チェック機能

### Phase 2: 拡張機能
1. 統計表示機能
2. 編集・削除機能
3. より詳細な飲み物種別対応
4. エラーハンドリング改善

### Phase 3: 運用改善
1. PostgreSQL移行
2. ログ機能
3. パフォーマンス最適化
4. 一般公開準備

## Slack App設定済み情報
- ✅ Slack App作成完了
- ✅ Bot設定完了
- ✅ ワークスペースにインストール完了
- ✅ Bot User OAuth Token取得済み
- ✅ Signing Secret取得済み

## 次のステップ
1. `/drink` Slash Command設定
2. Go プロジェクト初期化
3. 基本的なHTTPサーバー実装
4. Slack認証実装
5. データベース設計・実装

## 将来の拡張案
- Webアプリとの連携
- より詳細な統計・グラフ表示
- 他の健康管理機能との統合
- 複数ワークスペース対応