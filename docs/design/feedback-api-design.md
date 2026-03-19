# フィードバック機能 設計書

## 概要

ランチャーでゲームをプレイしたユーザーからフィードバックを収集し、管理画面で確認できる機能を実装する。

### 関連Issue
- [#1058 ランチャーで遊んだ人からのフィードバックを集める](https://github.com/traPtitech/trap-collection-server/issues/1058)
- [#1316 フィードバック機能のOpenAPIとDBスキーマの設計](https://github.com/traPtitech/trap-collection-server/issues/1316)
- [#1283 ゲーム班へのフィードバックヒアリング](https://github.com/traPtitech/trap-collection-server/issues/1283) (完了)

## 要件まとめ

### ヒアリング結果からの決定事項
1. **回答形式**: Yes/No形式に加え、5段階評価にも対応
2. **質問はゲーム単位**: エディションに依存せず、常に最新の質問が表示される
3. **自由記述欄**: フィードバック全体に1つ
4. **質問の保存場所**: データベースに保存（プログラムにハードコードしない）
5. **質問の管理**: PUTで質問リスト全体を一括差し替え
6. **アーカイブ**: 不要になった質問はアーカイブ。過去の回答は取得可能だが、新規回答対象外
7. **削除**: 削除された質問の回答は取得不可

### ユーザー側（ランチャー）
- プレイしたゲームに対して、任意で質問への回答を送信可能
- EditionID + GameVersionIDに紐づけてフィードバックを記録

### 管理画面側
- ゲームのmaintainer以上がフィードバック質問を管理可能
- 全スタッフがアンケート回答を閲覧可能
- ゲームバージョンごとに回答を確認（エディション情報付き）
- ゲームごとに回答を確認（バージョン・エディション情報付き）

---

## データベース設計

### ER図（概念）

```
game_feedback_configs (フィードバック設定 — ゲーム単位)
├── game_id: UUID (PK, FK -> games)
├── enabled: BOOLEAN                   -- フィードバック機能のon/off
└── [FK] game -> GameTable2

feedback_questions (質問マスタ — ゲーム単位)
├── id: UUID (PK)
├── game_id: UUID (FK -> games)         -- ゲーム
├── question_text: VARCHAR(256)         -- 質問文
├── answer_type: TINYINT                -- 回答形式 (0=yesNo, 1=fiveScale)
├── question_order: INT                 -- 表示順序
├── created_at: DATETIME
├── archived_at: DATETIME NULL          -- アーカイブ日時（過去回答は取得可、新規回答対象外）
├── deleted_at: DATETIME NULL           -- 論理削除（回答も取得不可）
└── [FK] game -> GameTable2

game_feedbacks (フィードバック本体)
├── id: UUID (PK)
├── edition_id: UUID (FK -> editions)
├── game_version_id: UUID (FK -> v2_game_versions)
├── comment: TEXT NULL -- 自由記述欄
├── created_at: DATETIME
├── [FK] edition -> EditionTable
└── [FK] game_version -> GameVersionTable2

game_feedback_answers (質問ごとの回答)
├── id: UUID (PK)
├── feedback_id: UUID (FK -> game_feedbacks)
├── question_id: UUID (FK -> feedback_questions)
├── answer: INT                         -- 回答値 (yesNo: 0/1, fiveScale: 1-5)
├── [FK] feedback -> GameFeedbackTable
└── [FK] question -> FeedbackQuestionTable
```

### Go構造体定義（実装済み: `src/repository/gorm2/schema/v2.go`）

```go
// GameFeedbackConfigTable フィードバック設定（ゲーム単位のon/off）
type GameFeedbackConfigTable struct {
	GameID  uuid.UUID  `gorm:"type:varchar(36);not null;primaryKey"`
	Enabled bool       `gorm:"type:boolean;not null;default:false"`
	Game    GameTable2 `gorm:"foreignKey:GameID"`
}

// FeedbackQuestionTable フィードバック質問マスタ
// ゲーム単位でYes/Noまたは5段階評価の質問を管理する
type FeedbackQuestionTable struct {
	ID            uuid.UUID      `gorm:"type:varchar(36);not null;primaryKey"`
	GameID        uuid.UUID      `gorm:"type:varchar(36);not null;index"`
	QuestionText  string         `gorm:"type:varchar(256);not null"`
	AnswerType    int            `gorm:"type:tinyint;not null"`
	QuestionOrder int            `gorm:"type:int;not null"`
	CreatedAt     time.Time      `gorm:"type:datetime;not null;default:CURRENT_TIMESTAMP"`
	ArchivedAt    sql.NullTime   `gorm:"type:DATETIME NULL;default:NULL"`
	DeletedAt     gorm.DeletedAt `gorm:"type:DATETIME NULL;default:NULL"`
	Game          GameTable2     `gorm:"foreignKey:GameID"`
}

// GameFeedbackTable フィードバック本体
type GameFeedbackTable struct {
	ID            uuid.UUID                 `gorm:"type:varchar(36);not null;primaryKey"`
	EditionID     uuid.UUID                 `gorm:"type:varchar(36);not null;index"`
	GameVersionID uuid.UUID                 `gorm:"type:varchar(36);not null;index"`
	Comment       sql.NullString            `gorm:"type:text;default:NULL"`
	CreatedAt     time.Time                 `gorm:"type:datetime;not null;default:CURRENT_TIMESTAMP"`
	Edition       EditionTable              `gorm:"foreignKey:EditionID"`
	GameVersion   GameVersionTable2         `gorm:"foreignKey:GameVersionID"`
	Answers       []GameFeedbackAnswerTable `gorm:"foreignKey:FeedbackID"`
}

// GameFeedbackAnswerTable 各質問への回答
type GameFeedbackAnswerTable struct {
	ID         uuid.UUID             `gorm:"type:varchar(36);not null;primaryKey"`
	FeedbackID uuid.UUID             `gorm:"type:varchar(36);not null;index"`
	QuestionID uuid.UUID             `gorm:"type:varchar(36);not null;index"`
	Answer     int                   `gorm:"type:int;not null"`
	Feedback   GameFeedbackTable     `gorm:"foreignKey:FeedbackID"`
	Question   FeedbackQuestionTable `gorm:"foreignKey:QuestionID"`
}
```

### マイグレーションファイル（生成済み）

`migrations/20260225125932_create_game_feedbacks.sql`

---

## API設計（実装済み: `docs/openapi/v2.yaml`）

### エンドポイント一覧

| メソッド | パス | 認証 | 説明 |
|---------|------|------|------|
| GET | `/games/{gameID}/feedback-questions` | GameMaintainerAuth or EditionAuth | フィードバック質問一覧取得（有効な質問のみ） |
| PUT | `/games/{gameID}/feedback-questions` | GameMaintainerAuth | フィードバック質問の一括設定 |
| POST | `/editions/{editionID}/games/{gameID}/feedbacks` | EditionAuth | フィードバック送信（ランチャー用） |
| GET | `/games/{gameID}/feedbacks` | TrapMemberAuth | ゲームのフィードバック一覧取得 |
| GET | `/games/{gameID}/versions/{gameVersionID}/feedbacks` | TrapMemberAuth | ゲームバージョンのフィードバック一覧取得 |

### 主要スキーマ

#### AnswerType（回答形式）
```yaml
AnswerType:
  type: string
  enum: [yesNo, fiveScale]
```

#### FeedbackQuestionsConfig（質問設定レスポンス）
```yaml
FeedbackQuestionsConfig:
  enabled: boolean          # フィードバック機能のon/off
  questions: FeedbackQuestion[]
```

#### PutFeedbackQuestionsRequest（質問一括設定リクエスト）
```yaml
PutFeedbackQuestionsRequest:
  enabled: boolean          # フィードバック機能のon/off
  questions: FeedbackQuestionInput[]
  # 配列の順序がquestion_orderになる
  # リストから外された既存の質問はアーカイブされる
```

#### FeedbackAnswerInput（回答入力）
```yaml
FeedbackAnswerInput:
  properties:
    questionID: UUID
    answer: integer (min: 0, max: 5)  # yesNo: 0/1, fiveScale: 1-5
```

#### FeedbackAnswer（回答レスポンス）
```yaml
FeedbackAnswer:
  properties:
    questionID: UUID
    questionText: string
    answerType: AnswerType
    answer: integer
```

---

## 質問の状態遷移

```
有効 (active)
 ├── PUTリストから除外 → アーカイブ (archived_at に日時が入る)
 └── 削除操作 → 削除 (deleted_at に日時が入る)

アーカイブ (archived)
 ├── GETで返されない、新規回答対象外
 ├── 過去の回答はフィードバック取得APIで返される
 └── 削除操作 → 削除

削除 (deleted)
 ├── GETで返されない、新規回答対象外
 └── 過去の回答もフィードバック取得APIで返されない
```

---

## 実装フェーズ

### Phase 1: DBスキーマ定義とマイグレーション ✅ 完了
- [x] `src/repository/gorm2/schema/v2.go` にテーブル定義を追加
- [x] `task migrate:new -- create_game_feedbacks` でマイグレーションファイル生成

### Phase 2: OpenAPI定義 ✅ 完了
- [x] `docs/openapi/v2.yaml` にAPI定義を追加

### Phase 3: Domain層実装（TODO）
- [ ] `src/domain/` にフィードバック関連のエンティティを定義
- [ ] `src/domain/values/` にフィードバック関連の値オブジェクトを追加

### Phase 4: Repository層実装（TODO）
- [ ] `src/repository/` にインターフェースを定義
- [ ] `src/repository/gorm2/` に実装を追加

### Phase 5: Service層実装（TODO）
- [ ] `src/service/v2/` にビジネスロジックを実装

### Phase 6: Handler層実装（TODO）
- [ ] `go generate` でハンドラーコードを生成
- [ ] `src/handler/v2/` にAPIハンドラーを実装

### Phase 7: テスト（TODO）
- [ ] 各層のユニットテスト
- [ ] 統合テスト

---

## 検討事項・未決定事項

### 1. 回答バリデーション
- サービス層で質問の `answer_type` に基づき回答値を検証する
  - yesNo: 0 または 1
  - fiveScale: 1〜5

### 2. 質問の一括管理
- PUTで質問リスト全体を差し替える方式を採用
- 配列の順序がそのまま `question_order` になる
- 送信されなかった既存の質問はアーカイブされる（削除ではない）

### 3. フィードバック機能のon/off
- `game_feedback_configs` テーブルで `enabled` フラグを管理
- PUTリクエストの `enabled` で切り替え可能
- 質問を保持したままフィードバック収集を一時停止できる
- ランチャーはGETレスポンスの `enabled` を見て表示/非表示を判断

### 4. 回答の匿名性
- ランチャーからの回答はEdition認証のみで、ユーザー情報は紐づけない
- 完全匿名のフィードバックとして扱う

### 5. フィードバック統計API
- 現時点では統計APIは実装しない（将来対応）
- 例: 質問ごとの回答分布など

### 6. 重複送信の制御
- 同一ユーザーからの重複送信を許可するか
- 現設計では制限なし（匿名のため識別不可）

### 7. ソフトデリートとユニーク制約
- `(game_id, question_order)` のユニーク制約は付けない
- 削除済み・アーカイブ済みレコードが残るためMySQLでは部分インデックスが使えない
- order の一意性はサービス層で保証する

---

## 変更履歴

| 日付 | 内容 |
|------|------|
| 2026-01-21 | 初版作成、DBスキーマとOpenAPI定義を実装 |
| 2026-01-21 | 仕様に基づきシンプル化（Yes/No形式のみ、自由記述欄はフィードバック全体に1つ） |
| 2026-02-25 | 5段階評価対応、回答をintに変更、APIをPUT一括設定方式に変更 |
| 2026-02-25 | 質問をゲーム単位に変更（edition_id除去）、アーカイブ機能追加、フィードバックon/off機能追加 |
