# trap-collection-server

[![codecov](https://codecov.io/gh/traPtitech/trap-collection-server/branch/main/graph/badge.svg)](https://codecov.io/gh/traPtitech/trap-collection-server)
[![](https://github.com/traPtitech/trap-collection-server/workflows/Release/badge.svg)](https://github.com/traPtitech/trap-collection-server/actions)
[![OpenAPI(v1)](https://img.shields.io/badge/OpenAPI(v1)-docs-brightgreen)](https://apis.trap.jp/?urls.primaryName=traP%20Collection%20v1%20API)
[![OpenAPI(v2)](https://img.shields.io/badge/OpenAPI(v2)-docs-brightgreen)](https://apis.trap.jp/?urls.primaryName=traP%20Collection%20v2%20API)
[![go report](https://goreportcard.com/badge/traPtitech/trap-collection-server)](https://goreportcard.com/report/traPtitech/trap-collection-server)
[![Ask DeepWiki](https://deepwiki.com/badge.svg)](https://deepwiki.com/traPtitech/trap-collection-server)

traPのゲームランチャーtraP Collectionのサーバーサイドです。

## 開発に必要なツール

- Go (最新のもの)
- Docker
- [Task](https://taskfile.dev/)
  - タスクランナー
- [Atlas](https://atlasgo.io/)
  - DBマイグレーション用のツール
  - マイグレーションの詳細は [migration.md](docs/migration.md) を参照してください。
- [golangci-lint](https://golangci-lint.run/)
  - 静的解析用のツール

必要なものをインストールしたら、次に以下のコマンドを実行すると、Goの依存関係がインストールされ、コードが生成されます。

```bash
task
```

`src/service/mock` ディレクトリなどにファイルが生成されていればよいです。

## 開発環境の起動

まず、[traQ BOT Console](https://bot-console.trap.jp/docs/client/create)に従い、
traQのOAuthクライアントを作成します。リダイレクト先URLは`http://localhost:8080/callback`、スコープは「読み取り」、に設定してください。

次に、`docker/dev/.env`ファイルに以下のように書きます。

```.env
CLIENT_ID={{traQのClientのClientID}}
CLIENT_SECRET={{traQのClientのClientSecret}}
```

最後に以下のコマンドを実行することで開発環境がポート3000番で起動します。
また、管理画面のWeb UIをhttp://localhost:8080 で、[Adminer](https://www.adminer.org/)をhttp://localhost:8081 で開けるようになります。

```bash
task dev
```

### DBのデータ削除

以下のコマンドを実行することでDBのデータを削除できます。

```bash
task clean:db
```

## テストの実行

GitHub Actionsで走っているのと同様のテストを以下のコマンドで実行できます。
ログの停止からtestのログが流れ始めるまで20秒程度時間が空く点に注意してください。
DB に関わるテストは、Dockerが必要です。

```bash
task test
```

`task test`では全てのテストが実行されます。開発中に関数ごとのテストを実行したい場合は、コマンドでテスト対象の関数を指定して実行することもできますが、エディタの機能を使うとよいです。
例えば VSCode では、エディタのテスト関数の上にある「run test」ボタンやコマンドパレットなどから実行できます。

## DBスキーマの再生成

以下のコマンドでDBスキーマのドキュメント(`docs/db_schema`)を再生成できます。

```bash
task tbls
```

## マイグレーションについて

[docs/migration.md](docs/migration.md)を参照してください。
