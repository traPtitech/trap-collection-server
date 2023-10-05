# trap-collection-server
[![codecov](https://codecov.io/gh/traPtitech/trap-collection-server/branch/main/graph/badge.svg)](https://codecov.io/gh/traPtitech/trap-collection-server)
[![](https://github.com/traPtitech/trap-collection-server/workflows/Release/badge.svg)](https://github.com/traPtitech/trap-collection-server/actions)
[![OpenAPI(v1)](https://img.shields.io/badge/OpenAPI(v1)-docs-brightgreen)](https://apis.trap.jp/?urls.primaryName=traP%20Collection%20v1%20API)
[![OpenAPI(v2)](https://img.shields.io/badge/OpenAPI(v2)-docs-brightgreen)](https://apis.trap.jp/?urls.primaryName=traP%20Collection%20v2%20API)
[![go report](https://goreportcard.com/badge/traPtitech/trap-collection-server)](https://goreportcard.com/report/traPtitech/trap-collection-server)

traPのゲームランチャーtraP Collectionのサーバーサイドです。

## 準備
この後の手順では
- Go
- Docker

が必要となります。

[Task](https://taskfile.dev/)をタスクランナーとして使用しているので、
Goをinstallした上で以下のコマンドを実行してinstallしてください。
```bash
go install github.com/go-task/task/v3/cmd/task
```

次に以下のコマンドを実行することで環境構築が完了します。
```bash
task
```

## 開発環境の起動
まず、[traQ BOT Console](https://bot-console.trap.jp/docs/client/create)に従い、
traQのOAuthクライアントを作成します。

次に、`docker/dev/.env`ファイルに以下のように書きます。
```
CLIENT_ID={{traQのClientのClientID}}
CLIENT_SECRET={{traQのClientのClientSecret}}
```

最後に以下のコマンドを実行することで開発環境がポート3000番で起動します。
また、Web UIをhttp://localhost:8080 で、[Adminer](https://www.adminer.org/)をhttp://localhost:8081 で開けるようになります。
```bash
task dev
```

### DBのデータ削除
以下のコマンドを実行することでDBのデータを削除できます。
```bash
task clean:db
```

### v1のAPIのみで起動
以下のコマンドを実行することで、現在の本番環境と同様のv1のAPIのみが有効な状態で起動できます。
```bash
task dev:v1
```

## テストの実行
GitHub Actionsで走っているのと同様のテストを以下のコマンドで実行できます。
ログの停止からtestのログが流れ始めるまで20秒程度時間が空く点に注意してください。
```bash
task test
```

## DBスキーマの再生成
以下のコマンドでDBスキーマのドキュメント(`docs/db_schema`)を再生成できます。
```bash
task tbls
```
