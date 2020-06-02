package model

import (
	"errors"
	"fmt"
	"log"

	"github.com/jinzhu/gorm"
	gormbulk "github.com/t-tiger/gorm-bulk-insert"
	"github.com/traPtitech/trap-collection-server/openapi"
)

// InsertResponses 回答の追加
func InsertResponses(productKey string, res *openapi.NewResponse) (*openapi.NewResponse, error) {
	playerID, err := getPlayerIDByProductKey(productKey)
	if err != nil {
		return &openapi.NewResponse{}, fmt.Errorf("Failed In Getting PlayerID:%w", err)
	}
	err = db.Transaction(func(tx *gorm.DB) error {
		response := Response{
			ID: res.Id,
			PlayerID: playerID,
			Remark: res.Remark,
		}
		err = tx.Create(&response).Error
		if err != nil {
			return fmt.Errorf("Failed In Creating Response:%w", err)
		}
		questionIDs := make([]int32, 0, len(res.Answers))
		answerMap := make(map[uint]*openapi.Answer)
		for i,v := range res.Answers {
			questionIDs[i] = v.Id
			answerMap[uint(v.Id)] = &v
		}
		rows,err := tx.Table("questions").
			Select("id","type").
			Where("id IN ?", questionIDs).
			Rows()
		if err != nil {
			return fmt.Errorf("Failed In Getting Question Type:%w", err)
		}
		textAnswers := make([]interface{}, 0, len(res.Answers))
		optionAnswers := make([]interface{}, 0, len(res.Answers))
		for rows.Next() {
			var id uint
			var questionType uint8
			err = rows.Scan(&id, &questionType)
			if err != nil {
				return fmt.Errorf("Failed In Scaning Question Type:%w", err)
			}
			answer := answerMap[id]
			qType, ok := typeMap[questionType]
			if !ok {
				log.Println("error: unexpected invalid question type")
				return errors.New("Invalid Question Type")
			}
			switch qType {
				case "text":
					textAnswer := TextAnswer{
						ResponseID: res.Id,
						QuestionID: id,
						Content: answer.Contents.Text,
					}
					textAnswers = append(textAnswers, textAnswer)
				case "radio":
					if len(answer.Contents.Options) != 1 {
						return errors.New("Invalid Contents Length")
					}
					optionAnswer := OptionAnswer{
						ResponseID: res.Id,
						QuestionID: id,
						OptionID: uint(answer.Contents.Options[0]),
					}
					optionAnswers = append(optionAnswers, optionAnswer)
				case "checkbox":
					optionAnswer := OptionAnswer{
						ResponseID: res.Id,
						QuestionID: id,
						OptionID: uint(answer.Contents.Options[0]),
					}
					optionAnswers = append(optionAnswers, optionAnswer)
			}
		}
		err = gormbulk.BulkInsert(tx, textAnswers, 3000)
		if err != nil {
			return fmt.Errorf("Failed In Inserting Text Answer: %w", err)
		}
		err = gormbulk.BulkInsert(tx, optionAnswers, 3000)
		if err != nil {
			return fmt.Errorf("Failed In Inserting Option Answer: %w", err)
		}
		return nil
	})
	if err != nil {
		return &openapi.NewResponse{}, fmt.Errorf("Failed In Transaction: %w", err)
	}
	return res, nil
}
