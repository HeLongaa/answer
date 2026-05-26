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

package constant

import "github.com/apache/answer/internal/base/reason"

type Privilege struct {
	Key   string `json:"key"`
	Label string `json:"label"`
	Value int    `validate:"gte=1" json:"value"`
}

const (
	RankQuestionAddKey               = "rank.question.add"
	RankQuestionEditKey              = "rank.question.edit"
	RankQuestionDeleteKey            = "rank.question.delete"
	RankQuestionVoteUpKey            = "rank.question.vote_up"
	RankQuestionVoteDownKey          = "rank.question.vote_down"
	RankAnswerAddKey                 = "rank.answer.add"
	RankAnswerEditKey                = "rank.answer.edit"
	RankAnswerDeleteKey              = "rank.answer.delete"
	RankAnswerAcceptKey              = "rank.answer.accept"
	RankAnswerVoteUpKey              = "rank.answer.vote_up"
	RankAnswerVoteDownKey            = "rank.answer.vote_down"
	RankInviteSomeoneToAnswerKey     = "rank.answer.invite_someone_to_answer"
	RankCommentAddKey                = "rank.comment.add"
	RankCommentEditKey               = "rank.comment.edit"
	RankCommentDeleteKey             = "rank.comment.delete"
	RankReportAddKey                 = "rank.report.add"
	RankTagAddKey                    = "rank.tag.add"
	RankTagEditKey                   = "rank.tag.edit"
	RankTagDeleteKey                 = "rank.tag.delete"
	RankTagSynonymKey                = "rank.tag.synonym"
	RankLinkUrlLimitKey              = "rank.link.url_limit"
	RankVoteDetailKey                = "rank.vote.detail"
	RankCommentVoteUpKey             = "rank.comment.vote_up"
	RankCommentVoteDownKey           = "rank.comment.vote_down"
	RankQuestionEditWithoutReviewKey = "rank.question.edit_without_review"
	RankAnswerEditWithoutReviewKey   = "rank.answer.edit_without_review"
	RankTagEditWithoutReviewKey      = "rank.tag.edit_without_review"
	RankAnswerAuditKey               = "rank.answer.audit"
	RankQuestionAuditKey             = "rank.question.audit"
	RankTagAuditKey                  = "rank.tag.audit"
	RankQuestionCloseKey             = "rank.question.close"
	RankQuestionReopenKey            = "rank.question.reopen"
	RankTagUseReservedTagKey         = "rank.tag.use_reserved_tag"
)

var (
	RankRoleOnlyPrivilegeKeys = map[string]bool{
		RankTagAddKey:                    true,
		RankTagEditKey:                   true,
		RankQuestionEditKey:              true,
		RankQuestionEditWithoutReviewKey: true,
		RankAnswerEditWithoutReviewKey:   true,
		RankQuestionAuditKey:             true,
		RankAnswerAuditKey:               true,
		RankTagAuditKey:                  true,
		RankTagEditWithoutReviewKey:      true,
		RankTagSynonymKey:                true,
	}

	RankRoleOnlyPermissionKeys = map[string]bool{
		RankQuestionEditKey[len("rank."):]:              true,
		RankTagAddKey[len("rank."):]:                    true,
		RankTagEditKey[len("rank."):]:                   true,
		RankQuestionEditWithoutReviewKey[len("rank."):]: true,
		RankAnswerEditWithoutReviewKey[len("rank."):]:   true,
		RankQuestionAuditKey[len("rank."):]:             true,
		RankAnswerAuditKey[len("rank."):]:               true,
		RankTagAuditKey[len("rank."):]:                  true,
		RankTagEditWithoutReviewKey[len("rank."):]:      true,
		RankTagSynonymKey[len("rank."):]:                true,
	}

	RankAllPrivileges = []*Privilege{
		{Label: reason.RankQuestionAddLabel, Key: RankQuestionAddKey},
		{Label: reason.RankAnswerAddLabel, Key: RankAnswerAddKey},
		{Label: reason.RankCommentAddLabel, Key: RankCommentAddKey},
		{Label: reason.RankReportAddLabel, Key: RankReportAddKey},
		{Label: reason.RankCommentVoteUpLabel, Key: RankCommentVoteUpKey},
		{Label: reason.RankLinkUrlLimitLabel, Key: RankLinkUrlLimitKey},
		{Label: reason.RankQuestionVoteUpLabel, Key: RankQuestionVoteUpKey},
		{Label: reason.RankAnswerVoteUpLabel, Key: RankAnswerVoteUpKey},
		{Label: reason.RankQuestionVoteDownLabel, Key: RankQuestionVoteDownKey},
		{Label: reason.RankAnswerVoteDownLabel, Key: RankAnswerVoteDownKey},
		{Label: reason.RankInviteSomeoneToAnswerLabel, Key: RankInviteSomeoneToAnswerKey},
		{Label: reason.RankAnswerEditLabel, Key: RankAnswerEditKey},
	}
)
