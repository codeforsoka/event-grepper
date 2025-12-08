# event-grepper

草加市で開催されるイベント情報を自動収集し、Webサイトに掲載するプロジェクトです。

## 機能

- **複数ソースからの情報収集**
  - 広報そうか（自動で年度URLを検出）
  - 草加市イベント情報ページ
  - 松原団地記念公園のイベント情報

- **Webサイト機能**
  - モダンなレスポンシブデザイン
  - リスト表示とカレンダー表示の切り替え
  - カテゴリー別フィルタリング
  - イベント詳細情報の表示（開催場所、日時、内容）

- **自動更新**
  - GitHub Actionsによる毎日定期実行
  - GitHub Pagesへの自動デプロイ

## 開発環境構築

### 必要な環境
- Go 1.22.1以上
- Node.js 20以上
- Ollama（イベント情報の自動抽出に使用）

### Ollamaのセットアップ

このプロジェクトでは、LLMを使用してWebページからイベント情報を抽出します。

#### 方法1: Docker Compose（推奨）

```bash
# Ollamaをバックグラウンドで起動
docker-compose up -d

# 初回のみ：モデルをダウンロード（llama3.1:8bはより大きく高精度）
docker exec event-grepper-ollama ollama pull llama3.1:8b

# 確認
docker exec event-grepper-ollama ollama list
```

#### 方法2: ローカルインストール

1. Ollamaをインストール:
```bash
# macOS
brew install ollama

# Linux
curl -fsSL https://ollama.ai/install.sh | sh
```

2. Ollamaを起動:
```bash
ollama serve
```

3. 高精度モデルをダウンロード:
```bash
# llama3.1:8b（約5GB、より高精度な抽出が可能）
ollama pull llama3.1:8b

# または軽量版（約2GB、高速だが精度は低い）
ollama pull llama3.2
```

**注意**: Ollamaが起動していない場合、プログラムは終了します。

### main.goのビルドと実行

```bash
# Goプログラムのビルド
$ go build

# イベント情報の収集を実行
# event-grepper-app/src配下にpark.jsonが生成されます
$ ./event-grepper
```

### React Appのローカル起動

```bash
$ cd event-grepper-app

# 依存パッケージのインストール（初回のみ）
$ npm install

# 開発サーバーの起動
$ npm start
```

ブラウザで http://localhost:3000 を開くと、イベント情報ページが表示されます。

## デプロイ

GitHub Actionsが自動的に以下を実行します：

1. **Ollamaコンテナを起動** - LLMモデル（llama3.2）をダウンロード
2. **Goプログラムでイベント情報を収集** - LLMを使用してWebページから抽出
3. **Reactアプリをビルド** - 収集したデータを含めてビルド
4. **GitHub Pagesにデプロイ** - 静的サイトを公開

### GitHub Actionsの実行

- **自動実行**: 毎日9時（UTC）に自動実行
- **手動実行**: GitHubリポジトリの「Actions」タブから「Build and Deploy」ワークフローを選択し、「Run workflow」をクリック
- **プッシュ時**: mainブランチへのプッシュで自動実行

### 注意事項

GitHub Actionsでは、Ollamaコンテナの起動とモデルのダウンロードに約5-10分かかります。初回実行時は特に時間がかかる場合があります。

## プロジェクト構成

```
event-grepper/
├── main.go                 # イベント情報収集プログラム
├── .github/
│   └── workflows/
│       └── build.yml       # GitHub Actionsワークフロー
└── event-grepper-app/      # Reactアプリケーション
    ├── src/
    │   ├── App.tsx         # メインコンポーネント
    │   ├── App.css         # スタイル定義
    │   ├── EventCalendar.tsx  # カレンダー表示コンポーネント
    │   └── park.json       # 収集されたイベントデータ
    └── package.json
```

## ライセンス

このプロジェクトは草加市の公開情報を収集・表示するものです。
