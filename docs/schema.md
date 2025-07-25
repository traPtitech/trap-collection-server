## DB schema

### games

| Name        | Type        | Null | Key | Default           | Extra | 説明 |
| ----------- | ----------- | ---- | --- | ----------------- | ----- | ---- |
| id          | varchar(36) | NO   | PRI |                   |       | UUID |
| name        | varchar(32) | NO   |     |                   |       |      |
| description | text        |      |     |                   |       |      |
| created_at  | datetime    | NO   |     | CURRENT_TIMESTAMP |       |      |
| deleted_at  | datetime    |      |     | NULL              |       |      |

### game_versions

| Name        | Type        | Null | Key | Default           | Extra | 説明                 |
| ----------- | ----------- | ---- | --- | ----------------- | ----- | -------------------- |
| id          | varchar(36) | NO   | PRI |                   |       | UUID                 |
| game_id     | varchar(36) | NO   | MUL |                   |       | UUID                 |
| name        | varchar(32) | NO   |     |                   |       | ゲームのバージョン名 |
| description | text        |      |     |                   |       |                      |
| created_at  | datetime    | NO   |     | CURRENT_TIMESTAMP |       |                      |

### game_assets

| Name            | Type                              | Null | Key | Default | Extra | 説明                    |
| --------------- | --------------------------------- | ---- | --- | ------- | ----- | ----------------------- |
| id              | varchar(36)                       | NO   | PRI |         |       | UUID                    |
| game_version_id | int(11)                           | NO   | MUL |         |       |                         |
| type            | enum('url','jar','windows','mac') | NO   |     |         |       | `                       |
| md5             | char(32)                          |      |     |         |       | type が`url`時は NULL   |
| url             | text                              |      |     |         |       | type が`url`以外で NULL |

### game_introductions

| Name       | Type        | Null | Key | Default           | Extra | 説明                             |
| ---------- | ----------- | ---- | --- | ----------------- | ----- | -------------------------------- |
| id         | varchar(36) | NO   | PRI |                   |       | UUID                             |
| game_id    | varchar(36) | NO   | MUL |                   |       | UUID                             |
| role       | tinyint     | NO   |     |                   |       | 0:`image`,1:`video`              |
| extension  | tinyint     | NO   |     |                   |       | 0:`jpeg`,1:`png`,2:`gif`,3:`mp4` |
| created_at | datetime    | NO   |     | CURRENT_TIMESTAMP |       |                                  |

### maintainers

| Name       | Type        | Null | Key | Default           | Extra | 説明                                                                          |
| ---------- | ----------- | ---- | --- | ----------------- | ----- | ----------------------------------------------------------------------------- |
| id         | varchar(36) | NO   | PRI |                   |       | UUID                                                                          |
| game_id    | varchar(36) | NO   | MUL |                   |       | UUID                                                                          |
| user_id    | varchar(32) | NO   |     |                   |       | traPID(UUID)                                                                  |
| role       | tinyint     | NO   |     | 0                 |       | 0:ゲームの更新の権限のみ,1:ゲームの更新と更新権限を持つ使途の追加の権限を持つ |
| created_at | datetime    | NO   |     | CURRENT_TIMESTAMP |       |                                                                               |
| deleted_at | datetime    |      |     | NULL              |       |                                                                               |

### launcher_versions

| Name        | Type        | Null | Key | Default           | Extra | 説明 |
| ----------- | ----------- | ---- | --- | ----------------- | ----- | ---- |
| id          | varchar(36) | NO   | PRI |                   |       | UUID |
| name        | varchar(32) | NO   | UNI |                   |       |      |
| anke_to_url | text        |      |     | NULL              |       |      |
| created_at  | datetime    | NO   |     | CURRENT_TIMESTAMP |       |      |
| deleted_at  | datetime    |      |     | NULL              |       |      |

### game_version_relations

| Name                | Type        | Null | Key | Default | Extra | 説明 |
| ------------------- | ----------- | ---- | --- | ------- | ----- | ---- |
| launcher_version_id | varchar(36) | NO   | MUL |         |       |      |
| game_id             | varchar(36) | NO   | MUL |         |       |      |

### product_keys

| Name                | Type        | Null | Key | Default | Extra | 説明 |
| ------------------- | ----------- | ---- | --- | ------- | ----- | ---- | --- |
| id                  | varchar(36) | NO   | PRI |         |       | UUID |     |
| key                 | char(29)    | NO   | UNI |         |       |      |
| launcher_version_id | varchar(36) | NO   | MUL |         |       |      |
| created_at          | datetime    | NO   |     |         |       |      |
| deleted_at          | datetime    |      |     | NULL    |       |      |

### access_tokens

| Name         | Type        | Null | Key | Default | Extra | 説明 |
| ------------ | ----------- | ---- | --- | ------- | ----- | ---- |
| id           | varchar(36) | NO   | PRI |         |       | UUID |
| key_id       | varchar(36) | NO   | MUL |         |       |      |
| access_token | varchar(36) | NO   | UNI |         |       |      |
| expires_at   | datetime    | NO   |     |         |       |      |
| created_at   | datetime    | NO   |     |         |       |      |
| deleted_at   | datetime    |      |     | NULL    |       |      |

### seat_versions

| Name       | Type        | Null | Key | Default           | Extra | 説明 |
| ---------- | ----------- | ---- | --- | ----------------- | ----- | ---- |
| id         | varchar(36) | NO   | PRI |                   |       | UUID |
| width      | int(11)     | NO   |     |                   |       |      |
| height     | int(11)     | NO   |     |                   |       |      |
| created_at | datetime    | NO   |     | CURRENT_TIMESTAMP |       |      |
| deleted_at | datetime    |      |     | NULL              |       |      |

### seats

| Name            | Type        | Null | Key | Default           | Extra | 説明     |
| --------------- | ----------- | ---- | --- | ----------------- | ----- | -------- |
| id              | varchar(36) | NO   | PRI |                   |       | UUID     |
| seat_version_id | int(11)     | NO   | MUL |                   |       |          |
| row             | int(11)     | NO   |     |                   |       |          |
| column          | int(11)     | NO   |     |                   |       |          |
| started_at      | datetime    | NO   |     | CURRENT_TIMESTAMP |       | 着席時刻 |
| ended_at        | datetime    |      |     | NULL              |       | 離席時刻 |

### game_play_logs

| Name            | Type        | Null | Key | Default           | Extra                       | 説明           |
| --------------- | ----------- | ---- | --- | ----------------- | --------------------------- | -------------- |
| id              | varchar(36) | NO   | PRI |                   |                             | UUID           |
| edition_id      | varchar(36) | NO   | MUL |                   |                             |                |
| game_id         | varchar(36) | NO   | MUL |                   |                             |                |
| game_version_id | varchar(36) | NO   | MUL |                   |                             |                |
| start_time      | datetime    | NO   |     | CURRENT_TIMESTAMP |                             | ゲーム起動時刻 |
| end_time        | datetime    |      |     | NULL              |                             | ゲーム終了時刻 |
| duration        | int         |      |     | NULL              |                             | プレイ時間     |
| created_at      | datetime    | NO   |     | CURRENT_TIMESTAMP |                             |                |
| updated_at      | datetime    | NO   |     | CURRENT_TIMESTAMP | ON UPDATE CURRENT_TIMESTAMP |                |
