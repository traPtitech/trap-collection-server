# game_genre_relations

## Description

ゲームとジャンルの関係テーブル

<details>
<summary><strong>Table Definition</strong></summary>

```sql
CREATE TABLE `game_genre_relations` (
  `genre_id` varchar(36) NOT NULL,
  `game_id` varchar(36) NOT NULL,
  PRIMARY KEY (`game_id`,`genre_id`),
  KEY `fk_game_genre_relations_game_genre_table` (`genre_id`),
  CONSTRAINT `fk_game_genre_relations_game_genre_table` FOREIGN KEY (`genre_id`) REFERENCES `game_genres` (`id`),
  CONSTRAINT `fk_game_genre_relations_game_table2` FOREIGN KEY (`game_id`) REFERENCES `games` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4
```

</details>

## Columns

| Name | Type | Default | Nullable | Children | Parents | Comment |
| ---- | ---- | ------- | -------- | -------- | ------- | ------- |
| genre_id | varchar(36) |  | false |  | [game_genres](game_genres.md) | ゲームのジャンルのID |
| game_id | varchar(36) |  | false |  | [games](games.md) | ゲームUUID |

## Constraints

| Name | Type | Definition |
| ---- | ---- | ---------- |
| fk_game_genre_relations_game_genre_table | FOREIGN KEY | FOREIGN KEY (genre_id) REFERENCES game_genres (id) |
| fk_game_genre_relations_game_table2 | FOREIGN KEY | FOREIGN KEY (game_id) REFERENCES games (id) |
| PRIMARY | PRIMARY KEY | PRIMARY KEY (game_id, genre_id) |

## Indexes

| Name | Definition |
| ---- | ---------- |
| fk_game_genre_relations_game_genre_table | KEY fk_game_genre_relations_game_genre_table (genre_id) USING BTREE |
| PRIMARY | PRIMARY KEY (game_id, genre_id) USING BTREE |

## Relations

![er](game_genre_relations.svg)

---

> Generated by [tbls](https://github.com/k1LoW/tbls)
