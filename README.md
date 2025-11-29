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

1. Goプログラムでイベント情報を収集
2. Reactアプリをビルド
3. GitHub Pagesにデプロイ

手動でワークフローを実行する場合は、GitHubリポジトリの「Actions」タブから「Build and Deploy」ワークフローを選択し、「Run workflow」をクリックしてください。

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
