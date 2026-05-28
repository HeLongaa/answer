package migrations

import (
	"context"
	"fmt"

	"github.com/apache/answer/internal/entity"
	"xorm.io/xorm"
)

func addTaskSquareAndPoints(ctx context.Context, x *xorm.Engine) error {
	if err := x.Context(ctx).Sync(
		new(entity.Task),
		new(entity.TaskSubmission),
		new(entity.UserPointAccount),
		new(entity.PointTransaction),
		new(entity.FeaturedPost),
	); err != nil {
		return fmt.Errorf("sync task square and points tables failed: %w", err)
	}
	return nil
}
