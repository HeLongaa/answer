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

func addAIConversationBranches(ctx context.Context, x *xorm.Engine) error {
	if err := x.Context(ctx).Sync(new(entity.AIConversationRecord)); err != nil {
		return fmt.Errorf("sync ai conversation branch fields failed: %w", err)
	}

	records := make([]*entity.AIConversationRecord, 0)
	if err := x.Context(ctx).Asc("conversation_id", "id").Find(&records); err != nil {
		return fmt.Errorf("list ai conversation records failed: %w", err)
	}

	lastUserByConversation := map[string]string{}
	branchCountByParent := map[string]int{}
	for _, record := range records {
		if record.MessageID != "" {
			if record.Role == "user" {
				lastUserByConversation[record.ConversationID] = record.MessageID
			}
			continue
		}

		record.MessageID = fmt.Sprintf("legacy-%d", record.ID)
		record.Active = true
		record.BranchIndex = 0
		if record.Role == "assistant" {
			parentID := lastUserByConversation[record.ConversationID]
			record.ParentMessageID = parentID
			record.BranchIndex = branchCountByParent[parentID]
			branchCountByParent[parentID]++
		}
		if record.Role == "user" {
			lastUserByConversation[record.ConversationID] = record.MessageID
		}

		if _, err := x.Context(ctx).ID(record.ID).
			Cols("message_id", "parent_message_id", "branch_index", "active").
			Update(record); err != nil {
			return fmt.Errorf("backfill ai conversation record branch fields failed: %w", err)
		}
	}
	return nil
}
