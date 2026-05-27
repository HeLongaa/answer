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

func addAIChatConfig(ctx context.Context, x *xorm.Engine) error {
	if err := x.Context(ctx).Sync(
		new(entity.AIProvider),
		new(entity.AIProviderModel),
		new(entity.AIModelMapping),
		new(entity.AIModelMappingItem),
		new(entity.AISubscriptionPlan),
		new(entity.AISubscriptionPlanModel),
		new(entity.AISubscriptionRedeemCode),
		new(entity.AIModelConsumeRate),
		new(entity.User),
	); err != nil {
		return fmt.Errorf("sync ai chat config tables failed: %w", err)
	}

	free := &entity.AISubscriptionPlan{
		PlanID:          "free",
		Name:            "FREE",
		Enabled:         true,
		MonthlyPrice:    0,
		ChatPoints:      0,
		ImageQuota:      0,
		TaskDescription: "Default free plan",
		SortOrder:       0,
	}
	exist, err := x.Context(ctx).Where("plan_id = ?", free.PlanID).Exist(new(entity.AISubscriptionPlan))
	if err != nil {
		return fmt.Errorf("check free plan failed: %w", err)
	}
	if !exist {
		if _, err := x.Context(ctx).Insert(free); err != nil {
			return fmt.Errorf("insert free plan failed: %w", err)
		}
	}

	if _, err := x.Context(ctx).Where("subscription_level = '' OR subscription_level IS NULL").
		Cols("subscription_level").Update(&entity.User{SubscriptionLevel: "free"}); err != nil {
		return fmt.Errorf("set user default subscription level failed: %w", err)
	}
	return nil
}
