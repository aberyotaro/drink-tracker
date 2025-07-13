# 飲酒管理Slackアプリ

毎日の飲酒量を管理するSlackアプリ。飲み過ぎを防ぐため、簡単に記録できて適量を超えた場合にアラートを出す機能を実装。

## 開発環境セットアップ

### 1. 開発環境起動（ホットリロード有効）
```bash
# 開発環境起動
docker-compose -f docker-compose.dev.yml up -d --build

# ログ確認
docker-compose -f docker-compose.dev.yml logs -f

# 停止
docker-compose -f docker-compose.dev.yml down
```

### 2. 本番環境
```bash
# 本番環境起動
docker-compose up -d --build

# 停止
docker-compose down
```

## 環境設定

1. `.env`ファイルを作成:
```bash
cp config.example.env .env
```

2. Slack設定を`.env`に記入:
```env
SLACK_BOT_TOKEN=xoxb-your-bot-token-here
SLACK_SIGNING_SECRET=your-signing-secret-here
```

## API エンドポイント

- `GET /health` - ヘルスチェック
- `POST /slack/command` - Slackスラッシュコマンド受信

## 本番環境デプロイ

### 初回セットアップ

1. **fly.io CLI インストール・ログイン**
```bash
brew install flyctl
flyctl auth login
```

2. **アプリ作成**
```bash
flyctl apps create drink-tracker
```

3. **GitHub Secrets設定**
リポジトリの Settings > Secrets and variables > Actions で以下を追加：
- `FLY_API_TOKEN` - `flyctl auth token` で取得
- `SLACK_BOT_TOKEN` - SlackのBot User OAuth Token
- `SLACK_SIGNING_SECRET` - SlackのSigning Secret

### 自動デプロイ
- mainブランチにpushすると自動デプロイ
- GitHub Actions でビルド・デプロイ実行

### 手動デプロイ
```bash
flyctl deploy
```

### マイグレーション制御
- 環境変数 `AUTO_MIGRATE=true` で自動実行
- 本番では慎重に設定（現在は自動実行有効）

### ログ確認
```bash
flyctl logs -a drink-tracker
```

## 使用技術

- **Go 1.24** - バックエンド言語
- **Echo** - Webフレームワーク
- **SQLite** - データベース（modernc.org/sqlite - 純Go実装）
- **Bob ORM** - ORM
- **Air** - ホットリロード（開発環境）
- **Docker** - コンテナ化
- **fly.io** - 本番環境ホスティング
- **GitHub Actions** - CI/CD

## スラッシュコマンド（予定）

- `/drink beer` - ビール1本記録
- `/drink stats` - 今日の飲酒量表示
- `/drink history` - 過去3日分の記録表示