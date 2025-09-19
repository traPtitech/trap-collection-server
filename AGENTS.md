# traP Collection Server / Agents Guide

## 1. アーキテクチャ (Big Picture)

- レイヤ順: `handler (REST)` → `service` → `repository` / `storage` / `auth` / `cache` / `domain`。
- 依存注入: `google/wire` (`src/wire/wire.go` → 生成物 `wire_gen.go`)。依存変更時は `task wire` か包括的に `task generate`。
- API: OpenAPI v2 (`docs/openapi/v2.yaml`) → `oapi-codegen` 生成コード `src/handler/v2/openapi`。実装は `src/handler/v2/*` に手書き。
- DB: MySQL (GORM2) リポジトリ具象は `src/repository/gorm2`。Atlas による SQL マイグレーション (`migrations/*.sql`, `atlas.hcl`)。
- ストレージ: 設定により `swift` / `local` / `s3` を `storageSwitch()` で選択 (`wire_gen.go`)。
- キャッシュ: Ristretto (`src/cache/ristretto`)。ユーザ/Seat など service + cache の 2 層。
- ドメイン値: `src/domain/values` (必ずコンストラクタで作る, 直接 struct リテラル禁止)。

## 2. よくある開発フロー (Task Runner)

| 目的                   | コマンド                   | 備考                                             |
| ---------------------- | -------------------------- | ------------------------------------------------ |
| 依存DL +コード 生成    | `task`                     | `download` + `generate` (wire + openapi + mocks) |
| 開発環境起動(v2)       | `task dev`                 | 事前に OAuth の `.env` の設定が必要              |
| テスト                 | `task test`                | Docker 必須 / Table Driven + testify             |
| Lint(diff対象)         | `task lint`                | 古いコードは無理に直さない                       |
| 自動修正試行           | `task lint:fix`            | 可能な範囲で整形                                 |
| 新規マイグレーション   | `task migrate:new -- name` | 空テンプレ生成                                   |
| マイグレーション Lint  | `task migrate:lint`        | 順序/競合検証                                    |
| DB スキーマ Doc の生成 | `task tbls`                | `docs/db_schema` 更新                            |
| 全停止                 | `task down`                | dev + tbls 停止                                  |
| DB データ削除          | `task clean:db`            | 破壊的操作注意                                   |

## 3. 新規エンドポイント追加手順 (標準パターン)

1. `docs/openapi/v2.yaml` にスキーマ/paths/definitions 追加。
2. `task openapi` (または `task generate`) でコード生成。
3. `src/service/v2` にビジネスロジック追加 (権限・集約操作ここに集約)。
4. `src/repository/gorm2` など service から操作するものの実装。
5. `src/handler/v2` に HTTP ハンドラーの実装。
6. 依存が増えたら `src/wire/wire.go` の Set に interface 追加 → `task wire`。
7. Table Driven Test を `*_test.go` に追加。

## 4. 生成/編集ポリシ (Do / Don't)

| Do                                                    | Don't                                             |
| ----------------------------------------------------- | ------------------------------------------------- |
| Interface 契約を先に考え実装は従う                    | 実装都合で interface を破壊的変更                 |
| OpenAPI / SQL / wire Set を更新し再生成               | 生成物 (`wire_gen.go`, `openapi` code) を直接編集 |
| 値オブジェクトはコンストラクタ利用                    | struct リテラル直書き                             |
| Repository では `fmt.Errorf("...: %w", err)` でラップ | 下位エラーそのまま丸投げ                          |

## 5. 権限/Visibility の要点

- Visibility(`public/limited/private`) はゲーム取得系でフィルタ必須。既存 handler/service を参照し同一条件を使い回す。
- Game Role: `owner` は削除/ロール管理まで、`maintainer` は編集/ファイル追加可。
- Admin: 全ゲーム操作 + Edition 作成権限。
- 権限は handler の checker で調べる。

## 6. テストパターン (例)

Table Driven Test を採用する。

```go
testCases := []struct{ 
    name string
    input X
    wantErr bool
  }{
	  {name: "ok", input: X{...}},
  	{name: "invalid id", input: X{...}, wantErr: true},
  }
for _, testCase := range testCases { t.Run(testCase.name, func(t *testing.T) { /* require/assert */ }) }
```

または

```go
testCases := map[string]struct{ 
    input X
    wantErr bool 
  }{
	  "ok": {input: X{...}},
	  "invalid id": {input: X{...}, wantErr: true},
  }
for name, testCase := range testCases { t.Run(name, func(t *testing.T) { /* require/assert */ }) }
```

## 7. 変更の判断基準

| 判断観点           | 質問例                      | 維持方針                                         |
| ------------------ | --------------------------- | ------------------------------------------------ |
| 新規ロジックの所在 | ビジネスルールか I/O 変換か | ルール=service, I/O=handler                      |
| 依存の粒度         | DB/sql 操作が複数集約か     | 再利用/権限境界は service で統合                 |
| トランザクション   | どこで開始/commit すべきか  | service で開始し repository interface に Tx 注入 |

## 8. よくある落とし穴

- wire 再生成忘れ (`task wire` or `task generate`) → ビルドエラー。
- OpenAPI スキーマ未更新で handler 追加 → 生成コード不整合。
- 生成コードを直接編集 → 次回再生成で消滅。
- Visibility / 権限チェック抜け → 認可バグ。既存 service メソッド再利用を優先。
- 古いコードへ過度な lint 適用 → 差分が膨らむ。変更範囲最小化。

## 9. ディレクトリ構成 (詳細)

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

設計の詳細は `src/README.md` も参照してください。

## このアプリが扱う概念について

### game

ゲーム。
タイトルと説明が紐づけられている。

### game file

ゲームを動かすためのファイル。実体は zip ファイル。1つのゲームに複数のゲームファイルが存在する。
Windows 用の exe ファイルを含むもの、Mac OS 用のアプリケーションを含むもの、JVM 用の jar ファイルを含むものがある。

### game video

ゲームの動画。1つの game に複数の game video が存在する。

### game image

ゲームの画像。1つの game に複数の game image が存在する。

### game version

ゲームのバージョン。1つの game に複数の game version が存在する。
1つのバージョンには、1つ以上の game file または URL、1つの game video、1つの game image が含まれる。

### edition

ランチャーのエディション。複数の game version をまとめたもの。
エディションは、 game version をまとめて配布するためのもの。
例えば、x年のコミケ用のエディション、y年の工大祭用のエディションなど。
以前は Launcher Version と呼ばれており、一部にその名前が残っている。

### game role

game に対する権限。

- owner: game の所有者。game の編集、削除、game roleの変更、削除、game fileのアップロード、game videoのアップロード、game imageのアップロード、game versionの作成ができる。
- maintainer: game の管理者。game の編集、game roleの変更、game fileのアップロード、game videoのアップロード、game imageのアップロード、game versionの作成ができる。

### admin

traP Collection の管理者。
edition を作成できるほか、全ての game について操作を行うことができる。

### seat

座席。工大祭などで座席の管理をするのに使用する。
ランチャーから今の座席の状態が送信され、空いている座席を確認できる。

### visibility

ゲームの公開状態を表す。`public`、`limited`、`private` の3つの値をとる。

| 公開状態 | game file                | game のタイトル、説明、game image、game video |
| -------- | ------------------------ | --------------------------------------------- |
| public   | 誰でもダウンロード可能   | 誰でも閲覧可能                                |
| limited  | 部員のみダウンロード可能 | 誰でも閲覧可能                                |
| private  | 部員のみダウンロード可能 | 部員のみ閲覧可能                              |
