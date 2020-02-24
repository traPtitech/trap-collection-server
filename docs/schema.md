## DB schema

### games
| Name | Type | Null | Key | Default | Extra | 説明 |
| --- | --- | --- | --- | --- | --- | --- |
| id | varchar(36) | NO | PRI |  |  | UUID |
| name | varchar(32) | NO |  |  |  |  |
| type | varchar(7) | NO |  | "browser" |  | (browser,java,exe) |
| md5 | binary(16) |  |  |  |  |  |
| updated_at | datetime | NO |  | CURRENT_TIMESTAMP |  |  |
| deleted_at | datetime |  |  | NULL |  |  |

### versions
| Name | Type | Null | Key | Default | Extra | 説明 |
| --- | --- | --- | --- | --- | --- | --- |
| id | int(11) | NO | PRI |  | AUTO_INCREMENT,unsigned |  |
| name | varchar(32) | NO | UNI |  |  |  |
| created_at | datetime | NO |  | CURRENT_TIMESTAMP |  |  |
| deleted_at | datetime |  |  | NULL |  |  |

### seats
| Name | Type | Null | Key | Default | Extra | 説明 |
| --- | --- | --- | --- | --- | --- | --- |
| id | int(11) | NO | PRI |  | AUTO_INCREMENT,unsigned |  |
| seat_id | int(11) | NO |  |  |  |  |
| version_id | int(11) | NO | MUL |  |  |  |
| created_at | datetime | NO |  | CURRENT_TIMESTAMP |  | 着席時刻 |
| deleted_at | datetime |  |  | NULL |  | 離席時刻 |

### game_version_relations
| Name | Type | Null | Key | Default | Extra | 説明 |
| --- | --- | --- | --- | --- | --- | --- |
| id | int(11) | NO | PRI |  | AUTO_INCREMENT,unsigned |  |
| version_id | int(11) | NO | MUL |  |  |  |
| game_id | varchar(36) | NO | MUL |  |  |  |
| created_at | datetime | NO |  | CURRENT_TIMESTAMP |  |  |
| deleted_at | datetime |  |  | NULL |  |  |

### questions
| Name | Type | Null | Key | Default | Extra | 説明 |
| --- | --- | --- | --- | --- | --- | --- |
| id | int(11) | NO | PRI |  | AUTO_INCREMENT,unsigned |  |
| version_id | int(11) | NO | MUL |  |  |  |
| type | varchar(8) | NO |  |  | 'radio','checkbox','text' |  |
| content | text | NO |  |  |  | 質問文 |
| required | boolean | NO |  | true |  |  |
| created_at | datetime | NO |  | CURRENT_TIMESTAMP |  |  |
| deleted_at | datetime |  |  | NULL |  |  |

### choices
| Name | Type | Null | Key | Default | Extra | 説明 |
| --- | --- | --- | --- | --- | --- | --- |
| id | int(11) | NO | PRI |  | AUTO_INCREMENT,unsigned |  |
| question_id | int(11) | NO | MUL |  |  |  |
| text | text | NO |  |  |  |  |

### responses
| Name | Type | Null | Key | Default | Extra | 説明 |
| --- | --- | --- | --- | --- | --- | --- |
| id | varchar(36) | NO | PRI |  |  |  |
| remark | text |  |  |  |  |  |
| created_at | datetime | NO |  | CURRENT_TIMESTAMP |  |  |

### answer_responses
| Name | Type | Null | Key | Default | Extra | 説明 |
| --- | --- | --- | --- | --- | --- | --- |
| id | int(11) | NO | PRI |  | AUTO_INCREMENT,unsigned |  |
| response_id | varchar(36) | NO | MUL |  |  |  |
| question_id | int(11) | NO | MUL |  |  |  |
| content | text | NO |  |  |  |  |

### game_responses
| Name | Type | Null | Key | Default | Extra | 説明 |
| --- | --- | --- | --- | --- | --- | --- |
| id | int(11) | NO | PRI |  | AUTO_INCREMENT,unsigned |  |
| response_id | varchar(36) | NO | MUL |  |  |  |
| game_id | varchar(36) | NO | MUL |  |  |  |
| star | tinyint(3) | NO |  |  | unsigned |  |
| time | int(11) | NO |  |  | unsigned |  |
