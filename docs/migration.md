# DB のマイグレーションについて

[Atlas](https://atlasgo.io/) を使っています。このツールでは、DBスキーマを宣言的に管理することができます。

## Atlas のインストール

手元で開発を行うためにはAtlasをインストールしておく必要があります。
公式ドキュメントのインストール手順に従ってください。

https://atlasgo.io/getting-started#installation

## マイグレーションに関係するファイル・ディレクトリ

- [`migrations/*.sql`](../migrations/): マイグレーションファイルそのものです。タイムスタンプがついており、この順番にマイグレーションが実行されます。
- [`migrations/atlas.sum`](../migrations/atlas.sum): マイグレーションファイルのハッシュ値が記録されています。基本的には自動で更新されますが、マイグレーションファイルを変更した場合は、`atlas migrate hash`を実行してこのファイルを更新する必要があります。
- [`atlash.hcl`](../atlash.hcl): Atlasの設定ファイルです。DBの接続情報やマイグレーションの設定が記述されています。
- [`src/repository/gorm2/schema/*.go`](../src/repository/gorm2/schema/): DBスキーマを定義するGoの構造体が記述されています。

## マイグレーションの書き方

Atlas では、Declarative Workflows と Versioned Workflows の2つのワークフローをサポートしています。traP Collection では、Versioned Workflows を採用しています。ドキュメントなどで調べる際は Declarative Workflows の方を見ないように注意してください。

スキーマは Gorm の構造体を使って定義します。DBスキーマを変更したい場合は、以下の手順に従ってください。

### スキーマの定義

[`src/repository/gorm2/schema`](../src/repository/gorm2/schema/) に書かれている構造体を変更します。既存のテーブルを書き換えたい場合は、既存の構造体を変更します。新しいテーブルを追加したい場合は、新しい構造体を追加します。`gorm`タグがついている構造体がスキーマとして認識されます。

### マイグレーションファイルの生成

マイグレーションファイルを生成するには、以下のコマンドを実行します。

```bash
task migrate:new -- <migration_name>
```

`<migration_name>` には、マイグレーションの名前を指定します。例えば、`add_users_table` のようにします。

このコマンドを実行すると、[`migrations`](../migrations/)ディレクトリにマイグレーションのSQLファイルが生成されます。

生成されたファイルに対して手動で変更を加えることもできます。例えば、データをINSERTしたいときなどです。
この場合は、SQLファイルを変更した後、`atlas.sum`ファイルを更新するために以下のコマンドを実行する必要があります。

```bash
atlas migrate hash
```

### マイグレーションの実行

マイグレーションの実行はアプリケーションに組み込まれているので、特にコマンドを実行する必要はありません。
アプリケーションを起動すると、自動的にマイグレーションが実行されます。

マイグレーションの実行は[`src/repository/gorm2/db.go`](../src/repository/gorm2/db.go)で記述されています。具体的には、`migrations`ディレクトリの SQL ファイルが`embed`パッケージによってバイナリに埋め込まれており、アプリケーション起動時に Atlas が提供しているライブラリを使って実行されます。

## マイグレーションの削除

開発中に自分が書いたマイグレーションを取りやめて書き直したいときがあると思います。その場合は、以下のコマンドを実行してください。

```bash
task migrate:down
```

このコマンドを実行すると、DBに適用されている最新の1つのマイグレーションが元に戻され、マイグレーションファイルが削除されます。
マイグレーションを元に戻したいが、マイグレーションファイルは残しておきたい場合は、以下のコマンドを実行してください。

```bash
task migrate:down:down-only
```

## マイグレーションの Lint

Atlas の Lint 機能を使って、マイグレーションファイルのチェックを行うことができます。
ただし、通常の Atlas CLI の free plan では、Lint 機能を使うことができません。必要であれば community edition をインストールして使ってください。
GitHub Actions の CI では community edition を使って Lint を実行しています。
以下のコマンドを実行してください。

```bash
altas migrate lint --config file://atlas.ci.hcl --env local
```

このコマンドでは、main ブランチと比較してマイグレーションが正しいかをチェックします。
error が出た場合は、マイグレーションファイルを修正してください。warning の場合は、修正しなくても問題ありませんが、修正した方が良い場合があります。必要に応じて修正してください。

Lint 結果の出力の意味がよく分からない場合は、公式ドキュメントを確認してください。

https://atlasgo.io/lint/analyzers

また、Lint は GitHub Actions でも実行しています。手元で Lint が通っても、手元の main ブランチが最新でなかったり Atlas のバージョンが違ったりすると CI でエラーになることがあります。

## 以前のマイグレーションからの移行

Atlas 導入以前は、[gormigrate](https://github.com/go-gormigrate/gormigrate) を使っていました。このプログラムは [`src/repository/gorm2/migrate`](../src/repository/gorm2/migrate/) にあります。
[`gorm2`](../src/repository/gorm2/)パッケージのソースコードには、まだ `migrate` パッケージの構造体を参照している部分があります。この構造体に対しては、新しい`schema` パッケージの構造体への型エイリアスが貼ってあり、実体は新しい `schema` パッケージの構造体になっています。`migrate` パッケージを参照している部分は徐々に `schema` パッケージの構造体に置き換えていく予定です。`migrate` パッケージは、Atlas に完全に移行した後も残しておく予定ですが、今後の新しいプログラムでは `schema` パッケージを使ってください。
