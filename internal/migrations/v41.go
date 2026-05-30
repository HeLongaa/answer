/*
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements.  See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership.  The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License.  You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

package migrations

import (
	"context"
	"fmt"

	"github.com/apache/answer/internal/entity"
	"xorm.io/xorm"
)

func addAIImageGeneration(ctx context.Context, x *xorm.Engine) error {
	if err := x.Context(ctx).Sync(
		new(entity.AIImageProvider),
		new(entity.AIImageModel),
		new(entity.AIImageSetting),
		new(entity.AIImageGeneration),
		new(entity.AIVideoProvider),
		new(entity.AIVideoModel),
		new(entity.AIVideoSetting),
		new(entity.AIVideoGeneration),
		new(entity.AISubscriptionPlan),
	); err != nil {
		return fmt.Errorf("sync ai image and video generation tables failed: %w", err)
	}

	setting := &entity.AIImageSetting{ID: 1, RetentionDays: 30}
	exist, err := x.Context(ctx).ID(1).Exist(new(entity.AIImageSetting))
	if err != nil {
		return fmt.Errorf("check ai image setting failed: %w", err)
	}
	if !exist {
		if _, err := x.Context(ctx).Insert(setting); err != nil {
			return fmt.Errorf("insert ai image setting failed: %w", err)
		}
	}
	videoSetting := &entity.AIVideoSetting{ID: 1, RetentionDays: 30}
	videoExist, err := x.Context(ctx).ID(1).Exist(new(entity.AIVideoSetting))
	if err != nil {
		return fmt.Errorf("check ai video setting failed: %w", err)
	}
	if !videoExist {
		if _, err := x.Context(ctx).Insert(videoSetting); err != nil {
			return fmt.Errorf("insert ai video setting failed: %w", err)
		}
	}
	return nil
}
