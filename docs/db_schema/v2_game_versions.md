# v2_game_versions

## Description

ゲームバージョンテーブル(v2)

<details>
<summary><strong>Table Definition</strong></summary>

```sql
CREATE TABLE `v2_game_versions` (
  `id` varchar(36) NOT NULL,
  `game_id` varchar(36) NOT NULL,
  `game_image_id` varchar(36) NOT NULL,
  `game_video_id` varchar(36) NOT NULL,
  `name` varchar(32) NOT NULL,
  `description` text NOT NULL,
  `created_at` datetime NOT NULL DEFAULT current_timestamp(),
  PRIMARY KEY (`id`),
  KEY `fk_v2_game_versions_game_image` (`game_image_id`),
  KEY `fk_v2_game_versions_game_video` (`game_video_id`),
  CONSTRAINT `fk_v2_game_versions_game_image` FOREIGN KEY (`game_image_id`) REFERENCES `v2_game_images` (`id`),
  CONSTRAINT `fk_v2_game_versions_game_video` FOREIGN KEY (`game_video_id`) REFERENCES `v2_game_videos` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4
```

</details>

## Columns

| Name | Type | Default | Nullable | Children | Parents | Comment |
| ---- | ---- | ------- | -------- | -------- | ------- | ------- |
| id | varchar(36) |  | false | [edition_game_version_relations](edition_game_version_relations.md) [game_version_game_file_relations](game_version_game_file_relations.md) |  | ゲームバージョンUUID |
| game_id | varchar(36) |  | false |  |  | ゲームUUID |
| game_image_id | varchar(36) |  | false |  | [v2_game_images](v2_game_images.md) | ゲーム画像UUID |
| game_video_id | varchar(36) |  | false |  | [v2_game_videos](v2_game_videos.md) | ゲーム動画UUID |
| name | varchar(32) |  | false |  |  | ゲームバージョン名 |
| description | text |  | false |  |  | ゲームバージョンの説明 |
| created_at | datetime | current_timestamp() | false |  |  | 作成日時 |

## Constraints

| Name | Type | Definition |
| ---- | ---- | ---------- |
| fk_v2_game_versions_game_image | FOREIGN KEY | FOREIGN KEY (game_image_id) REFERENCES v2_game_images (id) |
| fk_v2_game_versions_game_video | FOREIGN KEY | FOREIGN KEY (game_video_id) REFERENCES v2_game_videos (id) |
| PRIMARY | PRIMARY KEY | PRIMARY KEY (id) |

## Indexes

| Name | Definition |
| ---- | ---------- |
| fk_v2_game_versions_game_image | KEY fk_v2_game_versions_game_image (game_image_id) USING BTREE |
| fk_v2_game_versions_game_video | KEY fk_v2_game_versions_game_video (game_video_id) USING BTREE |
| PRIMARY | PRIMARY KEY (id) USING BTREE |

## Relations

![er](v2_game_versions.svg)

---

> Generated by [tbls](https://github.com/k1LoW/tbls)