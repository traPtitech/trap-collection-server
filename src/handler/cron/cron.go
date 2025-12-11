package cron

import (
	"context"
	"log"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/traPtitech/trap-collection-server/src/service"
)

// Cron 定期実行ジョブを管理する構造体
type Cron struct {
	deletePlayLogService service.GamePlayLogV2
	scheduler            *cron.Cron
}

func NewCron(deletePlayLogService service.GamePlayLogV2) *Cron {
	return &Cron{
		deletePlayLogService: deletePlayLogService,
	}
}

func (c *Cron) Start() error {
	c.scheduler = cron.New()

	_, err := c.scheduler.AddFunc("@hourly", c.deleteLongLogs)
	if err != nil {
		return err
	}

	c.scheduler.Start()
	return nil
}

func (c *Cron) Stop() {
	if c.scheduler != nil {
		ctx := c.scheduler.Stop()
		<-ctx.Done()
		log.Println("Cron: 停止完了")
	}
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
