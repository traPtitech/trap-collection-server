package cron

import (
	"context"
	"log"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/traPtitech/trap-collection-server/src/service"
)

// 定期実行コードの構造体
type Cron struct {
	deletePlayLogService service.GamePlayLogV2
}

func NewCron(deletePlayLogService service.GamePlayLogV2) *Cron {
	return &Cron{
		deletePlayLogService: deletePlayLogService,
	}
}

func (c *Cron) Start() error {
	scheduler := cron.New()

	_, err := scheduler.AddFunc("@hourly", c.deleteLongLogs)
	if err != nil {
		return err
	}

	scheduler.Start()
	return nil
}

func (c *Cron) deleteLongLogs() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	log.Println("DeleteLongLogs: 開始")
	deletedIDs, err := c.deletePlayLogService.DeleteLongLogs(ctx)
	if err != nil {
		log.Printf("DeleteLongLogs: エラー: %v\n", err)
		return
	}
	log.Printf("DeleteLongLogs: 終了 削除件数: %d\n", len(deletedIDs))
	for i, id := range deletedIDs {
		log.Printf("  [%d] %s\n", i+1, id)
	}
}
