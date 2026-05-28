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

func backfillFeaturedPosts(ctx context.Context, x *xorm.Engine) error {
	transactions := make([]*entity.PointTransaction, 0)
	if err := x.Context(ctx).
		Where("source_type = ?", entity.PointSourceFeaturedPostReward).
		Find(&transactions); err != nil {
		return fmt.Errorf("list featured post reward transactions failed: %w", err)
	}

	for _, transaction := range transactions {
		exist, err := x.Context(ctx).
			Where("question_id = ?", transaction.SourceID).
			Exist(new(entity.FeaturedPost))
		if err != nil {
			return fmt.Errorf("check featured post %s failed: %w", transaction.SourceID, err)
		}
		if exist {
			continue
		}

		question := &entity.Question{ID: transaction.SourceID}
		hasQuestion, err := x.Context(ctx).Get(question)
		if err != nil {
			return fmt.Errorf("get featured question %s failed: %w", transaction.SourceID, err)
		}
		if !hasQuestion {
			continue
		}

		featured := &entity.FeaturedPost{
			QuestionID:   transaction.SourceID,
			AuthorID:     question.UserID,
			OperatorID:   transaction.OperatorID,
			Title:        question.Title,
			RewardPoints: transaction.Delta,
			Active:       true,
			Revoked:      false,
		}
		if _, err = x.Context(ctx).Insert(featured); err != nil {
			return fmt.Errorf("backfill featured post %s failed: %w", transaction.SourceID, err)
		}
	}
	return nil
}
