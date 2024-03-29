# v2_game_files

## Description

ゲームファイルテーブル(v2)

<details>
<summary><strong>Table Definition</strong></summary>

```sql
CREATE TABLE `v2_game_files` (
  `id` varchar(36) NOT NULL,
  `game_id` varchar(36) NOT NULL,
  `file_type_id` tinyint(4) NOT NULL,
  `hash` char(32) NOT NULL,
  `entry_point` text NOT NULL,
  `created_at` datetime NOT NULL DEFAULT current_timestamp(),
  PRIMARY KEY (`id`),
  KEY `fk_games_game_files` (`game_id`),
  KEY `fk_v2_game_files_game_file_type` (`file_type_id`),
  CONSTRAINT `fk_games_game_files` FOREIGN KEY (`game_id`) REFERENCES `games` (`id`),
  CONSTRAINT `fk_v2_game_files_game_file_type` FOREIGN KEY (`file_type_id`) REFERENCES `game_file_types` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4
```

</details>

## Columns

| Name | Type | Default | Nullable | Children | Parents | Comment |
| ---- | ---- | ------- | -------- | -------- | ------- | ------- |
| id | varchar(36) |  | false | [game_version_game_file_relations](game_version_game_file_relations.md) |  | ゲームファイルUUID |
| game_id | varchar(36) |  | false |  | [games](games.md) | ゲームUUID |
| file_type_id | tinyint(4) |  | false |  | [game_file_types](game_file_types.md) | ゲームファイルの種類のUUID |
| hash | char(32) |  | false |  |  | ファイルのmd5ハッシュ |
| entry_point | text |  | false |  |  | 実行時のエントリーポイント |
| created_at | datetime | current_timestamp() | false |  |  | 作成日時 |

## Constraints

| Name | Type | Definition |
| ---- | ---- | ---------- |
| fk_games_game_files | FOREIGN KEY | FOREIGN KEY (game_id) REFERENCES games (id) |
| fk_v2_game_files_game_file_type | FOREIGN KEY | FOREIGN KEY (file_type_id) REFERENCES game_file_types (id) |
| PRIMARY | PRIMARY KEY | PRIMARY KEY (id) |

## Indexes

| Name | Definition |
| ---- | ---------- |
| fk_games_game_files | KEY fk_games_game_files (game_id) USING BTREE |
| fk_v2_game_files_game_file_type | KEY fk_v2_game_files_game_file_type (file_type_id) USING BTREE |
| PRIMARY | PRIMARY KEY (id) USING BTREE |

## Relations

![er](v2_game_files.svg)

---

> Generated by [tbls](https://github.com/k1LoW/tbls)
