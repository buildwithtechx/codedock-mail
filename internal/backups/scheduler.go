package backups

import (
	"context"
	"fmt"

	"github.com/robfig/cron/v3"

	"github.com/buildwithtechx/codedock-mail/internal/models"
)

type Scheduler struct {
	runner *cron.Cron
}

func NewScheduler(spec string, retention int, dsn, dir, mailDataDir string, onBackup func(context.Context, models.BackupManifest, error)) (*Scheduler, error) {
	runner := cron.New()
	_, err := runner.AddFunc(spec, func() {
		ctx := context.Background()
		manifest, err := Create(ctx, dsn, dir, mailDataDir)
		if err == nil && retention > 0 {
			err = Prune(dir, retention)
		}
		if onBackup != nil {
			onBackup(ctx, manifest, err)
		}
	})
	if err != nil {
		return nil, fmt.Errorf("parse backup schedule: %w", err)
	}
	return &Scheduler{runner: runner}, nil
}

func (s *Scheduler) Start() {
	s.runner.Start()
}

func (s *Scheduler) Stop() {
	<-s.runner.Stop().Done()
}

func Prune(dir string, keep int) error {
	if keep < 0 {
		keep = 0
	}
	manifests, err := List(dir)
	if err != nil {
		return err
	}
	if len(manifests) <= keep {
		return nil
	}
	for _, stale := range manifests[keep:] {
		if err := Delete(dir, stale.ID); err != nil {
			return err
		}
	}
	return nil
}
