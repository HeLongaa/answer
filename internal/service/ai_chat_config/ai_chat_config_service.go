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
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	stderrors "errors"
	"fmt"
	"io"
	"math"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/apache/answer/internal/base/constant"
	"github.com/apache/answer/internal/base/reason"
	"github.com/apache/answer/internal/entity"
	ai_chat_config_repo "github.com/apache/answer/internal/repo/ai_chat_config"
	"github.com/apache/answer/internal/schema"
	"github.com/apache/answer/internal/service/service_config"
	usercommon "github.com/apache/answer/internal/service/user_common"
	"github.com/apache/answer/pkg/uid"
	"github.com/go-resty/resty/v2"
	"github.com/segmentfault/pacman/errors"
	"github.com/segmentfault/pacman/log"
)

var (
	modelIDPattern    = regexp.MustCompile(`^[a-z0-9_-]+$`)
	base64DataPattern = regexp.MustCompile(`"b64_json"\s*:\s*"[^"]+"`)
	dataURLPattern    = regexp.MustCompile(`data:image/[^;]+;base64,[A-Za-z0-9+/=_-]+`)
)

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

	ListImageProviders(ctx context.Context) ([]*schema.AIImageProviderResp, error)
	CreateImageProvider(ctx context.Context, req *schema.AIImageProviderReq) (*schema.AIImageProviderResp, error)
	UpdateImageProvider(ctx context.Context, id int, req *schema.AIImageProviderReq) (*schema.AIImageProviderResp, error)
	DeleteImageProvider(ctx context.Context, id int) error
	ListImageModels(ctx context.Context, onlyEnabled bool) ([]*schema.AIImageModelResp, error)
	SaveImageModel(ctx context.Context, id int, req *schema.AIImageModelReq) (*schema.AIImageModelResp, error)
	DeleteImageModel(ctx context.Context, id int) error
	GetImageSetting(ctx context.Context) (*schema.AIImageSettingResp, error)
	SaveImageSetting(ctx context.Context, req *schema.AIImageSettingReq) (*schema.AIImageSettingResp, error)
	GenerateImage(ctx context.Context, userID string, req *schema.AIImageGenerateReq) (*schema.AIImageGenerateResp, error)
	EditImage(ctx context.Context, userID string, req *schema.AIImageEditReq) (*schema.AIImageGenerateResp, error)
	ListUserImageGenerations(ctx context.Context, userID string, limit int) ([]*schema.AIImageGenerationResp, error)
	GetUserImageFilePath(ctx context.Context, userID, ownerID, filename string) (string, error)
	ListVideoProviders(ctx context.Context) ([]*schema.AIVideoProviderResp, error)
	CreateVideoProvider(ctx context.Context, req *schema.AIVideoProviderReq) (*schema.AIVideoProviderResp, error)
	UpdateVideoProvider(ctx context.Context, id int, req *schema.AIVideoProviderReq) (*schema.AIVideoProviderResp, error)
	DeleteVideoProvider(ctx context.Context, id int) error
	ListVideoModels(ctx context.Context, onlyEnabled bool) ([]*schema.AIVideoModelResp, error)
	SaveVideoModel(ctx context.Context, id int, req *schema.AIVideoModelReq) (*schema.AIVideoModelResp, error)
	DeleteVideoModel(ctx context.Context, id int) error
	GetVideoSetting(ctx context.Context) (*schema.AIVideoSettingResp, error)
	SaveVideoSetting(ctx context.Context, req *schema.AIVideoSettingReq) (*schema.AIVideoSettingResp, error)
	GenerateVideo(ctx context.Context, userID string, req *schema.AIVideoGenerateReq) (*schema.AIVideoGenerateResp, error)
	ListUserVideoGenerations(ctx context.Context, userID string, limit int) ([]*schema.AIVideoGenerationResp, error)
	GetUserVideoFilePath(ctx context.Context, userID, ownerID, filename string) (string, error)
}

type aiChatConfigService struct {
	repo          ai_chat_config_repo.AIChatConfigRepo
	userRepo      usercommon.UserRepo
	serviceConfig *service_config.ServiceConfig
}

func NewAIChatConfigService(
	repo ai_chat_config_repo.AIChatConfigRepo,
	userRepo usercommon.UserRepo,
	serviceConfig *service_config.ServiceConfig,
) AiChatConfigService {
	return &aiChatConfigService{repo: repo, userRepo: userRepo, serviceConfig: serviceConfig}
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
	if err := s.ensureImageTables(ctx); err != nil {
		return nil, errors.InternalServer(reason.DatabaseError).WithError(err)
	}
	imageUsed, err := s.repo.CountUserImageGenerations(ctx, userID, monthStart, monthEnd)
	if err != nil {
		return nil, errors.InternalServer(reason.DatabaseError).WithError(err)
	}
	if err := s.ensureVideoTables(ctx); err != nil {
		return nil, errors.InternalServer(reason.DatabaseError).WithError(err)
	}
	dayStart, dayEnd := currentDayRange()
	videoDailyUsed, err := s.repo.CountUserVideoGenerations(ctx, userID, dayStart, dayEnd)
	if err != nil {
		return nil, errors.InternalServer(reason.DatabaseError).WithError(err)
	}
	videoUsed, err := s.repo.CountUserVideoGenerations(ctx, userID, monthStart, monthEnd)
	if err != nil {
		return nil, errors.InternalServer(reason.DatabaseError).WithError(err)
	}
	chatRemaining := plan.ChatPoints - chatUsed
	if plan.ChatPoints == -1 {
		chatRemaining = -1
	} else if chatRemaining < 0 {
		chatRemaining = 0
	}
	imageRemaining := plan.ImageQuota - imageUsed
	if plan.ImageQuota == -1 {
		imageRemaining = -1
	} else if imageRemaining < 0 {
		imageRemaining = 0
	}
	videoDailyRemaining := plan.VideoDailyQuota - videoDailyUsed
	if plan.VideoDailyQuota == -1 {
		videoDailyRemaining = -1
	} else if videoDailyRemaining < 0 {
		videoDailyRemaining = 0
	}
	videoRemaining := plan.VideoQuota - videoUsed
	if plan.VideoQuota == -1 {
		videoRemaining = -1
	} else if videoRemaining < 0 {
		videoRemaining = 0
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
		VideoDailyQuota:     plan.VideoDailyQuota,
		VideoDailyUsed:      videoDailyUsed,
		VideoDailyRemaining: videoDailyRemaining,
		VideoQuota:          plan.VideoQuota,
		VideoQuotaUsed:      videoUsed,
		VideoQuotaRemaining: videoRemaining,
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

func (s *aiChatConfigService) ListImageProviders(ctx context.Context) ([]*schema.AIImageProviderResp, error) {
	if err := s.ensureImageTables(ctx); err != nil {
		return nil, errors.InternalServer(reason.DatabaseError).WithError(err)
	}
	providers, err := s.repo.ListImageProviders(ctx)
	if err != nil {
		return nil, errors.InternalServer(reason.DatabaseError).WithError(err)
	}
	resp := make([]*schema.AIImageProviderResp, 0, len(providers))
	for _, provider := range providers {
		resp = append(resp, s.formatImageProvider(provider, true))
	}
	return resp, nil
}

func (s *aiChatConfigService) CreateImageProvider(ctx context.Context, req *schema.AIImageProviderReq) (*schema.AIImageProviderResp, error) {
	if err := s.ensureImageTables(ctx); err != nil {
		return nil, errors.InternalServer(reason.DatabaseError).WithError(err)
	}
	if strings.TrimSpace(req.APIKey) == "" {
		return nil, errors.BadRequest("api_key is required")
	}
	baseURL, err := normalizeBaseURL(req.BaseURL)
	if err != nil {
		return nil, errors.BadRequest("base_url is invalid")
	}
	provider := &entity.AIImageProvider{
		Name:    strings.TrimSpace(req.Name),
		BaseURL: baseURL,
		APIKey:  strings.TrimSpace(req.APIKey),
		Enabled: req.Enabled,
		Remark:  req.Remark,
	}
	if err := s.repo.CreateImageProvider(ctx, provider); err != nil {
		return nil, errors.InternalServer(reason.DatabaseError).WithError(err)
	}
	return s.formatImageProvider(provider, true), nil
}

func (s *aiChatConfigService) UpdateImageProvider(ctx context.Context, id int, req *schema.AIImageProviderReq) (*schema.AIImageProviderResp, error) {
	if err := s.ensureImageTables(ctx); err != nil {
		return nil, errors.InternalServer(reason.DatabaseError).WithError(err)
	}
	provider, exist, err := s.repo.GetImageProvider(ctx, id)
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
	provider.Remark = req.Remark
	cols := []string{"name", "base_url", "enabled", "remark"}
	if strings.TrimSpace(req.APIKey) != "" && !isAllMask(req.APIKey) {
		provider.APIKey = strings.TrimSpace(req.APIKey)
		cols = append(cols, "api_key")
	}
	if err := s.repo.UpdateImageProvider(ctx, provider, cols...); err != nil {
		return nil, errors.InternalServer(reason.DatabaseError).WithError(err)
	}
	return s.formatImageProvider(provider, true), nil
}

func (s *aiChatConfigService) DeleteImageProvider(ctx context.Context, id int) error {
	if err := s.ensureImageTables(ctx); err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err)
	}
	if err := s.repo.DeleteImageProvider(ctx, id); err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err)
	}
	return nil
}

func (s *aiChatConfigService) ListImageModels(ctx context.Context, onlyEnabled bool) ([]*schema.AIImageModelResp, error) {
	if err := s.ensureImageTables(ctx); err != nil {
		return nil, errors.InternalServer(reason.DatabaseError).WithError(err)
	}
	models, err := s.repo.ListImageModels(ctx, onlyEnabled)
	if err != nil {
		return nil, errors.InternalServer(reason.DatabaseError).WithError(err)
	}
	resp := make([]*schema.AIImageModelResp, 0, len(models))
	for _, model := range models {
		resp = append(resp, s.formatImageModel(ctx, model))
	}
	return resp, nil
}

func (s *aiChatConfigService) SaveImageModel(ctx context.Context, id int, req *schema.AIImageModelReq) (*schema.AIImageModelResp, error) {
	if err := s.ensureImageTables(ctx); err != nil {
		return nil, errors.InternalServer(reason.DatabaseError).WithError(err)
	}
	if req.ProviderID <= 0 {
		return nil, errors.BadRequest("provider_id is required")
	}
	if !modelIDPattern.MatchString(strings.TrimSpace(req.SiteModelID)) {
		return nil, errors.BadRequest("site_model_id can only contain lowercase letters, numbers, hyphen and underscore")
	}
	if strings.TrimSpace(req.ProviderModelID) == "" {
		return nil, errors.BadRequest("provider_model_id is required")
	}
	if strings.TrimSpace(req.DefaultSize) == "" {
		req.DefaultSize = "1024x1024"
	}
	if _, exist, err := s.repo.GetImageProvider(ctx, req.ProviderID); err != nil {
		return nil, errors.InternalServer(reason.DatabaseError).WithError(err)
	} else if !exist {
		return nil, errors.BadRequest("provider is not available")
	}
	model := &entity.AIImageModel{}
	if id > 0 {
		current, exist, err := s.repo.GetImageModel(ctx, id)
		if err != nil {
			return nil, errors.InternalServer(reason.DatabaseError).WithError(err)
		}
		if !exist {
			return nil, errors.BadRequest(reason.ObjectNotFound)
		}
		model = current
	}
	model.ProviderID = req.ProviderID
	model.SiteModelID = strings.TrimSpace(req.SiteModelID)
	model.ProviderModelID = strings.TrimSpace(req.ProviderModelID)
	model.DisplayName = req.DisplayName
	model.Description = req.Description
	model.DefaultSize = req.DefaultSize
	model.Enabled = req.Enabled
	model.SortOrder = req.SortOrder
	if err := s.repo.SaveImageModel(ctx, model); err != nil {
		return nil, errors.InternalServer(reason.DatabaseError).WithError(err)
	}
	return s.formatImageModel(ctx, model), nil
}

func (s *aiChatConfigService) DeleteImageModel(ctx context.Context, id int) error {
	if err := s.ensureImageTables(ctx); err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err)
	}
	if err := s.repo.DeleteImageModel(ctx, id); err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err)
	}
	return nil
}

func (s *aiChatConfigService) GetImageSetting(ctx context.Context) (*schema.AIImageSettingResp, error) {
	if err := s.ensureImageTables(ctx); err != nil {
		return nil, errors.InternalServer(reason.DatabaseError).WithError(err)
	}
	setting, exist, err := s.repo.GetImageSetting(ctx)
	if err != nil {
		return nil, errors.InternalServer(reason.DatabaseError).WithError(err)
	}
	if !exist {
		setting = &entity.AIImageSetting{ID: 1, RetentionDays: 30}
	}
	return s.formatImageSetting(setting), nil
}

func (s *aiChatConfigService) SaveImageSetting(ctx context.Context, req *schema.AIImageSettingReq) (*schema.AIImageSettingResp, error) {
	if err := s.ensureImageTables(ctx); err != nil {
		return nil, errors.InternalServer(reason.DatabaseError).WithError(err)
	}
	if req.RetentionDays < 1 || req.RetentionDays > 3650 {
		return nil, errors.BadRequest("retention_days must be between 1 and 3650")
	}
	setting := &entity.AIImageSetting{ID: 1, RetentionDays: req.RetentionDays}
	if err := s.repo.SaveImageSetting(ctx, setting); err != nil {
		return nil, errors.InternalServer(reason.DatabaseError).WithError(err)
	}
	return s.formatImageSetting(setting), nil
}

func (s *aiChatConfigService) GenerateImage(ctx context.Context, userID string, req *schema.AIImageGenerateReq) (*schema.AIImageGenerateResp, error) {
	if err := s.ensureImageTables(ctx); err != nil {
		return nil, errors.InternalServer(reason.DatabaseError).WithError(err)
	}
	if userID == "" {
		return nil, errors.Unauthorized(reason.UnauthorizedError)
	}
	req.Prompt = strings.TrimSpace(req.Prompt)
	req.Model = strings.TrimSpace(req.Model)
	if req.Prompt == "" || req.Model == "" {
		return nil, errors.BadRequest("prompt and model are required")
	}
	if req.Count <= 0 {
		req.Count = 1
	}
	if req.Count > 4 {
		return nil, errors.BadRequest("count cannot be greater than 4")
	}
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
	if plan.ImageQuota != -1 {
		monthStart, monthEnd := currentMonthRange()
		used, err := s.repo.CountUserImageGenerations(ctx, userID, monthStart, monthEnd)
		if err != nil {
			return nil, errors.InternalServer(reason.DatabaseError).WithError(err)
		}
		if used+req.Count > plan.ImageQuota {
			return nil, errors.BadRequest("image quota is insufficient")
		}
	}
	model, exist, err := s.repo.GetImageModelBySiteModelID(ctx, req.Model)
	if err != nil {
		return nil, errors.InternalServer(reason.DatabaseError).WithError(err)
	}
	if !exist || !model.Enabled {
		return nil, errors.BadRequest("image model is not available")
	}
	provider, exist, err := s.repo.GetImageProvider(ctx, model.ProviderID)
	if err != nil {
		return nil, errors.InternalServer(reason.DatabaseError).WithError(err)
	}
	if !exist || !provider.Enabled {
		return nil, errors.BadRequest("image provider is not available")
	}
	if strings.TrimSpace(req.Size) == "" {
		req.Size = model.DefaultSize
	}
	req.Size = normalizeOpenAIImageSize(req.Size, req.AspectRatio)
	req.AspectRatio = imageAspectRatio(req.Size)
	req.Quality = normalizeOpenAIImageQuality(req.Quality)
	generationID := "img_" + uid.IDStr()
	runCtx := context.WithoutCancel(ctx)
	setting, _ := s.GetImageSetting(runCtx)
	expiresAt := time.Now().AddDate(0, 0, setting.RetentionDays)
	pendingURLs, _ := json.Marshal([]string{})
	record := &entity.AIImageGeneration{
		GenerationID:    generationID,
		UserID:          userID,
		SiteModelID:     model.SiteModelID,
		ProviderID:      provider.ID,
		ProviderName:    provider.Name,
		ProviderModelID: model.ProviderModelID,
		Prompt:          req.Prompt,
		NegativePrompt:  req.NegativePrompt,
		AspectRatio:     req.AspectRatio,
		Size:            req.Size,
		Style:           req.Style,
		Quality:         req.Quality,
		Count:           req.Count,
		ImageURLs:       string(pendingURLs),
		Status:          "generating",
		ExpiresAt:       expiresAt,
	}
	if err := s.repo.CreateImageGeneration(runCtx, record); err != nil {
		log.Errorf("ai image generation pending record failed generation_id=%s user_id=%s error=%v", generationID, userID, err)
		return nil, errors.InternalServer(reason.DatabaseError).WithError(err)
	}
	log.Infof(
		"ai image generation start generation_id=%s user_id=%s site_model=%s provider=%s provider_model=%s size=%s aspect_ratio=%s quality=%s count=%d references=%d",
		generationID, userID, model.SiteModelID, provider.Name, model.ProviderModelID, req.Size, req.AspectRatio, req.Quality, req.Count, len(req.ReferenceImages),
	)
	imageURLs, err := s.callAndSaveImages(runCtx, provider, model, generationID, userID, req)
	if err != nil {
		log.Errorf(
			"ai image generation failed generation_id=%s user_id=%s site_model=%s provider=%s provider_model=%s references=%d error=%v",
			generationID, userID, model.SiteModelID, provider.Name, model.ProviderModelID, len(req.ReferenceImages), err,
		)
		updateErr := s.repo.UpdateImageGeneration(runCtx, generationID, &entity.AIImageGeneration{
			Status: "failed",
			Error:  err.Error(),
		}, "status", "error")
		if updateErr != nil {
			log.Errorf("ai image generation failed status update failed generation_id=%s user_id=%s error=%v", generationID, userID, updateErr)
		}
		return nil, errors.BadRequest(fmt.Sprintf("failed to generate image: %s", err.Error()))
	}
	rawURLs, _ := json.Marshal(imageURLs)
	if err := s.repo.UpdateImageGeneration(runCtx, generationID, &entity.AIImageGeneration{
		Count:     len(imageURLs),
		ImageURLs: string(rawURLs),
		Status:    "completed",
		Error:     "",
	}, "count", "image_urls", "status", "error"); err != nil {
		log.Errorf("ai image generation record update failed generation_id=%s user_id=%s error=%v", generationID, userID, err)
		return nil, errors.InternalServer(reason.DatabaseError).WithError(err)
	}
	log.Infof("ai image generation completed generation_id=%s user_id=%s image_count=%d", generationID, userID, len(imageURLs))
	return &schema.AIImageGenerateResp{
		GenerationID: generationID,
		Size:         req.Size,
		ImageURLs:    imageURLs,
		ExpiresAt:    expiresAt.Unix(),
	}, nil
}

func (s *aiChatConfigService) EditImage(ctx context.Context, userID string, req *schema.AIImageEditReq) (*schema.AIImageGenerateResp, error) {
	if strings.TrimSpace(req.Prompt) == "" || strings.TrimSpace(req.ImageURL) == "" || strings.TrimSpace(req.Model) == "" {
		return nil, errors.BadRequest("prompt, image_url and model are required")
	}
	reference, err := s.loadUserGeneratedImage(userID, req.ImageURL)
	if err != nil {
		log.Errorf("ai image edit reference load failed user_id=%s image_url=%s error=%v", userID, req.ImageURL, err)
		return nil, errors.BadRequest("image is not editable")
	}
	return s.GenerateImage(ctx, userID, &schema.AIImageGenerateReq{
		Prompt:          req.Prompt,
		Model:           req.Model,
		Size:            req.Size,
		Quality:         req.Quality,
		Count:           1,
		ReferenceImages: []string{reference.DataURL},
	})
}

func (s *aiChatConfigService) ListUserImageGenerations(ctx context.Context, userID string, limit int) ([]*schema.AIImageGenerationResp, error) {
	if err := s.ensureImageTables(ctx); err != nil {
		return nil, errors.InternalServer(reason.DatabaseError).WithError(err)
	}
	records, err := s.repo.ListUserImageGenerations(ctx, userID, limit)
	if err != nil {
		return nil, errors.InternalServer(reason.DatabaseError).WithError(err)
	}
	resp := make([]*schema.AIImageGenerationResp, 0, len(records))
	for _, record := range records {
		resp = append(resp, s.formatImageGeneration(record))
	}
	return resp, nil
}

func (s *aiChatConfigService) GetUserImageFilePath(ctx context.Context, userID, ownerID, filename string) (string, error) {
	if userID == "" {
		return "", errors.Unauthorized(reason.UnauthorizedError)
	}
	if userID != ownerID {
		return "", errors.BadRequest(reason.ForbiddenError)
	}
	if s.serviceConfig == nil || strings.TrimSpace(s.serviceConfig.UploadPath) == "" {
		return "", errors.InternalServer(reason.UnknownError).WithError(fmt.Errorf("upload path is not configured"))
	}
	filename = path.Base(strings.TrimSpace(filename))
	if filename == "." || filename == "/" || strings.Contains(filename, "..") {
		return "", errors.BadRequest(reason.RequestFormatError)
	}
	filePath := filepath.Join(s.serviceConfig.UploadPath, constant.AIImageSubPath, userID, filename)
	info, err := os.Stat(filePath)
	if err != nil || info.IsDir() {
		return "", errors.BadRequest(reason.ObjectNotFound)
	}
	return filePath, nil
}

func (s *aiChatConfigService) ListVideoProviders(ctx context.Context) ([]*schema.AIVideoProviderResp, error) {
	if err := s.ensureVideoTables(ctx); err != nil {
		return nil, errors.InternalServer(reason.DatabaseError).WithError(err)
	}
	providers, err := s.repo.ListVideoProviders(ctx)
	if err != nil {
		return nil, errors.InternalServer(reason.DatabaseError).WithError(err)
	}
	resp := make([]*schema.AIVideoProviderResp, 0, len(providers))
	for _, provider := range providers {
		resp = append(resp, s.formatVideoProvider(provider, true))
	}
	return resp, nil
}

func (s *aiChatConfigService) CreateVideoProvider(ctx context.Context, req *schema.AIVideoProviderReq) (*schema.AIVideoProviderResp, error) {
	if err := s.ensureVideoTables(ctx); err != nil {
		return nil, errors.InternalServer(reason.DatabaseError).WithError(err)
	}
	if strings.TrimSpace(req.APIKey) == "" {
		return nil, errors.BadRequest("api_key is required")
	}
	baseURL, err := normalizeBaseURL(req.BaseURL)
	if err != nil {
		return nil, errors.BadRequest("base_url is invalid")
	}
	provider := &entity.AIVideoProvider{
		Name:    strings.TrimSpace(req.Name),
		BaseURL: baseURL,
		APIKey:  strings.TrimSpace(req.APIKey),
		Enabled: req.Enabled,
		Remark:  req.Remark,
	}
	if err := s.repo.CreateVideoProvider(ctx, provider); err != nil {
		return nil, errors.InternalServer(reason.DatabaseError).WithError(err)
	}
	return s.formatVideoProvider(provider, true), nil
}

func (s *aiChatConfigService) UpdateVideoProvider(ctx context.Context, id int, req *schema.AIVideoProviderReq) (*schema.AIVideoProviderResp, error) {
	if err := s.ensureVideoTables(ctx); err != nil {
		return nil, errors.InternalServer(reason.DatabaseError).WithError(err)
	}
	provider, exist, err := s.repo.GetVideoProvider(ctx, id)
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
	provider.Remark = req.Remark
	cols := []string{"name", "base_url", "enabled", "remark"}
	if strings.TrimSpace(req.APIKey) != "" && !isAllMask(req.APIKey) {
		provider.APIKey = strings.TrimSpace(req.APIKey)
		cols = append(cols, "api_key")
	}
	if err := s.repo.UpdateVideoProvider(ctx, provider, cols...); err != nil {
		return nil, errors.InternalServer(reason.DatabaseError).WithError(err)
	}
	return s.formatVideoProvider(provider, true), nil
}

func (s *aiChatConfigService) DeleteVideoProvider(ctx context.Context, id int) error {
	if err := s.ensureVideoTables(ctx); err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err)
	}
	if err := s.repo.DeleteVideoProvider(ctx, id); err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err)
	}
	return nil
}

func (s *aiChatConfigService) ListVideoModels(ctx context.Context, onlyEnabled bool) ([]*schema.AIVideoModelResp, error) {
	if err := s.ensureVideoTables(ctx); err != nil {
		return nil, errors.InternalServer(reason.DatabaseError).WithError(err)
	}
	models, err := s.repo.ListVideoModels(ctx, onlyEnabled)
	if err != nil {
		return nil, errors.InternalServer(reason.DatabaseError).WithError(err)
	}
	resp := make([]*schema.AIVideoModelResp, 0, len(models))
	for _, model := range models {
		resp = append(resp, s.formatVideoModel(ctx, model))
	}
	return resp, nil
}

func (s *aiChatConfigService) SaveVideoModel(ctx context.Context, id int, req *schema.AIVideoModelReq) (*schema.AIVideoModelResp, error) {
	if err := s.ensureVideoTables(ctx); err != nil {
		return nil, errors.InternalServer(reason.DatabaseError).WithError(err)
	}
	if req.ProviderID <= 0 {
		return nil, errors.BadRequest("provider_id is required")
	}
	if !modelIDPattern.MatchString(strings.TrimSpace(req.SiteModelID)) {
		return nil, errors.BadRequest("site_model_id can only contain lowercase letters, numbers, hyphen and underscore")
	}
	if strings.TrimSpace(req.ProviderModelID) == "" {
		return nil, errors.BadRequest("provider_model_id is required")
	}
	if strings.TrimSpace(req.DefaultSize) == "" {
		req.DefaultSize = "1280x720"
	}
	if req.DefaultSeconds == 0 {
		req.DefaultSeconds = 6
	}
	if err := validateVideoSeconds(req.DefaultSeconds); err != nil {
		return nil, err
	}
	if strings.TrimSpace(req.DefaultResolution) == "" {
		req.DefaultResolution = "720p"
	}
	if strings.TrimSpace(req.DefaultPreset) == "" {
		req.DefaultPreset = "custom"
	}
	if err := validateVideoRequestOptions(req.DefaultSize, req.DefaultResolution, req.DefaultPreset); err != nil {
		return nil, err
	}
	if _, exist, err := s.repo.GetVideoProvider(ctx, req.ProviderID); err != nil {
		return nil, errors.InternalServer(reason.DatabaseError).WithError(err)
	} else if !exist {
		return nil, errors.BadRequest("provider is not available")
	}
	model := &entity.AIVideoModel{}
	if id > 0 {
		current, exist, err := s.repo.GetVideoModel(ctx, id)
		if err != nil {
			return nil, errors.InternalServer(reason.DatabaseError).WithError(err)
		}
		if !exist {
			return nil, errors.BadRequest(reason.ObjectNotFound)
		}
		model = current
	}
	model.ProviderID = req.ProviderID
	model.SiteModelID = strings.TrimSpace(req.SiteModelID)
	model.ProviderModelID = strings.TrimSpace(req.ProviderModelID)
	model.DisplayName = req.DisplayName
	model.Description = req.Description
	model.DefaultSize = req.DefaultSize
	model.DefaultSeconds = req.DefaultSeconds
	model.DefaultResolution = req.DefaultResolution
	model.DefaultPreset = req.DefaultPreset
	model.Enabled = req.Enabled
	model.SortOrder = req.SortOrder
	if err := s.repo.SaveVideoModel(ctx, model); err != nil {
		return nil, errors.InternalServer(reason.DatabaseError).WithError(err)
	}
	return s.formatVideoModel(ctx, model), nil
}

func (s *aiChatConfigService) DeleteVideoModel(ctx context.Context, id int) error {
	if err := s.ensureVideoTables(ctx); err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err)
	}
	if err := s.repo.DeleteVideoModel(ctx, id); err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err)
	}
	return nil
}

func (s *aiChatConfigService) GetVideoSetting(ctx context.Context) (*schema.AIVideoSettingResp, error) {
	if err := s.ensureVideoTables(ctx); err != nil {
		return nil, errors.InternalServer(reason.DatabaseError).WithError(err)
	}
	setting, exist, err := s.repo.GetVideoSetting(ctx)
	if err != nil {
		return nil, errors.InternalServer(reason.DatabaseError).WithError(err)
	}
	if !exist {
		setting = &entity.AIVideoSetting{ID: 1, RetentionDays: 30}
	}
	return s.formatVideoSetting(setting), nil
}

func (s *aiChatConfigService) SaveVideoSetting(ctx context.Context, req *schema.AIVideoSettingReq) (*schema.AIVideoSettingResp, error) {
	if err := s.ensureVideoTables(ctx); err != nil {
		return nil, errors.InternalServer(reason.DatabaseError).WithError(err)
	}
	if req.RetentionDays < 1 || req.RetentionDays > 3650 {
		return nil, errors.BadRequest("retention_days must be between 1 and 3650")
	}
	setting := &entity.AIVideoSetting{ID: 1, RetentionDays: req.RetentionDays}
	if err := s.repo.SaveVideoSetting(ctx, setting); err != nil {
		return nil, errors.InternalServer(reason.DatabaseError).WithError(err)
	}
	return s.formatVideoSetting(setting), nil
}

func (s *aiChatConfigService) GenerateVideo(ctx context.Context, userID string, req *schema.AIVideoGenerateReq) (*schema.AIVideoGenerateResp, error) {
	if err := s.ensureVideoTables(ctx); err != nil {
		return nil, errors.InternalServer(reason.DatabaseError).WithError(err)
	}
	if userID == "" {
		return nil, errors.Unauthorized(reason.UnauthorizedError)
	}
	req.Prompt = strings.TrimSpace(req.Prompt)
	req.Model = strings.TrimSpace(req.Model)
	if req.Prompt == "" || req.Model == "" {
		return nil, errors.BadRequest("prompt and model are required")
	}
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
	if err := s.checkVideoQuota(ctx, userID, plan); err != nil {
		return nil, err
	}
	model, exist, err := s.repo.GetVideoModelBySiteModelID(ctx, req.Model)
	if err != nil {
		return nil, errors.InternalServer(reason.DatabaseError).WithError(err)
	}
	if !exist || !model.Enabled {
		return nil, errors.BadRequest("video model is not available")
	}
	provider, exist, err := s.repo.GetVideoProvider(ctx, model.ProviderID)
	if err != nil {
		return nil, errors.InternalServer(reason.DatabaseError).WithError(err)
	}
	if !exist || !provider.Enabled {
		return nil, errors.BadRequest("video provider is not available")
	}
	if strings.TrimSpace(req.Size) == "" {
		req.Size = model.DefaultSize
	}
	if req.Seconds == 0 {
		req.Seconds = model.DefaultSeconds
	}
	if strings.TrimSpace(req.Quality) == "" {
		req.Quality = model.DefaultResolution
	}
	if strings.TrimSpace(req.Preset) == "" {
		req.Preset = model.DefaultPreset
	}
	if err := validateVideoSeconds(req.Seconds); err != nil {
		return nil, err
	}
	if err := validateVideoRequestOptions(req.Size, req.Quality, req.Preset); err != nil {
		return nil, err
	}

	generationID := "vid_" + uid.IDStr()
	runCtx := context.WithoutCancel(ctx)
	setting, _ := s.GetVideoSetting(runCtx)
	expiresAt := time.Now().AddDate(0, 0, setting.RetentionDays)
	referenceImages, _ := json.Marshal(req.ReferenceImages)
	record := &entity.AIVideoGeneration{
		GenerationID:    generationID,
		UserID:          userID,
		SiteModelID:     model.SiteModelID,
		ProviderID:      provider.ID,
		ProviderName:    provider.Name,
		ProviderModelID: model.ProviderModelID,
		Prompt:          req.Prompt,
		AspectRatio:     videoAspectRatio(req.Size),
		Size:            req.Size,
		Quality:         req.Quality,
		Seconds:         req.Seconds,
		Preset:          req.Preset,
		ReferenceImages: string(referenceImages),
		Status:          "queued",
		Progress:        0,
		ExpiresAt:       expiresAt,
	}
	if err := s.repo.CreateVideoGeneration(runCtx, record); err != nil {
		log.Errorf("ai video generation pending record failed generation_id=%s user_id=%s error=%v", generationID, userID, err)
		return nil, errors.InternalServer(reason.DatabaseError).WithError(err)
	}
	log.Infof(
		"ai video generation accepted generation_id=%s user_id=%s site_model=%s provider=%s provider_model=%s base_url=%s size=%s seconds=%d quality=%s preset=%s references=%d",
		generationID, userID, model.SiteModelID, provider.Name, model.ProviderModelID, strings.TrimRight(provider.BaseURL, "/"), req.Size, req.Seconds, req.Quality, req.Preset, len(req.ReferenceImages),
	)
	go s.runVideoGeneration(runCtx, provider, model, generationID, userID, req)
	return &schema.AIVideoGenerateResp{
		GenerationID: generationID,
		Status:       record.Status,
		Progress:     record.Progress,
		Size:         req.Size,
		Seconds:      req.Seconds,
		ExpiresAt:    expiresAt.Unix(),
	}, nil
}

func (s *aiChatConfigService) ListUserVideoGenerations(ctx context.Context, userID string, limit int) ([]*schema.AIVideoGenerationResp, error) {
	if err := s.ensureVideoTables(ctx); err != nil {
		return nil, errors.InternalServer(reason.DatabaseError).WithError(err)
	}
	records, err := s.repo.ListUserVideoGenerations(ctx, userID, limit)
	if err != nil {
		return nil, errors.InternalServer(reason.DatabaseError).WithError(err)
	}
	resp := make([]*schema.AIVideoGenerationResp, 0, len(records))
	for _, record := range records {
		resp = append(resp, s.formatVideoGeneration(record))
	}
	return resp, nil
}

func (s *aiChatConfigService) GetUserVideoFilePath(ctx context.Context, userID, ownerID, filename string) (string, error) {
	if userID == "" {
		return "", errors.Unauthorized(reason.UnauthorizedError)
	}
	if userID != ownerID {
		return "", errors.BadRequest(reason.ForbiddenError)
	}
	if s.serviceConfig == nil || strings.TrimSpace(s.serviceConfig.UploadPath) == "" {
		return "", errors.InternalServer(reason.UnknownError).WithError(fmt.Errorf("upload path is not configured"))
	}
	filename = path.Base(strings.TrimSpace(filename))
	if filename == "." || filename == "/" || strings.Contains(filename, "..") {
		return "", errors.BadRequest(reason.RequestFormatError)
	}
	filePath := filepath.Join(s.serviceConfig.UploadPath, constant.AIVideoSubPath, userID, filename)
	info, err := os.Stat(filePath)
	if err != nil || info.IsDir() {
		return "", errors.BadRequest(reason.ObjectNotFound)
	}
	return filePath, nil
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

func (s *aiChatConfigService) ensureImageTables(ctx context.Context) error {
	return s.repo.EnsureImageTables(ctx)
}

func (s *aiChatConfigService) ensureVideoTables(ctx context.Context) error {
	return s.repo.EnsureVideoTables(ctx)
}

func (s *aiChatConfigService) validatePlanReq(ctx context.Context, excludeID int, req *schema.AISubscriptionPlanReq) error {
	req.PlanID = strings.TrimSpace(req.PlanID)
	if !modelIDPattern.MatchString(req.PlanID) {
		return errors.BadRequest("plan_id can only contain lowercase letters, numbers, hyphen and underscore")
	}
	if req.MonthlyPrice < 0 || req.ChatPoints < -1 || req.ImageQuota < -1 || req.VideoDailyQuota < -1 || req.VideoQuota < -1 {
		return errors.BadRequest("monthly_price cannot be negative, quotas must be -1 or greater")
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
		VideoDailyQuota:   plan.VideoDailyQuota,
		VideoQuota:        plan.VideoQuota,
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

func (s *aiChatConfigService) formatImageProvider(provider *entity.AIImageProvider, mask bool) *schema.AIImageProviderResp {
	apiKey := provider.APIKey
	if mask {
		apiKey = maskSecret(apiKey)
	}
	return &schema.AIImageProviderResp{
		ID:        provider.ID,
		Name:      provider.Name,
		BaseURL:   provider.BaseURL,
		APIKey:    apiKey,
		Enabled:   provider.Enabled,
		Remark:    provider.Remark,
		CreatedAt: provider.CreatedAt.Unix(),
		UpdatedAt: provider.UpdatedAt.Unix(),
	}
}

func (s *aiChatConfigService) formatImageModel(ctx context.Context, model *entity.AIImageModel) *schema.AIImageModelResp {
	providerName := ""
	if provider, exist, _ := s.repo.GetImageProvider(ctx, model.ProviderID); exist {
		providerName = provider.Name
	}
	return &schema.AIImageModelResp{
		ID:              model.ID,
		ProviderID:      model.ProviderID,
		ProviderName:    providerName,
		SiteModelID:     model.SiteModelID,
		ProviderModelID: model.ProviderModelID,
		DisplayName:     fallbackText(model.DisplayName, model.SiteModelID),
		Description:     model.Description,
		DefaultSize:     model.DefaultSize,
		Enabled:         model.Enabled,
		SortOrder:       model.SortOrder,
		CreatedAt:       model.CreatedAt.Unix(),
		UpdatedAt:       model.UpdatedAt.Unix(),
	}
}

func (s *aiChatConfigService) formatImageSetting(setting *entity.AIImageSetting) *schema.AIImageSettingResp {
	return &schema.AIImageSettingResp{
		RetentionDays: setting.RetentionDays,
		CreatedAt:     unixOrZero(setting.CreatedAt),
		UpdatedAt:     unixOrZero(setting.UpdatedAt),
	}
}

func (s *aiChatConfigService) formatImageGeneration(record *entity.AIImageGeneration) *schema.AIImageGenerationResp {
	imageURLs := make([]string, 0)
	_ = json.Unmarshal([]byte(record.ImageURLs), &imageURLs)
	return &schema.AIImageGenerationResp{
		ID:              record.ID,
		GenerationID:    record.GenerationID,
		UserID:          record.UserID,
		SiteModelID:     record.SiteModelID,
		ProviderID:      record.ProviderID,
		ProviderName:    record.ProviderName,
		ProviderModelID: record.ProviderModelID,
		Prompt:          record.Prompt,
		NegativePrompt:  record.NegativePrompt,
		AspectRatio:     record.AspectRatio,
		Size:            record.Size,
		Style:           record.Style,
		Quality:         record.Quality,
		Count:           record.Count,
		ImageURLs:       imageURLs,
		Status:          record.Status,
		Error:           record.Error,
		ExpiresAt:       unixOrZero(record.ExpiresAt),
		CreatedAt:       unixOrZero(record.CreatedAt),
		UpdatedAt:       unixOrZero(record.UpdatedAt),
	}
}

func (s *aiChatConfigService) formatVideoProvider(provider *entity.AIVideoProvider, mask bool) *schema.AIVideoProviderResp {
	apiKey := provider.APIKey
	if mask {
		apiKey = maskSecret(apiKey)
	}
	return &schema.AIVideoProviderResp{
		ID:        provider.ID,
		Name:      provider.Name,
		BaseURL:   provider.BaseURL,
		APIKey:    apiKey,
		Enabled:   provider.Enabled,
		Remark:    provider.Remark,
		CreatedAt: provider.CreatedAt.Unix(),
		UpdatedAt: provider.UpdatedAt.Unix(),
	}
}

func (s *aiChatConfigService) formatVideoModel(ctx context.Context, model *entity.AIVideoModel) *schema.AIVideoModelResp {
	providerName := ""
	if provider, exist, _ := s.repo.GetVideoProvider(ctx, model.ProviderID); exist {
		providerName = provider.Name
	}
	return &schema.AIVideoModelResp{
		ID:                model.ID,
		ProviderID:        model.ProviderID,
		ProviderName:      providerName,
		SiteModelID:       model.SiteModelID,
		ProviderModelID:   model.ProviderModelID,
		DisplayName:       fallbackText(model.DisplayName, model.SiteModelID),
		Description:       model.Description,
		DefaultSize:       model.DefaultSize,
		DefaultSeconds:    model.DefaultSeconds,
		DefaultResolution: model.DefaultResolution,
		DefaultPreset:     model.DefaultPreset,
		Enabled:           model.Enabled,
		SortOrder:         model.SortOrder,
		CreatedAt:         model.CreatedAt.Unix(),
		UpdatedAt:         model.UpdatedAt.Unix(),
	}
}

func (s *aiChatConfigService) formatVideoSetting(setting *entity.AIVideoSetting) *schema.AIVideoSettingResp {
	return &schema.AIVideoSettingResp{
		RetentionDays: setting.RetentionDays,
		CreatedAt:     unixOrZero(setting.CreatedAt),
		UpdatedAt:     unixOrZero(setting.UpdatedAt),
	}
}

func (s *aiChatConfigService) formatVideoGeneration(record *entity.AIVideoGeneration) *schema.AIVideoGenerationResp {
	referenceImages := make([]string, 0)
	_ = json.Unmarshal([]byte(record.ReferenceImages), &referenceImages)
	return &schema.AIVideoGenerationResp{
		ID:              record.ID,
		GenerationID:    record.GenerationID,
		UpstreamID:      record.UpstreamID,
		UserID:          record.UserID,
		SiteModelID:     record.SiteModelID,
		ProviderID:      record.ProviderID,
		ProviderName:    record.ProviderName,
		ProviderModelID: record.ProviderModelID,
		Prompt:          record.Prompt,
		AspectRatio:     record.AspectRatio,
		Size:            record.Size,
		Quality:         record.Quality,
		Seconds:         record.Seconds,
		Preset:          record.Preset,
		ReferenceImages: referenceImages,
		VideoURL:        record.VideoURL,
		Status:          record.Status,
		Progress:        record.Progress,
		Error:           record.Error,
		ExpiresAt:       unixOrZero(record.ExpiresAt),
		CreatedAt:       unixOrZero(record.CreatedAt),
		UpdatedAt:       unixOrZero(record.UpdatedAt),
	}
}

func (s *aiChatConfigService) callAndSaveImages(
	ctx context.Context,
	provider *entity.AIImageProvider,
	model *entity.AIImageModel,
	generationID string,
	userID string,
	req *schema.AIImageGenerateReq,
) ([]string, error) {
	if req.Count > 1 {
		log.Infof("ai image batch split generation_id=%s requested_count=%d", generationID, req.Count)
		imageURLs := make([]string, 0, req.Count)
		singleReq := *req
		singleReq.Count = 1
		for i := 0; i < req.Count; i++ {
			partGenerationID := fmt.Sprintf("%s_%d", generationID, i)
			partURLs, err := s.callAndSaveImages(ctx, provider, model, partGenerationID, userID, &singleReq)
			if err != nil {
				return nil, err
			}
			imageURLs = append(imageURLs, partURLs...)
		}
		return imageURLs, nil
	}

	baseURL := strings.TrimRight(provider.BaseURL, "/")
	prompt := buildImagePrompt(req)
	payload := map[string]any{
		"model":  model.ProviderModelID,
		"prompt": prompt,
		"size":   req.Size,
	}
	if shouldRequestImageResponseFormat(model.ProviderModelID) {
		payload["response_format"] = "b64_json"
	}
	if req.Count > 1 {
		payload["n"] = req.Count
	}
	if req.Quality != "" {
		payload["quality"] = req.Quality
	}
	if len(req.ReferenceImages) > 0 {
		log.Infof("ai image preparing references generation_id=%s raw_reference_count=%d", generationID, len(req.ReferenceImages))
		preparedImages, err := prepareReferenceImages(ctx, req.ReferenceImages)
		if err != nil {
			log.Errorf("ai image reference prepare failed generation_id=%s error=%v", generationID, err)
			return nil, err
		}
		if len(preparedImages) == 0 {
			log.Errorf("ai image reference prepare failed generation_id=%s error=empty reference image", generationID)
			return nil, fmt.Errorf("reference image is empty")
		}
		log.Infof("ai image prepared references generation_id=%s %s", generationID, summarizeReferenceImages(preparedImages))
		return s.callAndSaveImagesWithReferences(
			ctx, baseURL, provider, model, generationID, userID, req, prompt, payload, preparedImages,
		)
	}
	log.Infof("ai image upstream request generation_id=%s endpoint=%s model=%s size=%s count=%d references=0", generationID, "/images/generations", model.ProviderModelID, req.Size, req.Count)
	resp, err := resty.New().
		SetRetryCount(1).
		SetHeader("Authorization", fmt.Sprintf("Bearer %s", provider.APIKey)).
		SetHeader("Content-Type", "application/json").
		R().
		SetContext(ctx).
		SetBody(payload).
		Post(baseURL + "/images/generations")
	if err != nil {
		log.Errorf("ai image upstream request failed generation_id=%s endpoint=%s error=%v", generationID, "/images/generations", err)
		return nil, err
	}
	if !resp.IsSuccess() {
		log.Errorf("ai image upstream non-success generation_id=%s endpoint=%s status=%d body=%s", generationID, "/images/generations", resp.StatusCode(), responseSnippet(resp.Body()))
		return nil, fmt.Errorf("status %d: %s", resp.StatusCode(), resp.String())
	}
	log.Infof("ai image upstream success generation_id=%s endpoint=%s status=%d bytes=%d", generationID, "/images/generations", resp.StatusCode(), len(resp.Body()))
	return s.saveImageAPIResponse(ctx, userID, generationID, resp.Body())
}

type preparedReferenceImage struct {
	Data    []byte
	MIME    string
	Ext     string
	DataURL string
}

func (s *aiChatConfigService) callAndSaveImagesWithReferences(
	ctx context.Context,
	baseURL string,
	provider *entity.AIImageProvider,
	model *entity.AIImageModel,
	generationID string,
	userID string,
	req *schema.AIImageGenerateReq,
	prompt string,
	basePayload map[string]any,
	referenceImages []*preparedReferenceImage,
) ([]string, error) {
	client := resty.New().SetRetryCount(1)
	var lastErr error

	if imageURLs, err := s.callImageEditsAPI(ctx, client, baseURL, provider, generationID, userID, req, prompt, basePayload, referenceImages, false); err == nil {
		log.Infof("ai image reference generation succeeded generation_id=%s strategy=images_edits field=image image_count=%d", generationID, len(imageURLs))
		return imageURLs, nil
	} else {
		log.Warnf("ai image reference generation attempt failed generation_id=%s strategy=images_edits field=image error=%v", generationID, err)
		lastErr = err
	}

	if len(referenceImages) > 1 {
		if imageURLs, err := s.callImageEditsAPI(ctx, client, baseURL, provider, generationID, userID, req, prompt, basePayload, referenceImages, true); err == nil {
			log.Infof("ai image reference generation succeeded generation_id=%s strategy=images_edits field=image_array image_count=%d", generationID, len(imageURLs))
			return imageURLs, nil
		} else {
			log.Warnf("ai image reference generation attempt failed generation_id=%s strategy=images_edits field=image_array error=%v", generationID, err)
			lastErr = err
		}
	}

	if imageURLs, err := s.callResponsesImageAPI(ctx, client, baseURL, provider, model, generationID, userID, req, prompt, referenceImages); err == nil {
		log.Infof("ai image reference generation succeeded generation_id=%s strategy=responses image_count=%d", generationID, len(imageURLs))
		return imageURLs, nil
	} else {
		log.Warnf("ai image reference generation attempt failed generation_id=%s strategy=responses error=%v", generationID, err)
		lastErr = err
	}

	return nil, fmt.Errorf("reference image generation failed: %w", lastErr)
}

func (s *aiChatConfigService) callResponsesImageAPI(
	ctx context.Context,
	client *resty.Client,
	baseURL string,
	provider *entity.AIImageProvider,
	model *entity.AIImageModel,
	generationID string,
	userID string,
	req *schema.AIImageGenerateReq,
	prompt string,
	referenceImages []*preparedReferenceImage,
) ([]string, error) {
	content := []map[string]any{{"type": "input_text", "text": prompt}}
	for _, image := range referenceImages {
		content = append(content, map[string]any{
			"type":      "input_image",
			"image_url": image.DataURL,
		})
	}
	tool := map[string]any{
		"type": "image_generation",
	}
	if strings.TrimSpace(req.Size) != "" && req.Size != "auto" {
		tool["size"] = req.Size
	}
	if req.Quality != "" {
		tool["quality"] = req.Quality
	}
	body := map[string]any{
		"model":  "gpt-5.5",
		"stream": false,
		"input": []map[string]any{
			{
				"role":    "user",
				"content": content,
			},
		},
		"tools": []map[string]any{tool},
	}
	resp, err := client.R().
		SetContext(ctx).
		SetHeader("Authorization", fmt.Sprintf("Bearer %s", provider.APIKey)).
		SetHeader("Content-Type", "application/json").
		SetBody(body).
		Post(baseURL + "/responses")
	if err != nil {
		log.Errorf("ai image upstream request failed generation_id=%s endpoint=%s model=%s references=%d error=%v", generationID, "/responses", "gpt-5.5", len(referenceImages), err)
		return nil, err
	}
	if !resp.IsSuccess() {
		log.Errorf("ai image upstream non-success generation_id=%s endpoint=%s model=%s references=%d status=%d body=%s", generationID, "/responses", "gpt-5.5", len(referenceImages), resp.StatusCode(), responseSnippet(resp.Body()))
		return nil, fmt.Errorf("responses status %d: %s", resp.StatusCode(), resp.String())
	}
	log.Infof("ai image upstream success generation_id=%s endpoint=%s model=%s references=%d status=%d bytes=%d", generationID, "/responses", "gpt-5.5", len(referenceImages), resp.StatusCode(), len(resp.Body()))
	return s.saveImageAPIResponse(ctx, userID, generationID, resp.Body())
}

func (s *aiChatConfigService) callImageEditsAPI(
	ctx context.Context,
	client *resty.Client,
	baseURL string,
	provider *entity.AIImageProvider,
	generationID string,
	userID string,
	req *schema.AIImageGenerateReq,
	prompt string,
	basePayload map[string]any,
	referenceImages []*preparedReferenceImage,
	useArrayField bool,
) ([]string, error) {
	formData := map[string]string{
		"model":  fmt.Sprint(basePayload["model"]),
		"prompt": prompt,
	}
	if shouldRequestImageResponseFormat(fmt.Sprint(basePayload["model"])) {
		formData["response_format"] = "b64_json"
	}
	if strings.TrimSpace(req.Size) != "" {
		formData["size"] = req.Size
	}
	if req.Count > 1 {
		formData["n"] = fmt.Sprint(req.Count)
	}
	if req.Quality != "" {
		formData["quality"] = req.Quality
	}
	request := client.R().
		SetContext(ctx).
		SetHeader("Authorization", fmt.Sprintf("Bearer %s", provider.APIKey)).
		SetFormData(formData)
	for index, image := range referenceImages {
		fieldName := "image"
		if useArrayField && index > 0 {
			fieldName = "image[]"
		}
		request.SetFileReader(fieldName, fmt.Sprintf("reference-%d%s", index+1, image.Ext), bytes.NewReader(image.Data))
	}
	fieldMode := "image"
	if useArrayField {
		fieldMode = "image_array"
	}
	log.Infof("ai image upstream request generation_id=%s endpoint=%s model=%s references=%d field_mode=%s size=%s count=%d", generationID, "/images/edits", formData["model"], len(referenceImages), fieldMode, formData["size"], req.Count)
	resp, err := request.Post(baseURL + "/images/edits")
	if err != nil {
		log.Errorf("ai image upstream request failed generation_id=%s endpoint=%s field_mode=%s error=%v", generationID, "/images/edits", fieldMode, err)
		return nil, err
	}
	if !resp.IsSuccess() {
		log.Errorf("ai image upstream non-success generation_id=%s endpoint=%s field_mode=%s status=%d body=%s", generationID, "/images/edits", fieldMode, resp.StatusCode(), responseSnippet(resp.Body()))
		return nil, fmt.Errorf("edits status %d: %s", resp.StatusCode(), resp.String())
	}
	log.Infof("ai image upstream success generation_id=%s endpoint=%s field_mode=%s status=%d bytes=%d", generationID, "/images/edits", fieldMode, resp.StatusCode(), len(resp.Body()))
	return s.saveImageAPIResponse(ctx, userID, generationID, resp.Body())
}

func (s *aiChatConfigService) saveImageAPIResponse(ctx context.Context, userID, generationID string, body []byte) ([]string, error) {
	var parsed struct {
		Data []struct {
			URL     string `json:"url"`
			B64JSON string `json:"b64_json"`
		} `json:"data"`
		Output []struct {
			Type    string `json:"type"`
			Result  string `json:"result"`
			Content []struct {
				Result   string `json:"result"`
				ImageURL string `json:"image_url"`
			} `json:"content"`
		} `json:"output"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		log.Errorf("ai image response parse failed generation_id=%s bytes=%d body=%s error=%v", generationID, len(body), responseSnippet(body), err)
		return nil, err
	}
	for _, item := range parsed.Output {
		if item.Type == "image_generation_call" && item.Result != "" {
			parsed.Data = append(parsed.Data, struct {
				URL     string `json:"url"`
				B64JSON string `json:"b64_json"`
			}{B64JSON: item.Result})
		}
		for _, entry := range item.Content {
			if entry.Result != "" {
				parsed.Data = append(parsed.Data, struct {
					URL     string `json:"url"`
					B64JSON string `json:"b64_json"`
				}{B64JSON: entry.Result})
			}
			if entry.ImageURL != "" {
				parsed.Data = append(parsed.Data, struct {
					URL     string `json:"url"`
					B64JSON string `json:"b64_json"`
				}{URL: entry.ImageURL})
			}
		}
	}
	if len(parsed.Data) == 0 {
		log.Errorf("ai image response empty generation_id=%s bytes=%d body=%s", generationID, len(body), responseSnippet(body))
		return nil, fmt.Errorf("empty image response")
	}
	imageURLs := make([]string, 0, len(parsed.Data))
	for i, item := range parsed.Data {
		var data []byte
		ext := ".png"
		var err error
		if item.B64JSON != "" {
			data, ext, err = decodeImageData(item.B64JSON)
			if err != nil {
				log.Errorf("ai image response b64 decode failed generation_id=%s index=%d error=%v", generationID, i, err)
				return nil, err
			}
		} else if item.URL != "" {
			data, ext, err = downloadImage(ctx, item.URL)
			if err != nil {
				log.Errorf("ai image response image download failed generation_id=%s index=%d error=%v", generationID, i, err)
				return nil, err
			}
		} else {
			continue
		}
		url, err := s.saveGeneratedImage(userID, generationID, i, ext, data)
		if err != nil {
			log.Errorf("ai image save file failed generation_id=%s index=%d ext=%s bytes=%d error=%v", generationID, i, ext, len(data), err)
			return nil, err
		}
		imageURLs = append(imageURLs, url)
	}
	if len(imageURLs) == 0 {
		log.Errorf("ai image response no usable data generation_id=%s bytes=%d body=%s", generationID, len(body), responseSnippet(body))
		return nil, fmt.Errorf("no image data in response")
	}
	log.Infof("ai image response saved generation_id=%s image_count=%d", generationID, len(imageURLs))
	return imageURLs, nil
}

func (s *aiChatConfigService) saveGeneratedImage(userID, generationID string, index int, ext string, data []byte) (string, error) {
	if s.serviceConfig == nil || strings.TrimSpace(s.serviceConfig.UploadPath) == "" {
		return "", fmt.Errorf("upload path is not configured")
	}
	if ext == "" || len(ext) > 8 {
		ext = ".png"
	}
	dir := filepath.Join(s.serviceConfig.UploadPath, constant.AIImageSubPath, userID)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	filename := fmt.Sprintf("%s-%d%s", generationID, index+1, ext)
	if err := os.WriteFile(filepath.Join(dir, filename), data, 0644); err != nil {
		return "", err
	}
	return fmt.Sprintf("/uploads/%s/%s/%s", constant.AIImageSubPath, userID, filename), nil
}

func (s *aiChatConfigService) runVideoGeneration(ctx context.Context, provider *entity.AIVideoProvider, model *entity.AIVideoModel, generationID, userID string, req *schema.AIVideoGenerateReq) {
	log.Infof("ai video generation start generation_id=%s user_id=%s provider=%s provider_model=%s", generationID, userID, provider.Name, model.ProviderModelID)
	if err := s.repo.UpdateVideoGeneration(ctx, generationID, &entity.AIVideoGeneration{
		Status:   "in_progress",
		Progress: 1,
	}, "status", "progress"); err != nil {
		log.Errorf("ai video generation start status update failed generation_id=%s user_id=%s error=%v", generationID, userID, err)
	}
	upstreamID, err := s.createUpstreamVideo(ctx, provider, model, generationID, req)
	if err != nil {
		log.Errorf("ai video upstream create failed generation_id=%s user_id=%s provider=%s provider_model=%s error=%v", generationID, userID, provider.Name, model.ProviderModelID, err)
		if updateErr := s.repo.UpdateVideoGeneration(ctx, generationID, &entity.AIVideoGeneration{
			Status: "failed",
			Error:  err.Error(),
		}, "status", "error"); updateErr != nil {
			log.Errorf("ai video generation failed status update failed generation_id=%s user_id=%s error=%v", generationID, userID, updateErr)
		}
		return
	}
	log.Infof("ai video upstream created generation_id=%s user_id=%s upstream_id=%s", generationID, userID, upstreamID)
	if err := s.repo.UpdateVideoGeneration(ctx, generationID, &entity.AIVideoGeneration{
		UpstreamID: upstreamID,
	}, "upstream_id"); err != nil {
		log.Errorf("ai video upstream id update failed generation_id=%s user_id=%s upstream_id=%s error=%v", generationID, userID, upstreamID, err)
	}

	videoURL, err := s.waitAndSaveUpstreamVideo(ctx, provider, generationID, upstreamID, userID)
	if err != nil {
		log.Errorf("ai video generation failed generation_id=%s user_id=%s upstream_id=%s error=%v", generationID, userID, upstreamID, err)
		if updateErr := s.repo.UpdateVideoGeneration(ctx, generationID, &entity.AIVideoGeneration{
			Status: "failed",
			Error:  err.Error(),
		}, "status", "error"); updateErr != nil {
			log.Errorf("ai video generation failed status update failed generation_id=%s user_id=%s error=%v", generationID, userID, updateErr)
		}
		return
	}
	if err := s.repo.UpdateVideoGeneration(ctx, generationID, &entity.AIVideoGeneration{
		VideoURL: videoURL,
		Status:   "completed",
		Progress: 100,
		Error:    "",
	}, "video_url", "status", "progress", "error"); err != nil {
		log.Errorf("ai video generation completed status update failed generation_id=%s user_id=%s video_url=%s error=%v", generationID, userID, videoURL, err)
		return
	}
	log.Infof("ai video generation completed generation_id=%s user_id=%s upstream_id=%s video_url=%s", generationID, userID, upstreamID, videoURL)
}

func (s *aiChatConfigService) createUpstreamVideo(ctx context.Context, provider *entity.AIVideoProvider, model *entity.AIVideoModel, generationID string, req *schema.AIVideoGenerateReq) (string, error) {
	baseURL := strings.TrimRight(provider.BaseURL, "/")
	log.Infof("ai video upstream request generation_id=%s endpoint=%s model=%s size=%s seconds=%d quality=%s preset=%s references=%d", generationID, "/videos", model.ProviderModelID, req.Size, req.Seconds, req.Quality, req.Preset, len(req.ReferenceImages))
	request := resty.New().
		SetRetryCount(1).
		SetHeader("Authorization", fmt.Sprintf("Bearer %s", provider.APIKey)).
		R().
		SetContext(ctx).
		SetFormData(map[string]string{
			"model":           model.ProviderModelID,
			"prompt":          req.Prompt,
			"seconds":         fmt.Sprint(req.Seconds),
			"size":            req.Size,
			"resolution_name": req.Quality,
			"preset":          req.Preset,
		})
	preparedImages, err := prepareReferenceImages(ctx, req.ReferenceImages)
	if err != nil {
		log.Errorf("ai video reference prepare failed generation_id=%s error=%v", generationID, err)
		return "", err
	}
	if len(preparedImages) > 0 {
		log.Infof("ai video prepared references generation_id=%s %s", generationID, summarizeReferenceImages(preparedImages))
	}
	for index, image := range preparedImages {
		request.SetFileReader("input_reference[]", fmt.Sprintf("reference-%d%s", index+1, image.Ext), bytes.NewReader(image.Data))
	}
	resp, err := request.Post(baseURL + "/videos")
	if err != nil {
		log.Errorf("ai video upstream request failed generation_id=%s endpoint=%s error=%v", generationID, "/videos", err)
		return "", err
	}
	if !resp.IsSuccess() {
		log.Errorf("ai video upstream non-success generation_id=%s endpoint=%s status=%d body=%s", generationID, "/videos", resp.StatusCode(), responseSnippet(resp.Body()))
		return "", fmt.Errorf("video create status %d: %s", resp.StatusCode(), resp.String())
	}
	var parsed struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(resp.Body(), &parsed); err != nil {
		log.Errorf("ai video upstream response parse failed generation_id=%s bytes=%d body=%s error=%v", generationID, len(resp.Body()), responseSnippet(resp.Body()), err)
		return "", err
	}
	if strings.TrimSpace(parsed.ID) == "" {
		log.Errorf("ai video upstream response missing id generation_id=%s bytes=%d body=%s", generationID, len(resp.Body()), responseSnippet(resp.Body()))
		return "", fmt.Errorf("video create response missing id")
	}
	log.Infof("ai video upstream success generation_id=%s endpoint=%s status=%d bytes=%d upstream_id=%s", generationID, "/videos", resp.StatusCode(), len(resp.Body()), parsed.ID)
	return parsed.ID, nil
}

func (s *aiChatConfigService) waitAndSaveUpstreamVideo(ctx context.Context, provider *entity.AIVideoProvider, generationID, upstreamID, userID string) (string, error) {
	baseURL := strings.TrimRight(provider.BaseURL, "/")
	client := resty.New().SetRetryCount(1)
	deadline := time.Now().Add(12 * time.Minute)
	lastStatus := ""
	lastProgress := -1
	for time.Now().Before(deadline) {
		resp, err := client.R().
			SetContext(ctx).
			SetHeader("Authorization", fmt.Sprintf("Bearer %s", provider.APIKey)).
			Get(baseURL + "/videos/" + url.PathEscape(upstreamID))
		if err != nil {
			log.Errorf("ai video upstream status request failed generation_id=%s upstream_id=%s error=%v", generationID, upstreamID, err)
			return "", err
		}
		if !resp.IsSuccess() {
			log.Errorf("ai video upstream status non-success generation_id=%s upstream_id=%s status=%d body=%s", generationID, upstreamID, resp.StatusCode(), responseSnippet(resp.Body()))
			return "", fmt.Errorf("video status %d: %s", resp.StatusCode(), resp.String())
		}
		var status struct {
			Status   string `json:"status"`
			Progress int    `json:"progress"`
			Error    struct {
				Message string `json:"message"`
			} `json:"error"`
		}
		if err := json.Unmarshal(resp.Body(), &status); err != nil {
			log.Errorf("ai video upstream status parse failed generation_id=%s upstream_id=%s bytes=%d body=%s error=%v", generationID, upstreamID, len(resp.Body()), responseSnippet(resp.Body()), err)
			return "", err
		}
		progress := max(1, min(99, status.Progress))
		normalizedStatus := normalizeVideoStatus(status.Status)
		recordStatus := normalizedStatus
		if status.Status == "completed" {
			recordStatus = "in_progress"
			progress = 99
		}
		if normalizedStatus != lastStatus || progress != lastProgress {
			log.Infof("ai video upstream status generation_id=%s upstream_id=%s status=%s progress=%d", generationID, upstreamID, normalizedStatus, progress)
			lastStatus = normalizedStatus
			lastProgress = progress
		}
		if err := s.repo.UpdateVideoGeneration(ctx, generationID, &entity.AIVideoGeneration{
			Status:   recordStatus,
			Progress: progress,
		}, "status", "progress"); err != nil {
			log.Errorf("ai video progress update failed generation_id=%s upstream_id=%s status=%s progress=%d error=%v", generationID, upstreamID, recordStatus, progress, err)
		}
		if status.Status == "completed" {
			raw, err := s.downloadUpstreamVideo(ctx, provider, upstreamID)
			if err != nil {
				log.Errorf("ai video download failed generation_id=%s upstream_id=%s error=%v", generationID, upstreamID, err)
				return "", err
			}
			return s.saveGeneratedVideo(userID, generationID, raw)
		}
		if status.Status == "failed" {
			if status.Error.Message != "" {
				return "", fmt.Errorf("%s", status.Error.Message)
			}
			return "", fmt.Errorf("video generation failed")
		}
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-time.After(3 * time.Second):
		}
	}
	log.Errorf("ai video generation timed out generation_id=%s upstream_id=%s", generationID, upstreamID)
	return "", fmt.Errorf("video generation timed out")
}

func (s *aiChatConfigService) downloadUpstreamVideo(ctx context.Context, provider *entity.AIVideoProvider, upstreamID string) ([]byte, error) {
	resp, err := resty.New().
		SetRetryCount(1).
		SetHeader("Authorization", fmt.Sprintf("Bearer %s", provider.APIKey)).
		R().
		SetContext(ctx).
		Get(strings.TrimRight(provider.BaseURL, "/") + "/videos/" + url.PathEscape(upstreamID) + "/content")
	if err != nil {
		log.Errorf("ai video content request failed upstream_id=%s error=%v", upstreamID, err)
		return nil, err
	}
	if !resp.IsSuccess() {
		log.Errorf("ai video content non-success upstream_id=%s status=%d body=%s", upstreamID, resp.StatusCode(), responseSnippet(resp.Body()))
		return nil, fmt.Errorf("video content status %d: %s", resp.StatusCode(), resp.String())
	}
	if len(resp.Body()) == 0 {
		log.Errorf("ai video content empty upstream_id=%s", upstreamID)
		return nil, fmt.Errorf("video content is empty")
	}
	log.Infof("ai video content downloaded upstream_id=%s bytes=%d", upstreamID, len(resp.Body()))
	return resp.Body(), nil
}

func (s *aiChatConfigService) saveGeneratedVideo(userID, generationID string, data []byte) (string, error) {
	if s.serviceConfig == nil || strings.TrimSpace(s.serviceConfig.UploadPath) == "" {
		return "", fmt.Errorf("upload path is not configured")
	}
	dir := filepath.Join(s.serviceConfig.UploadPath, constant.AIVideoSubPath, userID)
	if err := os.MkdirAll(dir, 0755); err != nil {
		log.Errorf("ai video save mkdir failed generation_id=%s user_id=%s dir=%s error=%v", generationID, userID, dir, err)
		return "", err
	}
	filename := fmt.Sprintf("%s.mp4", generationID)
	filePath := filepath.Join(dir, filename)
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		log.Errorf("ai video save write failed generation_id=%s user_id=%s path=%s bytes=%d error=%v", generationID, userID, filePath, len(data), err)
		return "", err
	}
	log.Infof("ai video saved generation_id=%s user_id=%s path=%s bytes=%d", generationID, userID, filePath, len(data))
	return fmt.Sprintf("/uploads/%s/%s/%s", constant.AIVideoSubPath, userID, filename), nil
}

func (s *aiChatConfigService) loadUserGeneratedImage(userID, imageURL string) (*preparedReferenceImage, error) {
	if s.serviceConfig == nil || strings.TrimSpace(s.serviceConfig.UploadPath) == "" {
		return nil, fmt.Errorf("upload path is not configured")
	}
	parsed, err := url.Parse(strings.TrimSpace(imageURL))
	if err != nil {
		return nil, err
	}
	cleanPath := path.Clean("/" + strings.TrimPrefix(parsed.Path, "/"))
	prefix := path.Join("/uploads", constant.AIImageSubPath, userID) + "/"
	if !strings.HasPrefix(cleanPath, prefix) {
		return nil, fmt.Errorf("image is outside current user upload path")
	}
	filename := path.Base(cleanPath)
	if filename == "." || filename == "/" || strings.Contains(filename, "..") {
		return nil, fmt.Errorf("image filename is invalid")
	}
	ext := strings.ToLower(filepath.Ext(filename))
	if ext == "" {
		ext = ".png"
	}
	filePath := filepath.Join(s.serviceConfig.UploadPath, constant.AIImageSubPath, userID, filename)
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	if len(data) > 20*1024*1024 {
		return nil, fmt.Errorf("image is too large")
	}
	mimeType := mime.TypeByExtension(ext)
	if mimeType == "" {
		mimeType = "image/png"
	}
	if !strings.HasPrefix(mimeType, "image/") {
		return nil, fmt.Errorf("file is not an image")
	}
	return &preparedReferenceImage{
		Data:    data,
		MIME:    mimeType,
		Ext:     ext,
		DataURL: fmt.Sprintf("data:%s;base64,%s", mimeType, base64.StdEncoding.EncodeToString(data)),
	}, nil
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

func currentDayRange() (time.Time, time.Time) {
	now := time.Now()
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	return start, start.AddDate(0, 0, 1)
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
		VideoDailyQuota: req.VideoDailyQuota,
		VideoQuota:      req.VideoQuota,
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

func buildImagePrompt(req *schema.AIImageGenerateReq) string {
	parts := []string{req.Prompt}
	if req.Style != "" {
		parts = append(parts, "Style: "+req.Style)
	}
	if req.AspectRatio != "" {
		parts = append(parts, "Aspect ratio: "+req.AspectRatio)
	}
	if req.NegativePrompt != "" {
		parts = append(parts, "Avoid: "+req.NegativePrompt)
	}
	return strings.Join(parts, "\n")
}

func normalizeOpenAIImageQuality(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", "auto":
		return "auto"
	case "low":
		return "low"
	case "medium", "standard", "标准":
		return "medium"
	case "high", "hd", "高清", "精修":
		return "high"
	default:
		return ""
	}
}

func shouldRequestImageResponseFormat(model string) bool {
	model = strings.ToLower(strings.TrimSpace(model))
	return strings.HasPrefix(model, "dall-e-")
}

func normalizeOpenAIImageSize(size, aspectRatio string) string {
	switch strings.ToLower(strings.TrimSpace(size)) {
	case "auto":
		return "auto"
	case "1024x1024":
		return "1024x1024"
	case "1536x1024":
		return "1536x1024"
	case "1024x1536":
		return "1024x1536"
	}
	switch strings.ToLower(strings.TrimSpace(aspectRatio)) {
	case "auto":
		return "auto"
	case "1:1":
		return "1024x1024"
	case "4:3", "16:9", "3:2":
		return "1536x1024"
	case "3:4", "9:16", "2:3":
		return "1024x1536"
	default:
		return "1024x1024"
	}
}

func imageAspectRatio(size string) string {
	switch strings.ToLower(strings.TrimSpace(size)) {
	case "auto":
		return "auto"
	case "1536x1024":
		return "3:2"
	case "1024x1536":
		return "2:3"
	default:
		return "1:1"
	}
}

func validateVideoSeconds(seconds int) error {
	switch seconds {
	case 6, 10, 12, 16, 20:
		return nil
	default:
		return errors.BadRequest("seconds must be one of 6, 10, 12, 16, 20")
	}
}

func validateVideoRequestOptions(size, quality, preset string) error {
	switch strings.TrimSpace(size) {
	case "720x1280", "1280x720", "1024x1024", "1024x1792", "1792x1024":
	default:
		return errors.BadRequest("size is not supported")
	}
	switch strings.TrimSpace(quality) {
	case "480p", "720p":
	default:
		return errors.BadRequest("quality must be 480p or 720p")
	}
	switch strings.TrimSpace(preset) {
	case "fun", "normal", "spicy", "custom":
	default:
		return errors.BadRequest("preset is not supported")
	}
	return nil
}

func videoAspectRatio(size string) string {
	switch strings.TrimSpace(size) {
	case "720x1280", "1024x1792":
		return "9:16"
	case "1280x720", "1792x1024":
		return "16:9"
	case "1024x1024":
		return "1:1"
	default:
		return "16:9"
	}
}

func normalizeVideoStatus(status string) string {
	switch strings.TrimSpace(status) {
	case "queued", "in_progress", "completed", "failed":
		return status
	default:
		return "in_progress"
	}
}

func (s *aiChatConfigService) checkVideoQuota(ctx context.Context, userID string, plan *entity.AISubscriptionPlan) error {
	if plan.VideoDailyQuota == -1 && plan.VideoQuota == -1 {
		return nil
	}
	dayStart, dayEnd := currentDayRange()
	monthStart, monthEnd := currentMonthRange()
	if plan.VideoDailyQuota != -1 {
		used, err := s.repo.CountUserVideoGenerations(ctx, userID, dayStart, dayEnd)
		if err != nil {
			return errors.InternalServer(reason.DatabaseError).WithError(err)
		}
		if used+1 > plan.VideoDailyQuota {
			return errors.BadRequest("daily video quota is insufficient")
		}
	}
	if plan.VideoQuota != -1 {
		used, err := s.repo.CountUserVideoGenerations(ctx, userID, monthStart, monthEnd)
		if err != nil {
			return errors.InternalServer(reason.DatabaseError).WithError(err)
		}
		if used+1 > plan.VideoQuota {
			return errors.BadRequest("monthly video quota is insufficient")
		}
	}
	return nil
}

func prepareReferenceImages(ctx context.Context, rawImages []string) ([]*preparedReferenceImage, error) {
	images := make([]*preparedReferenceImage, 0, len(rawImages))
	for _, rawImage := range rawImages {
		rawImage = strings.TrimSpace(rawImage)
		if rawImage == "" {
			continue
		}
		if strings.HasPrefix(rawImage, "http://") || strings.HasPrefix(rawImage, "https://") {
			data, ext, err := downloadImage(ctx, rawImage)
			if err != nil {
				return nil, err
			}
			mimeType := mime.TypeByExtension(ext)
			if mimeType == "" {
				mimeType = "image/png"
			}
			images = append(images, &preparedReferenceImage{
				Data:    data,
				MIME:    mimeType,
				Ext:     ext,
				DataURL: fmt.Sprintf("data:%s;base64,%s", mimeType, base64.StdEncoding.EncodeToString(data)),
			})
			continue
		}
		data, ext, err := decodeImageData(rawImage)
		if err != nil {
			return nil, err
		}
		mimeType := mime.TypeByExtension(ext)
		if mimeType == "" {
			mimeType = "image/png"
		}
		if !strings.HasPrefix(mimeType, "image/") {
			return nil, fmt.Errorf("reference image is not an image")
		}
		if len(data) > 20*1024*1024 {
			return nil, fmt.Errorf("reference image is too large")
		}
		images = append(images, &preparedReferenceImage{
			Data:    data,
			MIME:    mimeType,
			Ext:     ext,
			DataURL: fmt.Sprintf("data:%s;base64,%s", mimeType, base64.StdEncoding.EncodeToString(data)),
		})
	}
	return images, nil
}

func summarizeReferenceImages(images []*preparedReferenceImage) string {
	parts := make([]string, 0, len(images))
	for index, image := range images {
		parts = append(parts, fmt.Sprintf("reference_%d=%s/%dbytes", index+1, image.MIME, len(image.Data)))
	}
	return strings.Join(parts, " ")
}

func responseSnippet(body []byte) string {
	const limit = 1200
	text := strings.TrimSpace(string(body))
	if text == "" {
		return ""
	}
	text = base64DataPattern.ReplaceAllString(text, `"b64_json":"<omitted>"`)
	text = dataURLPattern.ReplaceAllString(text, `data:<omitted>`)
	if len(text) > limit {
		return text[:limit] + "...(truncated)"
	}
	return text
}

func decodeImageData(value string) ([]byte, string, error) {
	if commaIndex := strings.Index(value, ","); strings.HasPrefix(value, "data:") && commaIndex > 0 {
		meta := value[5:commaIndex]
		mimeType := strings.Split(meta, ";")[0]
		if mimeType == "" {
			mimeType = "image/png"
		}
		data, err := base64.StdEncoding.DecodeString(value[commaIndex+1:])
		if err != nil {
			return nil, "", err
		}
		ext := ".png"
		if exts, _ := mime.ExtensionsByType(mimeType); len(exts) > 0 {
			ext = exts[0]
		}
		return data, ext, nil
	}
	data, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		return nil, "", err
	}
	return data, ".png", nil
}

func downloadImage(ctx context.Context, rawURL string) ([]byte, string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, "", err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, "", fmt.Errorf("download status %d", resp.StatusCode)
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, 20*1024*1024))
	if err != nil {
		return nil, "", err
	}
	ext := ".png"
	if contentType := resp.Header.Get("Content-Type"); contentType != "" {
		if exts, _ := mime.ExtensionsByType(contentType); len(exts) > 0 {
			ext = exts[0]
		}
	}
	return data, ext, nil
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
