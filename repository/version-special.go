package repository

import (
	"fmt"
	"time"

	"github.com/traPtitech/trap-collection-server/model"
)

//GetGameListByVersion バージョンによるゲームの取得
func GetGameListByVersion(version string) ([]model.GameInfo,error) {
	type Period struct {
		StartPeriod time.Time `db:"start_period"`
		EndPeriod time.Time `db:"end_period"`
	}
	periodList := []Period{}
	err := Db.Select(&periodList,"(SELECT start_period,end_period FROM version_for_sale WHERE name=? AND deleted_at IS NULL) UNION (SELECT start_period,end_period FROM version_not_for_sale WHERE name=? AND deleted_at IS NULL)",version,version)
	if err!=nil {
		return nil,err
	}
	if periodList==nil {
		return nil,fmt.Errorf("Error: %s", "this version does not exist")
	}
	period := periodList[0]

	mainGameList := []model.GameInfo{}
	err = Db.Select(&mainGameList,"SELECT name,time FROM game WHERE time>? AND time<?",period.StartPeriod,period.EndPeriod)
	if err!=nil {
		return nil,err
	}
	mainGameListLength := len(mainGameList)

	inGameList := []model.GameInfo{}
	err = Db.Select(&inGameList,"SELECT game.name AS name,game.time AS time FROM special INNER JOIN game ON special.name=game.name WHERE special.version_name=? AND special.inout=in",version)
	if err!=nil {
		return nil,err
	}
	inGameListLength :=len(inGameList)

	outGameList := []model.GameInfo{}
	err = Db.Select(&outGameList,"SELECT game.name AS name,game.time AS time FROM special INNER JOIN game ON special.name=game.name WHERE special.version_name=? AND special.inout=out",version)
	if err!=nil {
		return nil,err
	}

	gameList := make([]model.GameInfo,0,mainGameListLength+inGameListLength)
	for _,v := range inGameList {
		gameList = append(gameList,v)
	}
	for _,mainv := range mainGameList {
		b := true
		for _,outv := range outGameList {
			if mainv==outv {
				b = false
				break
			}
		}
		if b {
			gameList = append(gameList,mainv)
		}
	}
	return gameList,nil
}

//GetQuestionnaireByVersionNotForSale バージョンによるアンケートの取得
func GetQuestionnaireByVersionNotForSale(version string) (int,error) {
	questionnair := -1
	err := Db.Get(&questionnair,"SELECT questionnaire_id FROM version_not_for_sale WHERE name=? AND deleted_at IS NULL",version,version)
	if err!=nil {
		return -1,err
	}
	if questionnair==-1 {
		return -1,fmt.Errorf("Error: %s", "this version does not exist")
	}
	return questionnair,nil
}

//GameCheckListByVersion バージョンによるチェック用のゲーム一覧の取得
func GameCheckListByVersion(version string) ([]model.GameCheck,error) {
	type Period struct {
		StartPeriod time.Time `db:"start_period"`
		EndPeriod time.Time `db:"end_period"`
	}
	periodList := []Period{}
	err := Db.Select(&periodList,"(SELECT start_period,end_period FROM version_for_sale WHERE name=? AND deleted_at IS NULL) UNION (SELECT start_period,end_period FROM version_not_for_sale WHERE name=? AND deleted_at IS NULL)",version,version)
	if err!=nil {
		return nil,err
	}
	if periodList==nil {
		return nil,fmt.Errorf("Error: %s", "this version does not exist")
	}
	period := periodList[0]

	mainGameList := []model.GameCheck{}
	err = Db.Select(&mainGameList,"SELECT id,name,md5 FROM game WHERE time>? AND time<?",period.StartPeriod,period.EndPeriod)
	if err!=nil {
		return nil,err
	}
	mainGameListLength := len(mainGameList)

	inGameList := []model.GameCheck{}
	err = Db.Select(&inGameList,"SELECT game.id AS id,game.name AS name,game.md5 AS md5 FROM special INNER JOIN game ON special.name=game.name WHERE special.version_name=? AND special.inout=in",version)
	if err!=nil {
		return nil,err
	}
	inGameListLength :=len(inGameList)

	outGameList := []model.GameCheck{}
	err = Db.Select(&outGameList,"SELECT game.id AS id,game.name AS name,game.md5 AS md5 FROM special INNER JOIN game ON special.name=game.name WHERE special.version_name=? AND special.inout=out",version)
	if err!=nil {
		return nil,err
	}

	gameList := make([]model.GameCheck,0,mainGameListLength+inGameListLength)
	for _,v := range inGameList {
		gameList = append(gameList,v)
	}
	for _,mainv := range mainGameList {
		b := true
		for _,outv := range outGameList {
			if mainv==outv {
				b = false
				break
			}
		}
		if b {
			gameList = append(gameList,mainv)
		}
	}
	return gameList,nil
}