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

package ai_chat_config

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	stderrors "errors"
	"fmt"
	"math"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/apache/answer/internal/base/reason"
	"github.com/apache/answer/internal/entity"
	ai_chat_config_repo "github.com/apache/answer/internal/repo/ai_chat_config"
	"github.com/apache/answer/internal/schema"
	usercommon "github.com/apache/answer/internal/service/user_common"
	"github.com/go-resty/resty/v2"
	"github.com/segmentfault/pacman/errors"
)

var modelIDPattern = regexp.MustCompile(`^[a-z0-9_-]+$`)

type AiChatConfigService interface {
	ListProviders(ctx context.Context) ([]*schema.AIProviderResp, error)
	CreateProvider(ctx context.Context, req *schema.AIProviderReq) (*schema.AIProviderResp, error)
	UpdateProvider(ctx context.Context, id int, req *schema.AIProviderReq) (*schema.AIProviderResp, error)
	DeleteProvider(ctx context.Context, id int) error
	FetchProviderModels(ctx context.Context, providerID int) ([]*schema.AIProviderModelResp, error)
	TestProviderModel(ctx context.Context, providerID int, providerModelID string) (*schema.AITestProviderModelResp, error)
	GetProvider(ctx context.Context, providerID int) (*schema.AIProviderResp, error)
	GetProviderModels(ctx context.Context, providerID int) ([]*schema.AIProviderModelResp, error)

	ListModelMappings(ctx context.Context) ([]*schema.AIModelMappingResp, error)
	CreateModelMapping(ctx context.Context, req *schema.AIModelMappingReq) (*schema.AIModelMappingResp, error)
	UpdateModelMapping(ctx context.Context, id int, req *schema.AIModelMappingReq) (*schema.AIModelMappingResp, error)
	DeleteModelMapping(ctx context.Context, id int) error
	GetModelMapping(ctx context.Context, siteModelID string) (*schema.AIModelMappingResp, error)
	ResolveUpstreamModel(ctx context.Context, siteModelID string) (*schema.AIUpstreamModelResp, error)

	ListSubscriptionPlans(ctx context.Context) ([]*schema.AISubscriptionPlanResp, error)
	CreateSubscriptionPlan(ctx context.Context, req *schema.AISubscriptionPlanReq) (*schema.AISubscriptionPlanResp, error)
	UpdateSubscriptionPlan(ctx context.Context, id int, req *schema.AISubscriptionPlanReq) (*schema.AISubscriptionPlanResp, error)
	DeleteSubscriptionPlan(ctx context.Context, id int) error
	GetAvailableModelsForPlan(ctx context.Context, planID string) ([]string, error)
	ListUserAvailableModels(ctx context.Context, userID string) ([]*schema.AIChatModelResp, error)
	CheckUserModelPermission(ctx context.Context, userID, siteModelID string) error
	GetSubscriptionOverview(ctx context.Context, userID string) (*schema.AISubscriptionOverviewResp, error)
	GetSubscriptionPurchase(ctx context.Context, userID string) (*schema.AISubscriptionPurchaseResp, error)
	ListSubscriptionRedeemCodes(ctx context.Context) ([]*schema.AISubscriptionRedeemCodeResp, error)
	GenerateSubscriptionRedeemCodes(ctx context.Context, req *schema.AISubscriptionRedeemCodeGenerateReq) ([]*schema.AISubscriptionRedeemCodeResp, error)
	RedeemSubscriptionCode(ctx context.Context, userID string, req *schema.AISubscriptionRedeemReq) (*schema.AISubscriptionRedeemResp, error)

	ListConsumeRates(ctx context.Context) ([]*schema.AIModelConsumeRateResp, error)
	SaveConsumeRate(ctx context.Context, id int, req *schema.AIModelConsumeRateReq) (*schema.AIModelConsumeRateResp, error)
	GetModelConsumeRate(ctx context.Context, siteModelID string) (float64, error)
	CalculateChatCost(ctx context.Context, siteModelID string, baseCost float64) (float64, error)
	DeductUserPoints(ctx context.Context, userID string, cost float64) error
	RecordChatUsage(ctx context.Context, req *schema.AIChatUsageLogReq) error
}

type aiChatConfigService struct {
	repo     ai_chat_config_repo.AIChatConfigRepo
	userRepo usercommon.UserRepo
}

func NewAIChatConfigService(repo ai_chat_config_repo.AIChatConfigRepo, userRepo usercommon.UserRepo) AiChatConfigService {
	return &aiChatConfigService{repo: repo, userRepo: userRepo}
}

func (s *aiChatConfigService) ListProviders(ctx context.Context) ([]*schema.AIProviderResp, error) {
	providers, err := s.repo.ListProviders(ctx)
	if err != nil {
		return nil, errors.InternalServer(reason.DatabaseError).WithError(err)
	}
	resp := make([]*schema.AIProviderResp, 0, len(providers))
	for _, provider := range providers {
		models, _ := s.repo.ListProviderModels(ctx, provider.ID)
		resp = append(resp, s.formatProvider(provider, models, true))
	}
	return resp, nil
}

func (s *aiChatConfigService) GetProvider(ctx context.Context, providerID int) (*schema.AIProviderResp, error) {
	provider, exist, err := s.repo.GetProvider(ctx, providerID)
	if err != nil {
		return nil, errors.InternalServer(reason.DatabaseError).WithError(err)
	}
	if !exist {
		return nil, errors.BadRequest(reason.ObjectNotFound)
	}
	models, _ := s.repo.ListProviderModels(ctx, provider.ID)
	return s.formatProvider(provider, models, true), nil
}

func (s *aiChatConfigService) CreateProvider(ctx context.Context, req *schema.AIProviderReq) (*schema.AIProviderResp, error) {
	if strings.TrimSpace(req.APIKey) == "" {
		return nil, errors.BadRequest("api_key is required")
	}
	baseURL, err := normalizeBaseURL(req.BaseURL)
	if err != nil {
		return nil, errors.BadRequest("base_url is invalid")
	}
	provider := &entity.AIProvider{
		Name:           strings.TrimSpace(req.Name),
		BaseURL:        baseURL,
		APIKey:         strings.TrimSpace(req.APIKey),
		Enabled:        req.Enabled,
		SupportsStream: req.SupportsStream,
		Remark:         req.Remark,
	}
	if err := s.repo.CreateProvider(ctx, provider); err != nil {
		return nil, errors.InternalServer(reason.DatabaseError).WithError(err)
	}
	return s.formatProvider(provider, nil, true), nil
}

func (s *aiChatConfigService) UpdateProvider(ctx context.Context, id int, req *schema.AIProviderReq) (*schema.AIProviderResp, error) {
	provider, exist, err := s.repo.GetProvider(ctx, id)
	if err != nil {
		return nil, errors.InternalServer(reason.DatabaseError).WithError(err)
	}
	if !exist {
		return nil, errors.BadRequest(reason.ObjectNotFound)
	}
	baseURL, err := normalizeBaseURL(req.BaseURL)
	if err != nil {
		return nil, errors.BadRequest("base_url is invalid")
	}
	provider.Name = strings.TrimSpace(req.Name)
	provider.BaseURL = baseURL
	provider.Enabled = req.Enabled
	provider.SupportsStream = req.SupportsStream
	provider.Remark = req.Remark
	cols := []string{"name", "base_url", "enabled", "supports_stream", "remark"}
	if strings.TrimSpace(req.APIKey) != "" && !isAllMask(req.APIKey) {
		provider.APIKey = strings.TrimSpace(req.APIKey)
		cols = append(cols, "api_key")
	}
	if err := s.repo.UpdateProvider(ctx, provider, cols...); err != nil {
		return nil, errors.InternalServer(reason.DatabaseError).WithError(err)
	}
	models, _ := s.repo.ListProviderModels(ctx, provider.ID)
	return s.formatProvider(provider, models, true), nil
}

func (s *aiChatConfigService) DeleteProvider(ctx context.Context, id int) error {
	if err := s.repo.DeleteProvider(ctx, id); err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err)
	}
	return nil
}

func (s *aiChatConfigService) FetchProviderModels(ctx context.Context, providerID int) ([]*schema.AIProviderModelResp, error) {
	provider, exist, err := s.repo.GetProvider(ctx, providerID)
	if err != nil {
		return nil, errors.InternalServer(reason.DatabaseError).WithError(err)
	}
	if !exist {
		return nil, errors.BadRequest(reason.ObjectNotFound)
	}
	modelIDs, err := fetchOpenAIModels(provider.BaseURL, provider.APIKey)
	if err != nil {
		return nil, errors.BadRequest(fmt.Sprintf("failed to fetch models: %s", err.Error()))
	}
	now := time.Now()
	models := make([]*entity.AIProviderModel, 0, len(modelIDs))
	for _, modelID := range modelIDs {
		models = append(models, &entity.AIProviderModel{
			ProviderID:      providerID,
			ProviderModelID: modelID,
			ModelName:       modelID,
			Enabled:         true,
			FetchedAt:       now,
		})
	}
	if err := s.repo.ReplaceProviderModels(ctx, providerID, models); err != nil {
		return nil, errors.InternalServer(reason.DatabaseError).WithError(err)
	}
	return s.formatProviderModels(models), nil
}

func (s *aiChatConfigService) GetProviderModels(ctx context.Context, providerID int) ([]*schema.AIProviderModelResp, error) {
	models, err := s.repo.ListProviderModels(ctx, providerID)
	if err != nil {
		return nil, errors.InternalServer(reason.DatabaseError).WithError(err)
	}
	return s.formatProviderModels(models), nil
}

func (s *aiChatConfigService) TestProviderModel(ctx context.Context, providerID int, providerModelID string) (*schema.AITestProviderModelResp, error) {
	provider, exist, err := s.repo.GetProvider(ctx, providerID)
	if err != nil {
		return nil, errors.InternalServer(reason.DatabaseError).WithError(err)
	}
	if !exist {
		return nil, errors.BadRequest(reason.ObjectNotFound)
	}
	providerModelID = strings.TrimSpace(providerModelID)
	if providerModelID == "" {
		return nil, errors.BadRequest("provider_model_id is required")
	}
	models, err := s.repo.ListProviderModels(ctx, providerID)
	if err != nil {
		return nil, errors.InternalServer(reason.DatabaseError).WithError(err)
	}
	modelFound := false
	for _, model := range models {
		if model.ProviderModelID == providerModelID && model.Enabled {
			modelFound = true
			break
		}
	}
	if !modelFound {
		return nil, errors.BadRequest("provider model is not available")
	}
	message, raw, err := testOpenAIChat(provider.BaseURL, provider.APIKey, providerModelID)
	if err != nil {
		return nil, errors.BadRequest(fmt.Sprintf("failed to test model: %s", err.Error()))
	}
	return &schema.AITestProviderModelResp{
		ProviderID:      providerID,
		ProviderModelID: providerModelID,
		Message:         message,
		RawResponse:     raw,
	}, nil
}

func (s *aiChatConfigService) ListModelMappings(ctx context.Context) ([]*schema.AIModelMappingResp, error) {
	mappings, err := s.repo.ListModelMappings(ctx)
	if err != nil {
		return nil, errors.InternalServer(reason.DatabaseError).WithError(err)
	}
	resp := make([]*schema.AIModelMappingResp, 0, len(mappings))
	for _, mapping := range mappings {
		items, _ := s.repo.ListModelMappingItems(ctx, mapping.ID)
		resp = append(resp, s.formatMapping(ctx, mapping, items))
	}
	return resp, nil
}

func (s *aiChatConfigService) CreateModelMapping(ctx context.Context, req *schema.AIModelMappingReq) (*schema.AIModelMappingResp, error) {
	if err := s.validateMappingReq(req); err != nil {
		return nil, err
	}
	mapping := &entity.AIModelMapping{
		SiteModelID:            req.SiteModelID,
		DisplayName:            req.DisplayName,
		Description:            req.Description,
		Enabled:                req.Enabled,
		SortOrder:              req.SortOrder,
		SupportsVision:         req.SupportsVision,
		FallbackEnabled:        req.FallbackEnabled,
		DefaultProviderModelID: req.DefaultProviderModelID,
	}
	items := mappingItemsFromReq(req.Items)
	if err := s.repo.CreateModelMapping(ctx, mapping, items); err != nil {
		return nil, errors.InternalServer(reason.DatabaseError).WithError(err)
	}
	return s.formatMapping(ctx, mapping, items), nil
}

func (s *aiChatConfigService) UpdateModelMapping(ctx context.Context, id int, req *schema.AIModelMappingReq) (*schema.AIModelMappingResp, error) {
	if err := s.validateMappingReq(req); err != nil {
		return nil, err
	}
	mapping, exist, err := s.repo.GetModelMapping(ctx, id)
	if err != nil {
		return nil, errors.InternalServer(reason.DatabaseError).WithError(err)
	}
	if !exist {
		return nil, errors.BadRequest(reason.ObjectNotFound)
	}
	mapping.SiteModelID = req.SiteModelID
	mapping.DisplayName = req.DisplayName
	mapping.Description = req.Description
	mapping.Enabled = req.Enabled
	mapping.SortOrder = req.SortOrder
	mapping.SupportsVision = req.SupportsVision
	mapping.FallbackEnabled = req.FallbackEnabled
	mapping.DefaultProviderModelID = req.DefaultProviderModelID
	items := mappingItemsFromReq(req.Items)
	if err := s.repo.UpdateModelMapping(ctx, mapping, items); err != nil {
		return nil, errors.InternalServer(reason.DatabaseError).WithError(err)
	}
	return s.formatMapping(ctx, mapping, items), nil
}

func (s *aiChatConfigService) DeleteModelMapping(ctx context.Context, id int) error {
	if err := s.repo.DeleteModelMapping(ctx, id); err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err)
	}
	return nil
}

func (s *aiChatConfigService) GetModelMapping(ctx context.Context, siteModelID string) (*schema.AIModelMappingResp, error) {
	mapping, exist, err := s.repo.GetModelMappingBySiteModelID(ctx, siteModelID)
	if err != nil {
		return nil, errors.InternalServer(reason.DatabaseError).WithError(err)
	}
	if !exist || !mapping.Enabled {
		return nil, errors.BadRequest(reason.ObjectNotFound)
	}
	items, _ := s.repo.ListModelMappingItems(ctx, mapping.ID)
	return s.formatMapping(ctx, mapping, items), nil
}

func (s *aiChatConfigService) ResolveUpstreamModel(ctx context.Context, siteModelID string) (*schema.AIUpstreamModelResp, error) {
	mapping, exist, err := s.repo.GetModelMappingBySiteModelID(ctx, siteModelID)
	if err != nil {
		return nil, errors.InternalServer(reason.DatabaseError).WithError(err)
	}
	if !exist || !mapping.Enabled {
		return nil, errors.BadRequest("model is not available")
	}
	items, err := s.repo.ListModelMappingItems(ctx, mapping.ID)
	if err != nil {
		return nil, errors.InternalServer(reason.DatabaseError).WithError(err)
	}
	sort.SliceStable(items, func(i, j int) bool { return items[i].Priority < items[j].Priority })
	for _, item := range items {
		if !item.Enabled {
			continue
		}
		if mapping.DefaultProviderModelID != "" && item.ProviderModelID != mapping.DefaultProviderModelID {
			continue
		}
		return s.upstreamFromItem(ctx, item, mapping.SupportsVision)
	}
	for _, item := range items {
		if item.Enabled {
			return s.upstreamFromItem(ctx, item, mapping.SupportsVision)
		}
	}
	return nil, errors.BadRequest("no enabled upstream model")
}

func (s *aiChatConfigService) upstreamFromItem(ctx context.Context, item *entity.AIModelMappingItem, supportsVision bool) (*schema.AIUpstreamModelResp, error) {
	provider, exist, err := s.repo.GetProvider(ctx, item.ProviderID)
	if err != nil {
		return nil, errors.InternalServer(reason.DatabaseError).WithError(err)
	}
	if !exist || !provider.Enabled {
		return nil, errors.BadRequest("provider is not available")
	}
	return &schema.AIUpstreamModelResp{
		ProviderID:      provider.ID,
		ProviderName:    provider.Name,
		ProviderModelID: item.ProviderModelID,
		BaseURL:         provider.BaseURL,
		APIKey:          provider.APIKey,
		SupportsVision:  supportsVision,
		SupportsStream:  provider.SupportsStream,
	}, nil
}

func (s *aiChatConfigService) ListSubscriptionPlans(ctx context.Context) ([]*schema.AISubscriptionPlanResp, error) {
	if err := s.repo.EnsureFreePlan(ctx); err != nil {
		return nil, errors.InternalServer(reason.DatabaseError).WithError(err)
	}
	plans, err := s.repo.ListSubscriptionPlans(ctx)
	if err != nil {
		return nil, errors.InternalServer(reason.DatabaseError).WithError(err)
	}
	resp := make([]*schema.AISubscriptionPlanResp, 0, len(plans))
	for _, plan := range plans {
		resp = append(resp, s.formatPlan(ctx, plan))
	}
	return resp, nil
}

func (s *aiChatConfigService) CreateSubscriptionPlan(ctx context.Context, req *schema.AISubscriptionPlanReq) (*schema.AISubscriptionPlanResp, error) {
	if err := s.validatePlanReq(ctx, 0, req); err != nil {
		return nil, err
	}
	plan := planFromReq(req)
	if err := s.repo.CreateSubscriptionPlan(ctx, plan, req.ModelMappingIDs); err != nil {
		return nil, errors.InternalServer(reason.DatabaseError).WithError(err)
	}
	return s.formatPlan(ctx, plan), nil
}

func (s *aiChatConfigService) UpdateSubscriptionPlan(ctx context.Context, id int, req *schema.AISubscriptionPlanReq) (*schema.AISubscriptionPlanResp, error) {
	plan, exist, err := s.repo.GetSubscriptionPlan(ctx, id)
	if err != nil {
		return nil, errors.InternalServer(reason.DatabaseError).WithError(err)
	}
	if !exist {
		return nil, errors.BadRequest(reason.ObjectNotFound)
	}
	if plan.PlanID == "free" && req.PlanID != "free" {
		return nil, errors.BadRequest("FREE plan_id cannot be changed")
	}
	if err := s.validatePlanReq(ctx, id, req); err != nil {
		return nil, err
	}
	updated := planFromReq(req)
	updated.ID = id
	if err := s.repo.UpdateSubscriptionPlan(ctx, updated, req.ModelMappingIDs); err != nil {
		return nil, errors.InternalServer(reason.DatabaseError).WithError(err)
	}
	return s.formatPlan(ctx, updated), nil
}

func (s *aiChatConfigService) DeleteSubscriptionPlan(ctx context.Context, id int) error {
	plan, exist, err := s.repo.GetSubscriptionPlan(ctx, id)
	if err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err)
	}
	if !exist {
		return errors.BadRequest(reason.ObjectNotFound)
	}
	if plan.PlanID == "free" {
		return errors.BadRequest("FREE plan cannot be deleted")
	}
	if err := s.repo.DeleteSubscriptionPlan(ctx, id); err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err)
	}
	return nil
}

func (s *aiChatConfigService) GetAvailableModelsForPlan(ctx context.Context, planID string) ([]string, error) {
	plan, exist, err := s.repo.GetSubscriptionPlanByPlanID(ctx, planID)
	if err != nil {
		return nil, errors.InternalServer(reason.DatabaseError).WithError(err)
	}
	if !exist || !plan.Enabled {
		return nil, errors.BadRequest("subscription plan is not available")
	}
	return s.formatPlan(ctx, plan).AvailableModelIDs, nil
}

func (s *aiChatConfigService) ListUserAvailableModels(ctx context.Context, userID string) ([]*schema.AIChatModelResp, error) {
	user, exist, err := s.userRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, errors.InternalServer(reason.DatabaseError).WithError(err)
	}
	if !exist {
		return nil, errors.BadRequest(reason.UserNotFound)
	}
	plan, _, err := s.getEffectiveUserPlan(ctx, user)
	if err != nil {
		return nil, err
	}
	relations, err := s.repo.ListSubscriptionPlanModels(ctx, plan.ID)
	if err != nil {
		return nil, errors.InternalServer(reason.DatabaseError).WithError(err)
	}
	models := make([]*schema.AIChatModelResp, 0, len(relations))
	for _, rel := range relations {
		mapping, exist, err := s.repo.GetModelMapping(ctx, rel.ModelMappingID)
		if err != nil {
			return nil, errors.InternalServer(reason.DatabaseError).WithError(err)
		}
		if !exist || !mapping.Enabled {
			continue
		}
		rate := 1.0
		if consumeRate, exist, err := s.repo.GetConsumeRateByModelMappingID(ctx, mapping.ID); err == nil && exist && consumeRate.Enabled {
			rate = consumeRate.ConsumeRate
		}
		models = append(models, &schema.AIChatModelResp{
			SiteModelID:    mapping.SiteModelID,
			DisplayName:    fallbackText(mapping.DisplayName, mapping.SiteModelID),
			Description:    mapping.Description,
			ConsumeRate:    rate,
			Enabled:        mapping.Enabled,
			SupportsVision: mapping.SupportsVision,
		})
	}
	return models, nil
}

func (s *aiChatConfigService) CheckUserModelPermission(ctx context.Context, userID, siteModelID string) error {
	user, exist, err := s.userRepo.GetByUserID(ctx, userID)
	if err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err)
	}
	if !exist {
		return errors.BadRequest(reason.UserNotFound)
	}
	plan, _, err := s.getEffectiveUserPlan(ctx, user)
	if err != nil {
		return err
	}
	models, err := s.GetAvailableModelsForPlan(ctx, plan.PlanID)
	if err != nil {
		return err
	}
	for _, model := range models {
		if model == siteModelID {
			return nil
		}
	}
	return errors.BadRequest("current subscription plan cannot use this model")
}

func (s *aiChatConfigService) GetSubscriptionOverview(ctx context.Context, userID string) (*schema.AISubscriptionOverviewResp, error) {
	user, exist, err := s.userRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, errors.InternalServer(reason.DatabaseError).WithError(err)
	}
	if !exist {
		return nil, errors.BadRequest(reason.UserNotFound)
	}
	if err := s.repo.EnsureFreePlan(ctx); err != nil {
		return nil, errors.InternalServer(reason.DatabaseError).WithError(err)
	}
	plan, _, err := s.getEffectiveUserPlan(ctx, user)
	if err != nil {
		return nil, err
	}
	planResp := s.formatPlan(ctx, plan)
	periodStart := user.SubscriptionStartedAt
	if plan.PlanID == "free" || periodStart.IsZero() {
		periodStart = user.CreatedAt
	}
	periodEnd := user.SubscriptionExpiresAt
	expiresAt := int64(0)
	if plan.PlanID != "free" && !periodEnd.IsZero() {
		expiresAt = periodEnd.Unix()
	}
	if plan.PlanID == "free" || periodEnd.IsZero() {
		periodEnd = time.Time{}
	}
	monthStart, monthEnd := currentMonthRange()
	chatUsage, err := s.repo.SumUserChatUsage(ctx, userID, monthStart, monthEnd)
	if err != nil {
		return nil, errors.InternalServer(reason.DatabaseError).WithError(err)
	}
	chatUsed := int(math.Ceil(chatUsage))
	imageUsed := 0
	chatRemaining := plan.ChatPoints - chatUsed
	if plan.ChatPoints == -1 {
		chatRemaining = -1
	} else if chatRemaining < 0 {
		chatRemaining = 0
	}
	imageRemaining := plan.ImageQuota - imageUsed
	if imageRemaining < 0 {
		imageRemaining = 0
	}
	return &schema.AISubscriptionOverviewResp{
		PlanID:              plan.PlanID,
		PlanName:            plan.Name,
		ChatPoints:          plan.ChatPoints,
		ChatPointsUsed:      chatUsed,
		ChatPointsRemaining: chatRemaining,
		ImageQuota:          plan.ImageQuota,
		ImageQuotaUsed:      imageUsed,
		ImageQuotaRemaining: imageRemaining,
		AvailableModels:     planResp.AvailableModelIDs,
		ConsumeRates:        s.listSubscriptionModelRates(ctx),
		PeriodStart:         unixOrZero(periodStart),
		PeriodEnd:           unixOrZero(periodEnd),
		ExpiresAt:           expiresAt,
	}, nil
}

func (s *aiChatConfigService) GetSubscriptionPurchase(ctx context.Context, userID string) (*schema.AISubscriptionPurchaseResp, error) {
	if err := s.repo.EnsureFreePlan(ctx); err != nil {
		return nil, errors.InternalServer(reason.DatabaseError).WithError(err)
	}
	currentPlanID := "free"
	if userID != "" {
		if user, exist, err := s.userRepo.GetByUserID(ctx, userID); err != nil {
			return nil, errors.InternalServer(reason.DatabaseError).WithError(err)
		} else if exist {
			if plan, _, err := s.getEffectiveUserPlan(ctx, user); err == nil {
				currentPlanID = plan.PlanID
			}
		}
	}
	plans, err := s.ListSubscriptionPlans(ctx)
	if err != nil {
		return nil, err
	}
	enabledPlans := make([]*schema.AISubscriptionPlanResp, 0, len(plans))
	for _, plan := range plans {
		if plan.Enabled {
			enabledPlans = append(enabledPlans, plan)
		}
	}
	return &schema.AISubscriptionPurchaseResp{
		CurrentPlanID: currentPlanID,
		Plans:         enabledPlans,
		ConsumeRates:  s.listSubscriptionModelRates(ctx),
	}, nil
}

func (s *aiChatConfigService) ListSubscriptionRedeemCodes(ctx context.Context) ([]*schema.AISubscriptionRedeemCodeResp, error) {
	codes, err := s.repo.ListRedeemCodes(ctx)
	if err != nil {
		return nil, errors.InternalServer(reason.DatabaseError).WithError(err)
	}
	resp := make([]*schema.AISubscriptionRedeemCodeResp, 0, len(codes))
	for _, code := range codes {
		resp = append(resp, s.formatRedeemCode(ctx, code))
	}
	return resp, nil
}

func (s *aiChatConfigService) GenerateSubscriptionRedeemCodes(ctx context.Context, req *schema.AISubscriptionRedeemCodeGenerateReq) ([]*schema.AISubscriptionRedeemCodeResp, error) {
	if req.Count <= 0 {
		req.Count = 1
	}
	if req.Count > 500 {
		return nil, errors.BadRequest("count cannot be greater than 500")
	}
	if req.DurationMonths <= 0 {
		req.DurationMonths = 1
	}
	if req.DurationMonths > 120 {
		return nil, errors.BadRequest("duration_months cannot be greater than 120")
	}
	plan, exist, err := s.repo.GetSubscriptionPlan(ctx, req.PlanID)
	if err != nil {
		return nil, errors.InternalServer(reason.DatabaseError).WithError(err)
	}
	if !exist || !plan.Enabled {
		return nil, errors.BadRequest("subscription plan is not available")
	}
	if plan.PlanID == "free" {
		return nil, errors.BadRequest("FREE plan does not need redeem codes")
	}
	prefix := normalizeRedeemPrefix(req.Prefix)
	if prefix == "" {
		prefix = strings.ToUpper(plan.PlanID)
	}
	batchNo := fmt.Sprintf("B%s", time.Now().Format("20060102150405"))
	codes := make([]*entity.AISubscriptionRedeemCode, 0, req.Count)
	seen := map[string]bool{}
	for len(codes) < req.Count {
		code, err := newRedeemCode(prefix)
		if err != nil {
			return nil, errors.InternalServer(reason.UnknownError).WithError(err)
		}
		if seen[code] {
			continue
		}
		if _, exist, err := s.repo.GetRedeemCodeByCode(ctx, code); err != nil {
			return nil, errors.InternalServer(reason.DatabaseError).WithError(err)
		} else if exist {
			continue
		}
		seen[code] = true
		codes = append(codes, &entity.AISubscriptionRedeemCode{
			Code:           code,
			PlanID:         plan.ID,
			DurationMonths: req.DurationMonths,
			Enabled:        true,
			BatchNo:        batchNo,
			Remark:         req.Remark,
		})
	}
	if err := s.repo.CreateRedeemCodes(ctx, codes); err != nil {
		return nil, errors.InternalServer(reason.DatabaseError).WithError(err)
	}
	resp := make([]*schema.AISubscriptionRedeemCodeResp, 0, len(codes))
	for _, code := range codes {
		resp = append(resp, s.formatRedeemCode(ctx, code))
	}
	return resp, nil
}

func (s *aiChatConfigService) RedeemSubscriptionCode(ctx context.Context, userID string, req *schema.AISubscriptionRedeemReq) (*schema.AISubscriptionRedeemResp, error) {
	if userID == "" {
		return nil, errors.Unauthorized(reason.UnauthorizedError)
	}
	codeText := normalizeRedeemCode(req.Code)
	if codeText == "" {
		return nil, errors.BadRequest("redeem code is required")
	}
	redeemCode, exist, err := s.repo.GetRedeemCodeByCode(ctx, codeText)
	if err != nil {
		return nil, errors.InternalServer(reason.DatabaseError).WithError(err)
	}
	if !exist || !redeemCode.Enabled {
		return nil, errors.BadRequest("redeem code is invalid")
	}
	if redeemCode.Used {
		return nil, errors.BadRequest("redeem code has been used")
	}
	plan, exist, err := s.repo.GetSubscriptionPlan(ctx, redeemCode.PlanID)
	if err != nil {
		return nil, errors.InternalServer(reason.DatabaseError).WithError(err)
	}
	if !exist || !plan.Enabled {
		return nil, errors.BadRequest("subscription plan is not available")
	}
	user, exist, err := s.userRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, errors.InternalServer(reason.DatabaseError).WithError(err)
	}
	if !exist {
		return nil, errors.BadRequest(reason.UserNotFound)
	}
	now := time.Now()
	startAt := now
	baseAt := now
	if !user.SubscriptionExpiresAt.IsZero() && user.SubscriptionExpiresAt.After(now) && user.SubscriptionLevel != "free" {
		baseAt = user.SubscriptionExpiresAt
		if !user.SubscriptionStartedAt.IsZero() {
			startAt = user.SubscriptionStartedAt
		}
	}
	expiresAt := baseAt.AddDate(0, redeemCode.DurationMonths, 0)
	user.SubscriptionLevel = plan.PlanID
	user.SubscriptionStartedAt = startAt
	user.SubscriptionExpiresAt = expiresAt
	redeemCode.Used = true
	redeemCode.UsedByUserID = userID
	redeemCode.UsedAt = now
	if err := s.repo.UseRedeemCode(ctx, redeemCode, user); err != nil {
		if stderrors.Is(err, ai_chat_config_repo.ErrRedeemCodeUsed) {
			return nil, errors.BadRequest("redeem code has been used")
		}
		return nil, errors.InternalServer(reason.DatabaseError).WithError(err)
	}
	return &schema.AISubscriptionRedeemResp{
		PlanID:    plan.PlanID,
		PlanName:  plan.Name,
		StartedAt: user.SubscriptionStartedAt.Unix(),
		ExpiresAt: user.SubscriptionExpiresAt.Unix(),
	}, nil
}

func (s *aiChatConfigService) ListConsumeRates(ctx context.Context) ([]*schema.AIModelConsumeRateResp, error) {
	rates, err := s.repo.ListConsumeRates(ctx)
	if err != nil {
		return nil, errors.InternalServer(reason.DatabaseError).WithError(err)
	}
	resp := make([]*schema.AIModelConsumeRateResp, 0, len(rates))
	for _, rate := range rates {
		resp = append(resp, s.formatRate(ctx, rate))
	}
	return resp, nil
}

func (s *aiChatConfigService) SaveConsumeRate(ctx context.Context, id int, req *schema.AIModelConsumeRateReq) (*schema.AIModelConsumeRateResp, error) {
	if req.ModelMappingID <= 0 {
		return nil, errors.BadRequest("model_mapping_id is required")
	}
	if req.ConsumeRate <= 0 {
		return nil, errors.BadRequest("consume_rate must be greater than 0")
	}
	mapping, exist, err := s.repo.GetModelMapping(ctx, req.ModelMappingID)
	if err != nil {
		return nil, errors.InternalServer(reason.DatabaseError).WithError(err)
	}
	if !exist {
		return nil, errors.BadRequest(reason.ObjectNotFound)
	}
	var rate *entity.AIModelConsumeRate
	if id > 0 {
		rate, exist, err = s.repo.GetConsumeRate(ctx, id)
		if err != nil {
			return nil, errors.InternalServer(reason.DatabaseError).WithError(err)
		}
		if !exist {
			return nil, errors.BadRequest(reason.ObjectNotFound)
		}
	} else if current, ok, _ := s.repo.GetConsumeRateByModelMappingID(ctx, req.ModelMappingID); ok {
		rate = current
	} else {
		rate = &entity.AIModelConsumeRate{}
	}
	rate.ModelMappingID = mapping.ID
	rate.ConsumeRate = req.ConsumeRate
	rate.Enabled = req.Enabled
	rate.Remark = req.Remark
	if err := s.repo.SaveConsumeRate(ctx, rate); err != nil {
		return nil, errors.InternalServer(reason.DatabaseError).WithError(err)
	}
	return s.formatRate(ctx, rate), nil
}

func (s *aiChatConfigService) GetModelConsumeRate(ctx context.Context, siteModelID string) (float64, error) {
	mapping, exist, err := s.repo.GetModelMappingBySiteModelID(ctx, siteModelID)
	if err != nil {
		return 0, errors.InternalServer(reason.DatabaseError).WithError(err)
	}
	if !exist || !mapping.Enabled {
		return 0, errors.BadRequest("model is not available")
	}
	rate, exist, err := s.repo.GetConsumeRateByModelMappingID(ctx, mapping.ID)
	if err != nil {
		return 0, errors.InternalServer(reason.DatabaseError).WithError(err)
	}
	if !exist || !rate.Enabled {
		return 1, nil
	}
	return rate.ConsumeRate, nil
}

func (s *aiChatConfigService) CalculateChatCost(ctx context.Context, siteModelID string, baseCost float64) (float64, error) {
	rate, err := s.GetModelConsumeRate(ctx, siteModelID)
	if err != nil {
		return 0, err
	}
	return baseCost * rate, nil
}

func (s *aiChatConfigService) DeductUserPoints(ctx context.Context, userID string, cost float64) error {
	user, exist, err := s.userRepo.GetByUserID(ctx, userID)
	if err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err)
	}
	if !exist {
		return errors.BadRequest(reason.UserNotFound)
	}
	plan, _, err := s.getEffectiveUserPlan(ctx, user)
	if err != nil {
		return err
	}
	if plan.ChatPoints == -1 {
		return nil
	}
	monthStart, monthEnd := currentMonthRange()
	used, err := s.repo.SumUserChatUsage(ctx, userID, monthStart, monthEnd)
	if err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err)
	}
	if int(math.Ceil(used+cost)) > plan.ChatPoints {
		return errors.BadRequest("chat points are insufficient")
	}
	return nil
}

func (s *aiChatConfigService) RecordChatUsage(ctx context.Context, req *schema.AIChatUsageLogReq) error {
	if req == nil || req.UserID == "" || req.ChatCompletionID == "" || req.SiteModelID == "" {
		return nil
	}
	if req.ConsumePoints <= 0 {
		req.ConsumePoints = 1
	}
	if err := s.repo.CreateUsageLog(ctx, &entity.AIChatUsageLog{
		UserID:           req.UserID,
		ConversationID:   req.ConversationID,
		ChatCompletionID: req.ChatCompletionID,
		SiteModelID:      req.SiteModelID,
		ProviderID:       req.ProviderID,
		ProviderName:     req.ProviderName,
		ProviderModelID:  req.ProviderModelID,
		ConsumePoints:    req.ConsumePoints,
	}); err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err)
	}
	return nil
}

func (s *aiChatConfigService) validateMappingReq(req *schema.AIModelMappingReq) error {
	req.SiteModelID = strings.TrimSpace(req.SiteModelID)
	if !modelIDPattern.MatchString(req.SiteModelID) {
		return errors.BadRequest("site_model_id can only contain lowercase letters, numbers, hyphen and underscore")
	}
	if len(req.Items) == 0 {
		return errors.BadRequest("at least one upstream model is required")
	}
	priority := make(map[int]bool)
	defaultFound := req.DefaultProviderModelID == ""
	for _, item := range req.Items {
		if item.ProviderID <= 0 || item.ProviderModelID == "" {
			return errors.BadRequest("upstream model is invalid")
		}
		if priority[item.Priority] {
			return errors.BadRequest("upstream priorities cannot be duplicated")
		}
		priority[item.Priority] = true
		if item.ProviderModelID == req.DefaultProviderModelID {
			defaultFound = true
		}
	}
	if !defaultFound {
		return errors.BadRequest("default upstream model must be included in upstream model list")
	}
	return nil
}

func (s *aiChatConfigService) validatePlanReq(ctx context.Context, excludeID int, req *schema.AISubscriptionPlanReq) error {
	req.PlanID = strings.TrimSpace(req.PlanID)
	if !modelIDPattern.MatchString(req.PlanID) {
		return errors.BadRequest("plan_id can only contain lowercase letters, numbers, hyphen and underscore")
	}
	if req.MonthlyPrice < 0 || req.ChatPoints < -1 || req.ImageQuota < 0 {
		return errors.BadRequest("monthly_price and image_quota cannot be negative, chat_points must be -1 or greater")
	}
	if len(req.ModelMappingIDs) == 0 {
		return errors.BadRequest("at least one available model is required")
	}
	count, err := s.repo.CountCustomSubscriptionPlans(ctx, excludeID)
	if err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err)
	}
	if req.PlanID != "free" && count >= 3 {
		return errors.BadRequest("only three custom subscription plans are allowed")
	}
	return nil
}

func (s *aiChatConfigService) formatProvider(provider *entity.AIProvider, models []*entity.AIProviderModel, mask bool) *schema.AIProviderResp {
	apiKey := provider.APIKey
	if mask {
		apiKey = maskSecret(apiKey)
	}
	return &schema.AIProviderResp{
		ID:             provider.ID,
		Name:           provider.Name,
		BaseURL:        provider.BaseURL,
		APIKey:         apiKey,
		Enabled:        provider.Enabled,
		SupportsStream: provider.SupportsStream,
		Remark:         provider.Remark,
		Models:         s.formatProviderModels(models),
		CreatedAt:      provider.CreatedAt.Unix(),
		UpdatedAt:      provider.UpdatedAt.Unix(),
	}
}

func (s *aiChatConfigService) formatProviderModels(models []*entity.AIProviderModel) []*schema.AIProviderModelResp {
	resp := make([]*schema.AIProviderModelResp, 0, len(models))
	for _, model := range models {
		resp = append(resp, &schema.AIProviderModelResp{
			ID:              model.ID,
			ProviderID:      model.ProviderID,
			ProviderModelID: model.ProviderModelID,
			ModelName:       model.ModelName,
			Enabled:         model.Enabled,
			FetchedAt:       model.FetchedAt.Unix(),
			CreatedAt:       model.CreatedAt.Unix(),
			UpdatedAt:       model.UpdatedAt.Unix(),
		})
	}
	return resp
}

func (s *aiChatConfigService) formatMapping(ctx context.Context, mapping *entity.AIModelMapping, items []*entity.AIModelMappingItem) *schema.AIModelMappingResp {
	respItems := make([]*schema.AIModelMappingItemResp, 0, len(items))
	for _, item := range items {
		providerName := ""
		if provider, exist, _ := s.repo.GetProvider(ctx, item.ProviderID); exist {
			providerName = provider.Name
		}
		respItems = append(respItems, &schema.AIModelMappingItemResp{
			ID:              item.ID,
			MappingID:       item.MappingID,
			ProviderID:      item.ProviderID,
			ProviderName:    providerName,
			ProviderModelID: item.ProviderModelID,
			Priority:        item.Priority,
			Enabled:         item.Enabled,
			CreatedAt:       item.CreatedAt.Unix(),
			UpdatedAt:       item.UpdatedAt.Unix(),
		})
	}
	return &schema.AIModelMappingResp{
		ID:                     mapping.ID,
		SiteModelID:            mapping.SiteModelID,
		DisplayName:            mapping.DisplayName,
		Description:            mapping.Description,
		Enabled:                mapping.Enabled,
		SortOrder:              mapping.SortOrder,
		SupportsVision:         mapping.SupportsVision,
		FallbackEnabled:        mapping.FallbackEnabled,
		DefaultProviderModelID: mapping.DefaultProviderModelID,
		Items:                  respItems,
		CreatedAt:              mapping.CreatedAt.Unix(),
		UpdatedAt:              mapping.UpdatedAt.Unix(),
	}
}

func (s *aiChatConfigService) formatPlan(ctx context.Context, plan *entity.AISubscriptionPlan) *schema.AISubscriptionPlanResp {
	relations, _ := s.repo.ListSubscriptionPlanModels(ctx, plan.ID)
	modelIDs := make([]int, 0, len(relations))
	siteModelIDs := make([]string, 0, len(relations))
	for _, rel := range relations {
		modelIDs = append(modelIDs, rel.ModelMappingID)
		if mapping, exist, _ := s.repo.GetModelMapping(ctx, rel.ModelMappingID); exist {
			siteModelIDs = append(siteModelIDs, mapping.SiteModelID)
		}
	}
	return &schema.AISubscriptionPlanResp{
		ID:                plan.ID,
		PlanID:            plan.PlanID,
		Name:              plan.Name,
		Enabled:           plan.Enabled,
		MonthlyPrice:      plan.MonthlyPrice,
		ChatPoints:        plan.ChatPoints,
		ImageQuota:        plan.ImageQuota,
		PurchaseURL:       plan.PurchaseURL,
		ModelMappingIDs:   modelIDs,
		AvailableModelIDs: siteModelIDs,
		TaskDescription:   plan.TaskDescription,
		SortOrder:         plan.SortOrder,
		CreatedAt:         plan.CreatedAt.Unix(),
		UpdatedAt:         plan.UpdatedAt.Unix(),
	}
}

func (s *aiChatConfigService) formatRedeemCode(ctx context.Context, code *entity.AISubscriptionRedeemCode) *schema.AISubscriptionRedeemCodeResp {
	planKey := ""
	planName := ""
	if plan, exist, _ := s.repo.GetSubscriptionPlan(ctx, code.PlanID); exist {
		planKey = plan.PlanID
		planName = plan.Name
	}
	return &schema.AISubscriptionRedeemCodeResp{
		ID:             code.ID,
		Code:           code.Code,
		PlanID:         code.PlanID,
		PlanKey:        planKey,
		PlanName:       planName,
		DurationMonths: code.DurationMonths,
		Enabled:        code.Enabled,
		Used:           code.Used,
		UsedByUserID:   code.UsedByUserID,
		UsedAt:         unixOrZero(code.UsedAt),
		BatchNo:        code.BatchNo,
		Remark:         code.Remark,
		CreatedAt:      code.CreatedAt.Unix(),
		UpdatedAt:      code.UpdatedAt.Unix(),
	}
}

func (s *aiChatConfigService) formatRate(ctx context.Context, rate *entity.AIModelConsumeRate) *schema.AIModelConsumeRateResp {
	siteModelID := ""
	if mapping, exist, _ := s.repo.GetModelMapping(ctx, rate.ModelMappingID); exist {
		siteModelID = mapping.SiteModelID
	}
	return &schema.AIModelConsumeRateResp{
		ID:             rate.ID,
		ModelMappingID: rate.ModelMappingID,
		SiteModelID:    siteModelID,
		ConsumeRate:    rate.ConsumeRate,
		Enabled:        rate.Enabled,
		Remark:         rate.Remark,
		CreatedAt:      rate.CreatedAt.Unix(),
		UpdatedAt:      rate.UpdatedAt.Unix(),
	}
}

func (s *aiChatConfigService) listSubscriptionModelRates(ctx context.Context) []*schema.AISubscriptionModelRate {
	mappings, err := s.repo.ListModelMappings(ctx)
	if err != nil {
		return nil
	}
	rates := make([]*schema.AISubscriptionModelRate, 0, len(mappings))
	for _, mapping := range mappings {
		if !mapping.Enabled {
			continue
		}
		consumeRate := 1.0
		if rate, exist, err := s.repo.GetConsumeRateByModelMappingID(ctx, mapping.ID); err == nil && exist && rate.Enabled {
			consumeRate = rate.ConsumeRate
		}
		rates = append(rates, &schema.AISubscriptionModelRate{
			SiteModelID: mapping.SiteModelID,
			ConsumeRate: consumeRate,
		})
	}
	return rates
}

func (s *aiChatConfigService) getEffectiveUserPlan(ctx context.Context, user *entity.User) (*entity.AISubscriptionPlan, bool, error) {
	planID := user.SubscriptionLevel
	if planID == "" {
		planID = "free"
	}
	expired := false
	if planID != "free" && (user.SubscriptionExpiresAt.IsZero() || !user.SubscriptionExpiresAt.After(time.Now())) {
		planID = "free"
		expired = true
	}
	plan, exist, err := s.repo.GetSubscriptionPlanByPlanID(ctx, planID)
	if err != nil {
		return nil, expired, errors.InternalServer(reason.DatabaseError).WithError(err)
	}
	if !exist || !plan.Enabled {
		plan, exist, err = s.repo.GetSubscriptionPlanByPlanID(ctx, "free")
		if err != nil {
			return nil, expired, errors.InternalServer(reason.DatabaseError).WithError(err)
		}
		if !exist {
			return nil, expired, errors.BadRequest("subscription plan is not available")
		}
		expired = true
	}
	return plan, expired, nil
}

func normalizeRedeemPrefix(value string) string {
	value = strings.ToUpper(strings.TrimSpace(value))
	value = regexp.MustCompile(`[^A-Z0-9]`).ReplaceAllString(value, "")
	if len(value) > 16 {
		value = value[:16]
	}
	return value
}

func normalizeRedeemCode(value string) string {
	value = strings.ToUpper(strings.TrimSpace(value))
	value = strings.ReplaceAll(value, " ", "")
	return value
}

func newRedeemCode(prefix string) (string, error) {
	buf := make([]byte, 8)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	raw := strings.ToUpper(hex.EncodeToString(buf))
	parts := []string{raw[0:4], raw[4:8], raw[8:12], raw[12:16]}
	if prefix != "" {
		return prefix + "-" + strings.Join(parts, "-"), nil
	}
	return strings.Join(parts, "-"), nil
}

func unixOrZero(value time.Time) int64 {
	if value.IsZero() {
		return 0
	}
	return value.Unix()
}

func currentMonthRange() (time.Time, time.Time) {
	now := time.Now()
	start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	return start, start.AddDate(0, 1, 0)
}

func fallbackText(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func mappingItemsFromReq(reqItems []*schema.AIModelMappingItemReq) []*entity.AIModelMappingItem {
	items := make([]*entity.AIModelMappingItem, 0, len(reqItems))
	for _, item := range reqItems {
		items = append(items, &entity.AIModelMappingItem{
			ProviderID:      item.ProviderID,
			ProviderModelID: item.ProviderModelID,
			Priority:        item.Priority,
			Enabled:         item.Enabled,
		})
	}
	return items
}

func planFromReq(req *schema.AISubscriptionPlanReq) *entity.AISubscriptionPlan {
	return &entity.AISubscriptionPlan{
		PlanID:          req.PlanID,
		Name:            req.Name,
		Enabled:         req.Enabled,
		MonthlyPrice:    req.MonthlyPrice,
		ChatPoints:      req.ChatPoints,
		ImageQuota:      req.ImageQuota,
		PurchaseURL:     req.PurchaseURL,
		TaskDescription: req.TaskDescription,
		SortOrder:       req.SortOrder,
	}
}

func normalizeBaseURL(raw string) (string, error) {
	raw = strings.TrimRight(strings.TrimSpace(raw), "/")
	parsed, err := url.ParseRequestURI(raw)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return "", err
	}
	return raw, nil
}

func fetchOpenAIModels(baseURL, apiKey string) ([]string, error) {
	baseURL = strings.TrimRight(baseURL, "/")
	resp, err := resty.New().
		SetHeader("Authorization", fmt.Sprintf("Bearer %s", apiKey)).
		SetHeader("Content-Type", "application/json").
		R().
		Get(baseURL + "/models")
	if err != nil {
		return nil, err
	}
	if !resp.IsSuccess() {
		return nil, fmt.Errorf("status %d: %s", resp.StatusCode(), resp.String())
	}
	payload := &schema.GetAIModelsResp{}
	if err := json.Unmarshal(resp.Body(), payload); err != nil {
		return nil, err
	}
	models := make([]string, 0, len(payload.Data))
	for _, model := range payload.Data {
		if model.Id != "" {
			models = append(models, model.Id)
		}
	}
	return models, nil
}

func testOpenAIChat(baseURL, apiKey, modelID string) (message, raw string, err error) {
	baseURL = strings.TrimRight(baseURL, "/")
	payload := map[string]any{
		"model": modelID,
		"messages": []map[string]string{
			{
				"role":    "user",
				"content": "hi",
			},
		},
		"stream": false,
	}
	resp, err := resty.New().
		SetHeader("Authorization", fmt.Sprintf("Bearer %s", apiKey)).
		SetHeader("Content-Type", "application/json").
		R().
		SetBody(payload).
		Post(baseURL + "/chat/completions")
	if err != nil {
		return "", "", err
	}
	raw = resp.String()
	if !resp.IsSuccess() {
		return "", raw, fmt.Errorf("status %d: %s", resp.StatusCode(), raw)
	}
	var parsed struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(resp.Body(), &parsed); err != nil {
		return "", raw, err
	}
	if len(parsed.Choices) > 0 {
		message = parsed.Choices[0].Message.Content
	}
	if message == "" {
		message = raw
	}
	return message, raw, nil
}

func maskSecret(secret string) string {
	if secret == "" {
		return ""
	}
	if len(secret) <= 8 {
		return strings.Repeat("*", len(secret))
	}
	return secret[:4] + strings.Repeat("*", len(secret)-8) + secret[len(secret)-4:]
}

func isAllMask(value string) bool {
	value = strings.TrimSpace(value)
	return value != "" && strings.Trim(value, "*") == ""
}
