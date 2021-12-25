package gorm2

import (
	"context"
	"fmt"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"gorm.io/gorm"
	gormPrometheus "gorm.io/plugin/prometheus"
)

type MetricsCollector struct {
	Prefix           string
	Interval         uint32
	accessTokenGauge *prometheus.GaugeVec
	gameGauge        *prometheus.GaugeVec
	gameImageGauge   *prometheus.GaugeVec
	gameVideoGauge   *prometheus.GaugeVec
	gameFileGauge    *prometheus.GaugeVec
	gameURLGauge     prometheus.Gauge
}

func (mc *MetricsCollector) Metrics(p *gormPrometheus.Prometheus) []prometheus.Collector {
	if mc.Prefix == "" {
		mc.Prefix = "gorm_trap_collection"
	}

	if mc.Interval == 0 {
		mc.Interval = p.RefreshInterval
	}

	if mc.accessTokenGauge == nil {
		mc.accessTokenGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: mc.Prefix,
			Subsystem: "access_token",
			Name:      "count",
			Help:      "Number of access tokens",
		}, []string{"status"})
	}

	if mc.gameGauge == nil {
		mc.gameGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: mc.Prefix,
			Subsystem: "game",
			Name:      "count",
			Help:      "Number of games",
		}, []string{"status"})
	}

	if mc.gameImageGauge == nil {
		mc.gameImageGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: mc.Prefix,
			Subsystem: "game_image",
			Name:      "count",
			Help:      "Number of game images",
		}, []string{"type"})
	}

	if mc.gameVideoGauge == nil {
		mc.gameVideoGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: mc.Prefix,
			Subsystem: "game_video",
			Name:      "count",
			Help:      "Number of game videos",
		}, []string{"type"})
	}

	if mc.gameFileGauge == nil {
		mc.gameFileGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: mc.Prefix,
			Subsystem: "game_file",
			Name:      "count",
			Help:      "Number of game files",
		}, []string{"type"})
	}

	if mc.gameURLGauge == nil {
		mc.gameURLGauge = promauto.NewGauge(prometheus.GaugeOpts{
			Namespace: mc.Prefix,
			Subsystem: "game_url",
			Name:      "count",
			Help:      "Number of game urls",
		})
	}

	go func() {
		for range time.Tick(time.Duration(mc.Interval) * time.Second) {
			mc.collect(p)
		}
	}()

	mc.collect(p)

	return []prometheus.Collector{
		mc.accessTokenGauge,
		mc.gameGauge,
		mc.gameImageGauge,
		mc.gameVideoGauge,
		mc.gameFileGauge,
		mc.gameURLGauge,
	}
}

func (mc *MetricsCollector) collect(p *gormPrometheus.Prometheus) {
	ctx := context.Background()

	err := mc.collectAccessTokenMetrics(ctx, p)
	if err != nil {
		p.DB.Logger.Error(ctx, "failed to collect access token metrics", err)
	}

	err = mc.collectGameMetrics(ctx, p)
	if err != nil {
		p.DB.Logger.Error(ctx, "failed to collect game metrics", err)
	}

	err = mc.collectGameImageMetrics(ctx, p)
	if err != nil {
		p.DB.Logger.Error(ctx, "failed to collect game image metrics", err)
	}

	err = mc.collectGameVideoMetrics(ctx, p)
	if err != nil {
		p.DB.Logger.Error(ctx, "failed to collect game video metrics", err)
	}

	err = mc.collectGameFileMetrics(ctx, p)
	if err != nil {
		p.DB.Logger.Error(ctx, "failed to collect game file metrics", err)
	}

	err = mc.collectGameURLMetrics(ctx, p)
	if err != nil {
		p.DB.Logger.Error(ctx, "failed to collect game url metrics", err)
	}
}

func (mc *MetricsCollector) collectAccessTokenMetrics(ctx context.Context, p *gormPrometheus.Prometheus) error {
	var accessTokenCounts []struct {
		IsDeleted bool  `gorm:"column:is_deleted"`
		Count     int64 `gorm:"column:count"`
	}

	err := p.DB.
		Session(&gorm.Session{}).
		Unscoped().
		Model(&LauncherSessionTable{}).
		Select("deleted_at IS NOT NULL OR expires_at < ? AS is_deleted, count(*) as count", time.Now().Add(-24*time.Hour)).
		Group("is_deleted").
		Find(&accessTokenCounts).Error
	if err != nil {
		return fmt.Errorf("failed to get access token counts: %w", err)
	}

	mc.accessTokenGauge.Reset()
	for _, count := range accessTokenCounts {
		var label string
		if count.IsDeleted {
			label = "deleted"
		} else {
			label = "active"
		}

		mc.accessTokenGauge.
			WithLabelValues(label).
			Set(float64(count.Count))
	}

	return nil
}

func (mc *MetricsCollector) collectGameMetrics(ctx context.Context, p *gormPrometheus.Prometheus) error {
	var gameCounts []struct {
		IsDeleted bool  `gorm:"column:is_deleted"`
		Count     int64 `gorm:"column:count"`
	}

	err := p.DB.
		Session(&gorm.Session{}).
		Unscoped().
		Model(&GameTable{}).
		Select("deleted_at IS NOT NULL AS is_deleted, count(*) as count").
		Group("is_deleted").
		Find(&gameCounts).Error
	if err != nil {
		return fmt.Errorf("failed to get game counts: %w", err)
	}

	mc.gameGauge.Reset()
	for _, count := range gameCounts {
		var label string
		if count.IsDeleted {
			label = "deleted"
		} else {
			label = "active"
		}

		mc.gameGauge.
			WithLabelValues(label).
			Set(float64(count.Count))
	}

	return nil
}

func (mc *MetricsCollector) collectGameFileMetrics(ctx context.Context, p *gormPrometheus.Prometheus) error {
	var gameFileCounts []struct {
		Type  string `gorm:"column:type"`
		Count int64  `gorm:"column:count"`
	}

	err := p.DB.
		Session(&gorm.Session{}).
		Unscoped().
		Model(&GameFileTable{}).
		Joins("JOIN game_file_types ON game_files.file_type_id = game_file_types.id AND game_file_types.active").
		Select("game_file_types.name AS type, count(*) as count").
		Group("type").
		Find(&gameFileCounts).Error
	if err != nil {
		return fmt.Errorf("failed to get game file counts: %w", err)
	}

	mc.gameFileGauge.Reset()
	for _, count := range gameFileCounts {
		mc.gameFileGauge.
			WithLabelValues(count.Type).
			Set(float64(count.Count))
	}

	return nil
}

func (mc *MetricsCollector) collectGameURLMetrics(ctx context.Context, p *gormPrometheus.Prometheus) error {
	var gameURLCount int64

	err := p.DB.
		Session(&gorm.Session{}).
		Unscoped().
		Model(&GameURLTable{}).
		Count(&gameURLCount).Error
	if err != nil {
		return fmt.Errorf("failed to get game url counts: %w", err)
	}

	mc.gameURLGauge.Set(float64(gameURLCount))

	return nil
}

func (mc *MetricsCollector) collectGameImageMetrics(ctx context.Context, p *gormPrometheus.Prometheus) error {
	var gameImageCounts []struct {
		Type  string `gorm:"column:type"`
		Count int64  `gorm:"column:count"`
	}

	err := p.DB.
		Session(&gorm.Session{}).
		Unscoped().
		Model(&GameImageTable{}).
		Joins("JOIN game_image_types ON game_images.image_type_id = game_image_types.id AND game_image_types.active").
		Select("game_image_types.name AS type, count(*) as count").
		Group("type").
		Find(&gameImageCounts).Error
	if err != nil {
		return fmt.Errorf("failed to get game image counts: %w", err)
	}

	mc.gameImageGauge.Reset()
	for _, count := range gameImageCounts {
		mc.gameImageGauge.
			WithLabelValues(count.Type).
			Set(float64(count.Count))
	}

	return nil
}

func (mc *MetricsCollector) collectGameVideoMetrics(ctx context.Context, p *gormPrometheus.Prometheus) error {
	var gameVideoCounts []struct {
		Type  string `gorm:"column:type"`
		Count int64  `gorm:"column:count"`
	}

	err := p.DB.
		Session(&gorm.Session{}).
		Unscoped().
		Model(&GameVideoTable{}).
		Joins("JOIN game_video_types ON game_videos.video_type_id = game_video_types.id AND game_video_types.active").
		Select("game_video_types.name AS type, count(*) as count").
		Group("type").
		Find(&gameVideoCounts).Error
	if err != nil {
		return fmt.Errorf("failed to get game video counts: %w", err)
	}

	mc.gameVideoGauge.Reset()
	for _, count := range gameVideoCounts {
		mc.gameVideoGauge.
			WithLabelValues(count.Type).
			Set(float64(count.Count))
	}

	return nil
}
