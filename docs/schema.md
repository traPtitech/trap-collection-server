## DB schema

### game_metas
| Name | Type | Null | Key | Default | Extra | 説明 |
| --- | --- | --- | --- | --- | --- | --- |
| id | varchar(36) | NO | PRI |  |  | UUID |
| name | varchar(32) | NO |  |  |  |  |
| type | tinyint | NO |  |  |  | 0:`browser`,1:`java`,2:`exe` |
| created_at | datetime | NO |  | CURRENT_TIMESTAMP |  |  |
| deleted_at | datetime |  |  | NULL |  |  |

### game_versions
| Name | Type | Null | Key | Default | Extra | 説明 |
| --- | --- | --- | --- | --- | --- | --- |
| id | int(11) | NO | PRI |  | AUTO_INCREMENT,unsigned |  |
| game_id | varchar(36) | NO | MUL |  |  | UUID |
| name | varchar(32) | NO |  |  |  | ゲームのバージョン名 |
| description | text |  |  |  |  |  |
| md5 | binary(16) |  |  |  |  | typeが`browser`時はNULL |
| url | text |  |  |  |  | typeが`browser`以外でNULL |
| created_at | datetime | NO |  | CURRENT_TIMESTAMP |  |  |

### game_assets
| Name | Type | Null | Key | Default | Extra | 説明 |
| --- | --- | --- | --- | --- | --- | --- |
| id | int(11) | NO | PRI |  | AUTO_INCREMENT,unsigned |  |
| game_version_id | int(11) | NO | MUL |  |  |  |
| type | tinyint | NO |  |  |  | 0:`url`,1:`jar`,2:`windows`,3:`mac` |

### game_introductions
| Name | Type | Null | Key | Default | Extra | 説明 |
| --- | --- | --- | --- | --- | --- | --- |
| id | int(11) | NO | PRI |  | AUTO_INCREMENT,unsigned |  |
| game_id | varchar(36) | NO | MUL |  |  | UUID |
| role | tinyint | NO |  |  |  | 0:`image`,1:`movie` |
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
| created_at | datetime | NO |  | CURRENT_TIMESTAMP |  |  |
| deleted_at | datetime |  |  | NULL |  |  |

### sessions
| Name | Type | Null | Key | Default | Extra | 説明 |
| --- | --- | --- | --- | --- | --- | --- |
| id | int(11) | NO | PRI |  | AUTO_INCREMENT,unsigned |  |
| seat_id | int(11) | NO |  |  |  |  |
| started_at | datetime | NO |  | CURRENT_TIMESTAMP |  | 着席時刻 |
| ended_at | datetime |  |  | NULL |  | 離席時刻 |

### game_version_relations
| Name | Type | Null | Key | Default | Extra | 説明 |
| --- | --- | --- | --- | --- | --- | --- |
| version_id | int(11) | NO | MUL |  |  |  |
| game_id | varchar(36) | NO | MUL |  |  |  |

### questions
| Name | Type | Null | Key | Default | Extra | 説明 |
| --- | --- | --- | --- | --- | --- | --- |
| id | int(11) | NO | PRI |  | AUTO_INCREMENT,unsigned |  |
| version_id | int(11) | NO | MUL |  |  |  |
| type | tinyint | NO |  |  |  | 0:`radio`,1:`checkbox`,2:`text` |
| content | text | NO |  |  |  | 質問文 |
| required | boolean | NO |  | true |  |  |
| created_at | datetime | NO |  | CURRENT_TIMESTAMP |  |  |
| deleted_at | datetime |  |  | NULL |  |  |

### question_options
| Name | Type | Null | Key | Default | Extra | 説明 |
| --- | --- | --- | --- | --- | --- | --- |
| id | int(11) | NO | PRI |  | AUTO_INCREMENT,unsigned |  |
| question_id | int(11) | NO | MUL |  |  |  |
| label | text | NO |  |  |  |  |

### responses
| Name | Type | Null | Key | Default | Extra | 説明 |
| --- | --- | --- | --- | --- | --- | --- |
| id | varchar(36) | NO | PRI |  |  |  |
| session_id | int(11) | NO | MUL |  |  |  |
| version_id | int(11) | NO | MUL |  |  |  |
| remark | text |  |  |  |  |  |
| created_at | datetime | NO |  | CURRENT_TIMESTAMP |  |  |

### answer_responses
| Name | Type | Null | Key | Default | Extra | 説明 |
| --- | --- | --- | --- | --- | --- | --- |
| id | int(11) | NO | PRI |  | AUTO_INCREMENT,unsigned |  |
| response_id | varchar(36) | NO | MUL |  |  |  |
| question_id | int(11) | NO | MUL |  |  |  |
| content | text | NO |  |  |  |  |

### game_ratings
| Name | Type | Null | Key | Default | Extra | 説明 |
| --- | --- | --- | --- | --- | --- | --- |
| id | int(11) | NO | PRI |  | AUTO_INCREMENT,unsigned |  |
| response_id | varchar(36) | NO | MUL |  |  |  |
| game_version_id | int(11) | NO | MUL |  |  |  |
| star | tinyint | NO |  |  | unsigned |  |
| play_time | int(11) | NO |  |  | unsigned | プレイ時間(ms) |
