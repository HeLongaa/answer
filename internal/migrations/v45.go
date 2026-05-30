package migrations

import (
	"context"
	"fmt"

	"github.com/apache/answer/internal/entity"
	"xorm.io/xorm"
)

func addAIVideoGeneration(ctx context.Context, x *xorm.Engine) error {
	if err := x.Context(ctx).Sync(
		new(entity.AISubscriptionPlan),
		new(entity.AIVideoProvider),
		new(entity.AIVideoModel),
		new(entity.AIVideoSetting),
		new(entity.AIVideoGeneration),
	); err != nil {
		return fmt.Errorf("sync ai video generation tables failed: %w", err)
	}

	setting := &entity.AIVideoSetting{ID: 1, RetentionDays: 30}
	exist, err := x.Context(ctx).ID(1).Exist(new(entity.AIVideoSetting))
	if err != nil {
		return fmt.Errorf("check ai video setting failed: %w", err)
	}
	if !exist {
		if _, err := x.Context(ctx).Insert(setting); err != nil {
			return fmt.Errorf("insert ai video setting failed: %w", err)
		}
	}
	return nil
}
