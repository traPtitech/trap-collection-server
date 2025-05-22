# traP Collection Server

このアプリは、サークル内で作られたゲームを管理・配信するシステムのサーバーサイドです。
Go 言語で書かれています。

## ディレクトリ構成

```txt
.
├── docker # Docker に関する設定
│   ├── base
│   ├── dev # 開発環境の設定
│   ├── production # 本番環境の設定
│   └── tbls # tbls による DB スキーマドキュメント生成用の設定
├── docs
│   ├── db_schema # DB スキーマドキュメント
│   ├── images
│   └── openapi # OpenAPI スキーマ
├── migrations # DB マイグレーションのSQLファイル
├── pkg
│   ├── context
│   ├── random # 暗号的に安全なランダム文字列生成
│   └── types # optionalな型
├── src
│   ├── auth # 認証関連
│   │   ├── mock
│   │   └── traQ
│   ├── cache # キャッシュ関連
│   │   ├── mock
│   │   └── ristretto
│   ├── config # 設定関連
│   │   ├── mock
│   │   └── v1 # 環境変数から読み込む実装
│   ├── domain # ドメインモデル
│   │   └── values # 値オブジェクト
│   ├── handler # HTTP ハンドラ
│   │   ├── common
│   │   └── v2 # ハンドラの実装
│   │       └── openapi # OpenAPI スキーマから生成されたコード
│   ├── repository # データの永続化。RDB を想定
│   │   ├── gorm2 # GORM を使った実装
│   │   │   ├── migrate # 古いマイグレーションのコード
│   │   │   └── schema # 新しいマイグレーションのコード
│   │   └── mock
│   ├── service # ビジネスロジック
│   │   ├── mock
│   │   ├── v1
│   │   └── v2
│   ├── storage # ストレージ関連。ファイルの保存先
│   │   ├── local # ローカルのファイルストレージ
│   │   ├── mock
│   │   ├── s3 # S3 ストレージ
│   │   └── swift # Swift ストレージ
│   └── wire # Dependency Injection
├── task # Taskfile の設定
└── testdata # テスト用に使うファイル
```

設計の詳細は [src/README.md ](src/README.md) も参照してください。

## 開発の進め方について

### テスト

ユニットテストを記述しています。ユニットテストでは、原則的に Table Driven Test を行います。
また、テスト時の assertion には、[testify](https://github.com/stretchr/testify) を使用しています。

### マイグレーション

マイグレーションには、[atlas](https://atlasgo.io/) を使用しています。
マイグレーションの詳細は、 [docs/migration.md](docs/migration.md) を参照してください。

### Linter について

golangci-lint を使用しています。
golangci-lint の設定は、[.golangci.yaml](.golangci.yaml) を参照してください。

## このアプリが扱う概念について

### game

ゲーム

### game file

ゲームのファイル。実体は zip ファイル。1つのゲームに複数のゲームファイルが存在する。
Windows用のexeファイルを含むもの、Mac OS用のアプリケーションを含むもの、JVM用のjarファイルを含むものがある。

### game video

ゲームの動画。1つのゲームに複数のゲーム動画が存在する。

### game image

ゲームの画像。1つのゲームに複数のゲーム画像が存在する。

### game version

ゲームのバージョン。1つのゲームに複数のバージョンが存在する。
1つのバージョンには、1つ以上のゲームファイルまたはURL、ゲームの動画、ゲームの画像が含まれる。

### edition

ランチャーのエディション。複数のgame versionをまとめたもの。
エディションは、ゲームのバージョンをまとめて配布するためのもの。
例えば、x年のコミケ用のエディション、y年の工大祭用のエディションなど。
以前は Launcher Version と呼ばれており、一部にその名前が残っている。

### game role

ゲームに対する権限。

- owner: ゲームの所有者。ゲームの編集、削除、game roleの変更、削除、game fileのアップロード、game videoのアップロード、game imageのアップロード、game versionの作成ができる。
- maintainer: ゲームの管理者。ゲームの編集、game roleの変更、game fileのアップロード、game videoのアップロード、game imageのアップロード、game versionの作成ができる。

### admin

traP Collection の管理者。
edition を作成できるほか、全てのゲームについて操作を行うことができる。

### seat

座席。工大祭などで座席の管理をするのに使用する。
ランチャーから今の座席の状態が送信され、空いている座席を確認できる。
