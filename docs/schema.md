## DB schema

### games
| Name | Type | Null | Key | Default | Extra | 説明 |
| --- | --- | --- | --- | --- | --- | --- |
| id | varchar(36) | NO | PRI |  |  | UUID |
| name | varchar(32) | NO |  |  |  |  |
| description | text |  |  |  |  |  |
| created_at | datetime | NO |  | CURRENT_TIMESTAMP |  |  |
| deleted_at | datetime |  |  | NULL |  |  |

### game_versions
| Name | Type | Null | Key | Default | Extra | 説明 |
| --- | --- | --- | --- | --- | --- | --- |
| id | int(11) | NO | PRI |  | AUTO_INCREMENT,unsigned |  |
| game_id | varchar(36) | NO | MUL |  |  | UUID |
| name | varchar(32) | NO |  |  |  | ゲームのバージョン名 |
| description | text |  |  |  |  |  |
| created_at | datetime | NO |  | CURRENT_TIMESTAMP |  |  |

### game_assets
| Name | Type | Null | Key | Default | Extra | 説明 |
| --- | --- | --- | --- | --- | --- | --- |
| id | int(11) | NO | PRI |  | AUTO_INCREMENT,unsigned |  |
| game_version_id | int(11) | NO | MUL |  |  |  |
| type | tinyint | NO |  |  |  | 0:`url`,1:`jar`,2:`windows`,3:`mac` |
| md5 | char(32) |  |  |  |  | typeが`url`時はNULL |
| url | text |  |  |  |  | typeが`url`以外でNULL |

### game_introductions
| Name | Type | Null | Key | Default | Extra | 説明 |
| --- | --- | --- | --- | --- | --- | --- |
| id | int(11) | NO | PRI |  | AUTO_INCREMENT,unsigned |  |
| game_id | varchar(36) | NO | MUL |  |  | UUID |
| role | tinyint | NO |  |  |  | 0:`image`,1:`video` |
| extension | tinyint | NO |  |  |  | 0:`jpeg`,1:`png`,2:`gif`,3:`mp4` |
| created_at | datetime | NO |  | CURRENT_TIMESTAMP |  |  |

### maintainers
| Name | Type | Null | Key | Default | Extra | 説明 |
| --- | --- | --- | --- | --- | --- | --- |
| id | int(11) | NO | PRI |  | AUTO_INCREMENT,unsigned |  |
| game_id | varchar(36) | NO | MUL |  |  | UUID |
| user_id | varchar(32) | NO |  |  |  | traPID(UUID) |
| role | tinyint | NO |  | 0 |  | 0:ゲームの更新の権限のみ,1:ゲームの更新と更新権限を持つ使途の追加の権限を持つ |
| created_at | datetime | NO |  | CURRENT_TIMESTAMP |  |  |
| deleted_at | datetime |  |  | NULL |  |  |

### launcher_versions
| Name | Type | Null | Key | Default | Extra | 説明 |
| --- | --- | --- | --- | --- | --- | --- |
| id | int(11) | NO | PRI |  | AUTO_INCREMENT,unsigned |  |
| name | varchar(32) | NO | UNI |  |  |  |
| anke-to | text |  |  | NULL |  |  |
| created_at | datetime | NO |  | CURRENT_TIMESTAMP |  |  |
| deleted_at | datetime |  |  | NULL |  |  |

### product_key
| Name | Type | Null | Key | Default | Extra | 説明 |
| --- | --- | --- | --- | --- | --- | --- |
| id | int(11) | NO | PRI |  | AUTO_INCREMENT,unsigned |  |
| key | varchar(36) | NO | UNI |  |  |  |
| launcher_version_id | int(11) | NO | MUL |  |  |  |
| used | boolean | NO |  | false |  | access_tokenが使用済みかどうか |

### game_version_relations
| Name | Type | Null | Key | Default | Extra | 説明 |
| --- | --- | --- | --- | --- | --- | --- |
| launcher_version_id | int(11) | NO | MUL |  |  |  |
| game_id | varchar(36) | NO | MUL |  |  |  |

### players
| Name | Type | Null | Key | Default | Extra | 説明 |
| --- | --- | --- | --- | --- | --- | --- |
| id | int(11) | NO | PRI |  | AUTO_INCREMENT,unsigned |  |
| product_key_id | int(11) | NO | MUL |  |  |  |
| started_at | datetime | NO |  | CURRENT_TIMESTAMP |  | 着席時刻 |
| ended_at | datetime |  |  | NULL |  | 離席時刻 |
