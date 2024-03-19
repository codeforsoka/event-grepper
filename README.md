# event-grepper
草加市で開催されるイベントを収集しまとめページに掲載します

# 開発環境構築
## main.goのビルド

```
$ go build

// ビルドされたバイナリファイルを実行すると、`event-grepper-app/src`配下に収集したイベント情報がpark.jsonとして生成されます。
$ ./event-grepper
```

## React Appのローカル起動

```
$ cd event-grepper-app

$ yarn start
```
