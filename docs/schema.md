## DB schema(現在)

現時点でのDBの状態。工大祭までに間に合わせようと急いでコードを書いた結果、version_for_sale,version_not_for_saleがほぼ同じ内容にもかかわらず分かれていたり、Keyが一つもなかったり、idがあるのにnameをKeyにしていたり、etc、自明にまずい点が大量にあるので、これから大幅に修正していく。

### game
| Name | Type | Null | Key | Default | Extra | 説明 |
| --- | --- | --- | --- | --- | --- | --- |
| id | varchar(36) |  |  |  |  | UUID |
| name | varchar(30) |  |  |  |  | ゲーム名 |
| container | text |  |  |  |  | ConoHaオブジェクトストレージ上のコンテナ名 |
| file_name | text |  |  |  |  | ConoHaオブジェクトストレージ上のファイル名 |
| md5 | binary(16) |  |  |  |  | ファイルのmd5 |
| time | timestamp |  |  | NULL |  | ゲームを入れるか入れないかを決める基準時間 |
| created_at | timestamp |  |  | NULL |  | アップロードされた時刻 |
| updated_at | timestamp |  |  | NULL |  | 最終更新時刻 |
| deleted_at | timestamp |  |  | NULL |  | 削除された時刻 |

### versions_for_sale
| Name | Type | Null | Key | Default | Extra | 説明 |
| --- | --- | --- | --- | --- | --- | --- |
| id | varchar(36) |  |  |  |  | UUID |
| name | varchar(30) |  |  |  |  |  |
| start_period | timestamp |  |  | NULL |  | 含めるゲームの基準時間の範囲(始点) |
| end_period | timestamp |  |  | NULL |  | 含めるゲームの基準時間の範囲(終点) |
| start_time | timestamp |  |  | NULL |  | 配信開始時刻 |
| created_at | timestamp |  |  | NULL |  | バージョンの作られた時刻 |
| updated_at | timestamp |  |  | NULL |  | 最終更新時刻 |
| deleted_at | timestamp |  |  | NULL |  | 削除された時刻 |

### versions_not_for_sale
| Name | Type | Null | Key | Default | Extra | 説明 |
| --- | --- | --- | --- | --- | --- | --- |
| id | varchar(36) |  |  |  |  | UUID |
| name | varchar(30) |  |  |  |  |  |
| questionnaire_id | varchar(36) |  |  |  |  |  |
| start_period | timestamp |  |  | NULL |  | 含めるゲームの基準時間の範囲(始点) |
| end_period | timestamp |  |  | NULL |  | 含めるゲームの基準時間の範囲(終点) |
| start_time | timestamp |  |  | NULL |  | 配信開始時刻 |
| created_at | timestamp |  |  | NULL |  | バージョンの作られた時刻 |
| updated_at | timestamp |  |  | NULL |  | 最終更新時刻 |
| deleted_at | timestamp |  |  | NULL |  | 削除された時刻 |

### seat
| Name | Type | Null | Key | Default | Extra | 説明 |
| --- | --- | --- | --- | --- | --- | --- |
| id | varchar(36) |  |  |  |  |  |
| seat_id | varchar(36) |  |  |  |  | 席のid |
| created_at | timestamp |  |  |  |  | 着席時刻 |
| deleted_at | timestamp |  |  |  |  | 離席時刻 |

### play_time
| Name | Type | Null | Key | Default | Extra | 説明 |
| --- | --- | --- | --- | --- | --- | --- |
| id | varchar(36) |  |  |  |  |  |
| version_id | varchar(36) |  |  |  |  | プレイされたバージョンのid |
| game_id | varchar(36) |  |  |  |  | プレイされたゲームのid |
| start_time | timestamp |  |  |  | NULL | プレイ開始時刻 |
| end_time | timestamp |  |  |  |  | プレイ終了時刻 |

### special
| Name | Type | Null | Key | Default | Extra | 説明 |
| --- | --- | --- | --- | --- | --- | --- |
| id | varchar(36) |  |  |  |  |  |
| version_id | varchar(36) |  |  |  |  | 特例を指定するバージョンのid |
| game_name | varchar(36) |  |  |  |  | 特例に指定するゲームの名前 |
| status |  |  |  |  |  | 状態の種類("in":特例として入れる、"out":特例として外す) |
| deleted_at | timestamp |  |  |  |  | 削除時刻 |

### administrators
| Name | Type | Null | Key | Default | Extra | 説明 |
| --- | --- | --- | --- | --- | --- | --- |
| user_traqid | char(30) |  |  |  |  |  |

### options
| Name | Type | Null | Key | Default | Extra | 説明 |
| --- | --- | --- | --- | --- | --- | --- |
| id | varchar(36) |  |  |  |  |
| question_id | varchar(36) |  |  |  |  |  |
| option_num | int(11) |  |  |  |  |  |
| body |  |  |  |  |  | 選択肢の内容 |

### question
| Name | Type | Null | Key | Default | Extra | 説明 |
| --- | --- | --- | --- | --- | --- | --- |
| id | varchar(36) |  |  |  |  |
| questionnaire_id | varchar(36) |  |  |  |  | どのアンケートの質問か |
| page_num | int(11) |  |  |  |  | アンケートの何ページ目の質問か |
| question_num | int(11)    |  |  |  |  | アンケートの質問のうち、何問目か |
| type | char(20) |  |  |  |  | どのタイプの質問か ("Text", "Number", "MultipleChoice", "Checkbox", "Dropdown", "LinearScale", "Date", "Time") |
| body | text |  |  |  |  | 質問の内容 |
| is_required | tinyint(4) |  |  |  |  | 回答が必須である (1) , ない(0) |
| deleted_at | timestamp |  |  | NULL |  | 質問が削除された日時 (削除されていない場合は NULL) |
| created_at | timestamp |  |  | NULL |  | 質問が作成された日時 |

### questionnaires
| Name | Type | Null | Key | Default | Extra | 説明 |
| --- | --- | --- | --- | --- | --- | --- |
| id | varchar(36) |  |  |  |  |  |  |
| title | char(50) |  |  |  |  | アンケートのタイトル |
| description | text |  |  |  |  | アンケートの説明 |
| res_time_limit | timestamp |  |  | NULL |  | 回答の締切日時 (締切がない場合は NULL) |
| deleted_at | timestamp |  |  | NULL |  | アンケートが削除された日時 (削除されていない場合は NULL) |
| created_at | timestamp |  |  | NULL |  | アンケートが作成された日時 |
| modified_at    | timestamp |  |  | NULL |  | アンケートが更新された日時 |

### respondents
| Name | Type | Null | Key | Default | Extra | 説明 |
| --- | --- | --- | --- | --- | --- | --- |
| response_id | varchar(36) |  |  |  |  | 一つのアンケートに対する一つの回答ごとに振られる ID |
| questionnaire_id | varchar(36) |  |  |  |  | どのアンケートへの回答か |
| modified_at | timestamp |  |  | NULL |  | 回答が変更された日時 |
| submitted_at | timestamp |  |  | NULL |  | 回答が送信された日時 (未送信の場合は NULL) |
| deleted_at | timestamp |  |  | NULL |  | 回答が破棄された日時 (破棄されていない場合は NULL) |

### response
| Name | Type | Null | Key | Default | Extra | 説明 |
| --- | --- | --- | --- | --- | --- | --- |
| response_id | varchar(36) |  |  |  |  | 一つのアンケートに対する一つの回答ごとに振られる ID |
| question_id | varchar(36) |  |  |  |  | どの質問への回答か |
| body | text |  |  |  |  | 回答の内容 |
| modified_at | timestamp |  |  | NULL |  | 回答が変更された日時 |
| deleted_at  | timestamp |  |  | NULL |  | 回答が破棄された日時 (破棄されていない場合は NULL) |

### scale_labels
| Name | Type | Null | Key | Default | Extra | 説明 |
| --- | --- | --- | --- | --- | --- | --- |
| question_id | varchar(36) |  |  |  |  | どの質問のラベルか |
| scale_label_left | text |  |  |  |  | 左側のラベル (ない場合は NULL) |
| scale_label_right | text |  |  |  |  | 右側のラベル (ない場合は NULL) |
| scale_min | int(11) |  |  |  |  | スケールの最小値 |
| scale_max | int(11) |  |  |  |  | スケールの最大値 |

## DB schema(修正)

今後修正してこの形にしていきます。

### game
| Name | Type | Null | Key | Default | Extra | 説明 |
| --- | --- | --- | --- | --- | --- | --- |
| id | varchar(36) | NO | PRI |  |  | UUID |
| name | varchar(30) | NO | UNI |  |  | ゲーム名 |
| md5 | binary(16) | NO |  |  |  | ファイルのmd5 |
| time | timestamp | NO |  | NULL |  | ゲームを入れるか入れないかを決める基準時間 |
| created_at | timestamp | NO |  | NULL |  | アップロードされた時刻 |
| updated_at | timestamp |  |  | NULL |  | 最終更新時刻 |
| deleted_at | timestamp |  |  | NULL |  | 削除された時刻 |

### versions
| Name | Type | Null | Key | Default | Extra | 説明 |
| --- | --- | --- | --- | --- | --- | --- |
| id | varchar(36) | NO | PRI |  |  | UUID |
| name | varchar(30) | NO | UNI |  |  |  |
| questionnaire_id | varchar(36) | NO | MUL | NULL |  |  |
| start_period | timestamp | NO |  | 0 |  | 含めるゲームの基準時間の範囲(始点) |
| end_period | timestamp | NO |  | NULL |  | 含めるゲームの基準時間の範囲(終点) |
| start_time | timestamp | NO |  | 0 |  | 配信開始時刻 |
| created_at | timestamp | NO |  | CURRENT_TIMESTAMP |  | バージョンの作られた時刻 |
| updated_at | timestamp |  |  | CURRENT_TIMESTAMP |  | 最終更新時刻 |
| deleted_at | timestamp |  |  | NULL |  | 削除された時刻 |

### seat
| Name | Type | Null | Key | Default | Extra | 説明 |
| --- | --- | --- | --- | --- | --- | --- |
| id | varchar(36) | NO | PRI |  |  | UUID |
| seat_id | varchar(36) | NO |  |  |  | 席のid |
| created_at | timestamp | NO |  | CURRENT_TIMESTAMP |  | 着席時刻 |
| deleted_at | timestamp |  |  | NULL |  | 離席時刻 |

### play_time
| Name | Type | Null | Key | Default | Extra | 説明 |
| --- | --- | --- | --- | --- | --- | --- |
| id | varchar(36) | NO | PRI |  |  | UUID |
| version_id | varchar(36) | NO | MUL |  |  | プレイされたバージョンのid |
| game_id | varchar(36) | NO | MUL |  |  | プレイされたゲームのid |
| start_time | timestamp | NO |  |  |  | プレイ開始時刻 |
| end_time | timestamp | NO |  |  |  | プレイ終了時刻 |

### special
| Name | Type | Null | Key | Default | Extra | 説明 |
| --- | --- | --- | --- | --- | --- | --- |
| id | varchar(36) | NO | PRI |  |  |  |
| version_id | varchar(36) | NO | MUL |  |  | 特例を指定するバージョンのid |
| game_id | varchar(36) | NO | MUL |  |  | 特例に指定するゲームのid |
| status | text | NO |  | "in" |  | 状態の種類("in":特例として入れる、"out":特例として外す) |
| deleted_at | timestamp |  |  | NULL |  | 削除時刻 |

### administrators
| Name | Type | Null | Key | Default | Extra | 説明 |
| --- | --- | --- | --- | --- | --- | --- |
| user_traqid | char(30) | NO | PRI |  |  |  |

### options
| Name | Type | Null | Key | Default | Extra | 説明 |
| --- | --- | --- | --- | --- | --- | --- |
| id | varchar(36) | NO | PRI |  |  | UUID |
| question_id | varchar(36) | NO | MUL |  |  |  |
| option_num | int(11) | NO |  |  |  |  |
| body | text |  |  |  |  | 選択肢の内容 |

### question
| Name | Type | Null | Key | Default | Extra | 説明 |
| --- | --- | --- | --- | --- | --- | --- |
| id | varchar(36) | NO | PRI |  |  | UUID |
| questionnaire_id | varchar(36) | MUL |  |  |  | どのアンケートの質問か |
| page_num | int(11) | NO |  |  |  | アンケートの何ページ目の質問か |
| question_num | int(11) | NO |  |  |  | アンケートの質問のうち、何問目か |
| type | char(20) | NO |  |  |  | どのタイプの質問か ("Text", "Number", "MultipleChoice", "Checkbox", "Dropdown", "LinearScale", "Date", "Time") |
| body | text |  |  |  |  | 質問の内容 |
| is_required | tinyint(4) | NO |  | 0 |  | 回答が必須である (1) , ない(0) |
| deleted_at | timestamp |  |  | NULL |  | 質問が削除された日時 (削除されていない場合は NULL) |
| created_at | timestamp | NO |  | CURRENT_TIMESTAMP |  | 質問が作成された日時 |

### questionnaires
| Name | Type | Null | Key | Default | Extra | 説明 |
| --- | --- | --- | --- | --- | --- | --- |
| id | varchar(36) | NO | PRI |  |  | UUID |
| title | char(50) | NO | UNI |  |  | アンケートのタイトル |
| description | text | NO |  |  |  | アンケートの説明 |
| res_time_limit | timestamp |  |  | NULL |  | 回答の締切日時 (締切がない場合は NULL) |
| deleted_at | timestamp |  |  | NULL |  | アンケートが削除された日時 (削除されていない場合は NULL) |
| created_at | timestamp | NO |  | CURRENT_TIMESTAMP |  | アンケートが作成された日時 |
| modified_at    | timestamp | NO |  | CURRENT_TIMESTAMP |  | アンケートが更新された日時 |

### respondents
| Name | Type | Null | Key | Default | Extra | 説明 |
| --- | --- | --- | --- | --- | --- | --- |
| response_id | varchar(36) | NO | PRI |  |  | 一つのアンケートに対する一つの回答ごとに振られる ID |
| questionnaire_id | varchar(36) | NO | MUL |  |  | どのアンケートへの回答か |
| modified_at | timestamp |  |  | NULL |  | 回答が変更された日時 |
| submitted_at | timestamp | NO |  | NULL |  | 回答が送信された日時 (未送信の場合は NULL) |
| deleted_at | timestamp |  |  | NULL |  | 回答が破棄された日時 (破棄されていない場合は NULL) |

### response
| Name | Type | Null | Key | Default | Extra | 説明 |
| --- | --- | --- | --- | --- | --- | --- |
| response_id | varchar(36) | NO | MUL |  |  | 一つのアンケートに対する一つの回答ごとに振られる ID |
| question_id | varchar(36) | NO | MUL |  |  | どの質問への回答か |
| body | text |  |  |  |  | 回答の内容 |
| modified_at | timestamp | NO |  | NULL |  | 回答が変更された日時 |
| deleted_at  | timestamp |  |  | NULL |  | 回答が破棄された日時 (破棄されていない場合は NULL) |

### scale_labels
| Name | Type | Null | Key | Default | Extra | 説明 |
| --- | --- | --- | --- | --- | --- | --- |
| question_id | varchar(36) | NO | PRI | NULL |  | どの質問のラベルか |
| scale_label_left | text |  |  | NULL |  | 左側のラベル (ない場合は NULL) |
| scale_label_right | text |  |  | NULL |  | 右側のラベル (ない場合は NULL) |
| scale_min | int(11) |  |  | NULL |  | スケールの最小値 |
| scale_max | int(11) |  |  | NULL |  | スケールの最大値 |