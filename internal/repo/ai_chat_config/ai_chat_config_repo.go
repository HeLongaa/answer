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
	"errors"

	"github.com/apache/answer/internal/base/data"
	"github.com/apache/answer/internal/entity"
	"xorm.io/xorm"
)

var ErrRedeemCodeUsed = errors.New("redeem code already used")

type AIChatConfigRepo interface {
	ListProviders(ctx context.Context) ([]*entity.AIProvider, error)
	GetProvider(ctx context.Context, id int) (*entity.AIProvider, bool, error)
	CreateProvider(ctx context.Context, provider *entity.AIProvider) error
	UpdateProvider(ctx context.Context, provider *entity.AIProvider, cols ...string) error
	DeleteProvider(ctx context.Context, id int) error
	ListProviderModels(ctx context.Context, providerID int) ([]*entity.AIProviderModel, error)
	ReplaceProviderModels(ctx context.Context, providerID int, models []*entity.AIProviderModel) error

	ListModelMappings(ctx context.Context) ([]*entity.AIModelMapping, error)
	GetModelMapping(ctx context.Context, id int) (*entity.AIModelMapping, bool, error)
	GetModelMappingBySiteModelID(ctx context.Context, siteModelID string) (*entity.AIModelMapping, bool, error)
	CreateModelMapping(ctx context.Context, mapping *entity.AIModelMapping, items []*entity.AIModelMappingItem) error
	UpdateModelMapping(ctx context.Context, mapping *entity.AIModelMapping, items []*entity.AIModelMappingItem) error
	DeleteModelMapping(ctx context.Context, id int) error
	ListModelMappingItems(ctx context.Context, mappingID int) ([]*entity.AIModelMappingItem, error)

	ListSubscriptionPlans(ctx context.Context) ([]*entity.AISubscriptionPlan, error)
	GetSubscriptionPlan(ctx context.Context, id int) (*entity.AISubscriptionPlan, bool, error)
	GetSubscriptionPlanByPlanID(ctx context.Context, planID string) (*entity.AISubscriptionPlan, bool, error)
	CountCustomSubscriptionPlans(ctx context.Context, excludeID int) (int64, error)
	CreateSubscriptionPlan(ctx context.Context, plan *entity.AISubscriptionPlan, modelIDs []int) error
	UpdateSubscriptionPlan(ctx context.Context, plan *entity.AISubscriptionPlan, modelIDs []int) error
	DeleteSubscriptionPlan(ctx context.Context, id int) error
	ListSubscriptionPlanModels(ctx context.Context, planID int) ([]*entity.AISubscriptionPlanModel, error)
	EnsureFreePlan(ctx context.Context) error
	ListRedeemCodes(ctx context.Context) ([]*entity.AISubscriptionRedeemCode, error)
	GetRedeemCodeByCode(ctx context.Context, code string) (*entity.AISubscriptionRedeemCode, bool, error)
	CreateRedeemCodes(ctx context.Context, codes []*entity.AISubscriptionRedeemCode) error
	UseRedeemCode(ctx context.Context, code *entity.AISubscriptionRedeemCode, user *entity.User) error

	ListConsumeRates(ctx context.Context) ([]*entity.AIModelConsumeRate, error)
	GetConsumeRate(ctx context.Context, id int) (*entity.AIModelConsumeRate, bool, error)
	GetConsumeRateByModelMappingID(ctx context.Context, modelMappingID int) (*entity.AIModelConsumeRate, bool, error)
	SaveConsumeRate(ctx context.Context, rate *entity.AIModelConsumeRate) error
	CreateUsageLog(ctx context.Context, log *entity.AIChatUsageLog) error
	SumUserChatUsage(ctx context.Context, userID string, startAt, endAt any) (float64, error)

	EnsureImageTables(ctx context.Context) error
	ListImageProviders(ctx context.Context) ([]*entity.AIImageProvider, error)
	GetImageProvider(ctx context.Context, id int) (*entity.AIImageProvider, bool, error)
	CreateImageProvider(ctx context.Context, provider *entity.AIImageProvider) error
	UpdateImageProvider(ctx context.Context, provider *entity.AIImageProvider, cols ...string) error
	DeleteImageProvider(ctx context.Context, id int) error
	ListImageModels(ctx context.Context, onlyEnabled bool) ([]*entity.AIImageModel, error)
	GetImageModel(ctx context.Context, id int) (*entity.AIImageModel, bool, error)
	GetImageModelBySiteModelID(ctx context.Context, siteModelID string) (*entity.AIImageModel, bool, error)
	SaveImageModel(ctx context.Context, model *entity.AIImageModel) error
	DeleteImageModel(ctx context.Context, id int) error
	GetImageSetting(ctx context.Context) (*entity.AIImageSetting, bool, error)
	SaveImageSetting(ctx context.Context, setting *entity.AIImageSetting) error
	CreateImageGeneration(ctx context.Context, generation *entity.AIImageGeneration) error
	UpdateImageGeneration(ctx context.Context, generationID string, generation *entity.AIImageGeneration, cols ...string) error
	ListUserImageGenerations(ctx context.Context, userID string, limit int) ([]*entity.AIImageGeneration, error)
	CountUserImageGenerations(ctx context.Context, userID string, startAt, endAt any) (int, error)
}

type aiChatConfigRepo struct {
	data *data.Data
}

func NewAIChatConfigRepo(data *data.Data) AIChatConfigRepo {
	return &aiChatConfigRepo{data: data}
}

func (r *aiChatConfigRepo) ListProviders(ctx context.Context) ([]*entity.AIProvider, error) {
	list := make([]*entity.AIProvider, 0)
	return list, r.data.DB.Context(ctx).Asc("id").Find(&list)
}

func (r *aiChatConfigRepo) GetProvider(ctx context.Context, id int) (*entity.AIProvider, bool, error) {
	provider := &entity.AIProvider{}
	exist, err := r.data.DB.Context(ctx).ID(id).Get(provider)
	return provider, exist, err
}

func (r *aiChatConfigRepo) CreateProvider(ctx context.Context, provider *entity.AIProvider) error {
	_, err := r.data.DB.Context(ctx).Insert(provider)
	return err
}

func (r *aiChatConfigRepo) UpdateProvider(ctx context.Context, provider *entity.AIProvider, cols ...string) error {
	session := r.data.DB.Context(ctx).ID(provider.ID)
	if len(cols) > 0 {
		session = session.Cols(cols...)
	}
	_, err := session.Update(provider)
	return err
}

func (r *aiChatConfigRepo) DeleteProvider(ctx context.Context, id int) error {
	_, err := r.data.DB.Transaction(func(session *xorm.Session) (any, error) {
		if _, err := session.Context(ctx).Where("provider_id = ?", id).Delete(new(entity.AIProviderModel)); err != nil {
			return nil, err
		}
		if _, err := session.Context(ctx).Where("provider_id = ?", id).Delete(new(entity.AIModelMappingItem)); err != nil {
			return nil, err
		}
		_, err := session.Context(ctx).ID(id).Delete(new(entity.AIProvider))
		return nil, err
	})
	return err
}

func (r *aiChatConfigRepo) ListProviderModels(ctx context.Context, providerID int) ([]*entity.AIProviderModel, error) {
	list := make([]*entity.AIProviderModel, 0)
	return list, r.data.DB.Context(ctx).Where("provider_id = ?", providerID).Asc("provider_model_id").Find(&list)
}

func (r *aiChatConfigRepo) ReplaceProviderModels(ctx context.Context, providerID int, models []*entity.AIProviderModel) error {
	_, err := r.data.DB.Transaction(func(session *xorm.Session) (any, error) {
		if _, err := session.Context(ctx).Where("provider_id = ?", providerID).Delete(new(entity.AIProviderModel)); err != nil {
			return nil, err
		}
		if len(models) > 0 {
			if _, err := session.Context(ctx).Insert(models); err != nil {
				return nil, err
			}
		}
		return nil, nil
	})
	return err
}

func (r *aiChatConfigRepo) ListModelMappings(ctx context.Context) ([]*entity.AIModelMapping, error) {
	list := make([]*entity.AIModelMapping, 0)
	return list, r.data.DB.Context(ctx).Asc("sort_order", "id").Find(&list)
}

func (r *aiChatConfigRepo) GetModelMapping(ctx context.Context, id int) (*entity.AIModelMapping, bool, error) {
	mapping := &entity.AIModelMapping{}
	exist, err := r.data.DB.Context(ctx).ID(id).Get(mapping)
	return mapping, exist, err
}

func (r *aiChatConfigRepo) GetModelMappingBySiteModelID(ctx context.Context, siteModelID string) (*entity.AIModelMapping, bool, error) {
	mapping := &entity.AIModelMapping{}
	exist, err := r.data.DB.Context(ctx).Where("site_model_id = ?", siteModelID).Get(mapping)
	return mapping, exist, err
}

func (r *aiChatConfigRepo) CreateModelMapping(ctx context.Context, mapping *entity.AIModelMapping, items []*entity.AIModelMappingItem) error {
	_, err := r.data.DB.Transaction(func(session *xorm.Session) (any, error) {
		if _, err := session.Context(ctx).Insert(mapping); err != nil {
			return nil, err
		}
		for _, item := range items {
			item.MappingID = mapping.ID
		}
		if len(items) > 0 {
			if _, err := session.Context(ctx).Insert(items); err != nil {
				return nil, err
			}
		}
		return nil, nil
	})
	return err
}

func (r *aiChatConfigRepo) UpdateModelMapping(ctx context.Context, mapping *entity.AIModelMapping, items []*entity.AIModelMappingItem) error {
	_, err := r.data.DB.Transaction(func(session *xorm.Session) (any, error) {
		if _, err := session.Context(ctx).ID(mapping.ID).AllCols().Update(mapping); err != nil {
			return nil, err
		}
		if _, err := session.Context(ctx).Where("mapping_id = ?", mapping.ID).Delete(new(entity.AIModelMappingItem)); err != nil {
			return nil, err
		}
		for _, item := range items {
			item.MappingID = mapping.ID
		}
		if len(items) > 0 {
			if _, err := session.Context(ctx).Insert(items); err != nil {
				return nil, err
			}
		}
		return nil, nil
	})
	return err
}

func (r *aiChatConfigRepo) DeleteModelMapping(ctx context.Context, id int) error {
	_, err := r.data.DB.Transaction(func(session *xorm.Session) (any, error) {
		if _, err := session.Context(ctx).Where("mapping_id = ?", id).Delete(new(entity.AIModelMappingItem)); err != nil {
			return nil, err
		}
		if _, err := session.Context(ctx).Where("model_mapping_id = ?", id).Delete(new(entity.AIModelConsumeRate)); err != nil {
			return nil, err
		}
		if _, err := session.Context(ctx).Where("model_mapping_id = ?", id).Delete(new(entity.AISubscriptionPlanModel)); err != nil {
			return nil, err
		}
		_, err := session.Context(ctx).ID(id).Delete(new(entity.AIModelMapping))
		return nil, err
	})
	return err
}

func (r *aiChatConfigRepo) ListModelMappingItems(ctx context.Context, mappingID int) ([]*entity.AIModelMappingItem, error) {
	list := make([]*entity.AIModelMappingItem, 0)
	return list, r.data.DB.Context(ctx).Where("mapping_id = ?", mappingID).Asc("priority", "id").Find(&list)
}

func (r *aiChatConfigRepo) ListSubscriptionPlans(ctx context.Context) ([]*entity.AISubscriptionPlan, error) {
	list := make([]*entity.AISubscriptionPlan, 0)
	return list, r.data.DB.Context(ctx).Asc("sort_order", "id").Find(&list)
}

func (r *aiChatConfigRepo) GetSubscriptionPlan(ctx context.Context, id int) (*entity.AISubscriptionPlan, bool, error) {
	plan := &entity.AISubscriptionPlan{}
	exist, err := r.data.DB.Context(ctx).ID(id).Get(plan)
	return plan, exist, err
}

func (r *aiChatConfigRepo) GetSubscriptionPlanByPlanID(ctx context.Context, planID string) (*entity.AISubscriptionPlan, bool, error) {
	plan := &entity.AISubscriptionPlan{}
	exist, err := r.data.DB.Context(ctx).Where("plan_id = ?", planID).Get(plan)
	return plan, exist, err
}

func (r *aiChatConfigRepo) CountCustomSubscriptionPlans(ctx context.Context, excludeID int) (int64, error) {
	session := r.data.DB.Context(ctx).Where("plan_id <> ?", "free")
	if excludeID > 0 {
		session = session.Where("id <> ?", excludeID)
	}
	return session.Count(new(entity.AISubscriptionPlan))
}

func (r *aiChatConfigRepo) CreateSubscriptionPlan(ctx context.Context, plan *entity.AISubscriptionPlan, modelIDs []int) error {
	_, err := r.data.DB.Transaction(func(session *xorm.Session) (any, error) {
		if _, err := session.Context(ctx).Insert(plan); err != nil {
			return nil, err
		}
		return nil, r.replacePlanModels(ctx, session, plan.ID, modelIDs)
	})
	return err
}

func (r *aiChatConfigRepo) UpdateSubscriptionPlan(ctx context.Context, plan *entity.AISubscriptionPlan, modelIDs []int) error {
	_, err := r.data.DB.Transaction(func(session *xorm.Session) (any, error) {
		if _, err := session.Context(ctx).ID(plan.ID).AllCols().Update(plan); err != nil {
			return nil, err
		}
		return nil, r.replacePlanModels(ctx, session, plan.ID, modelIDs)
	})
	return err
}

func (r *aiChatConfigRepo) replacePlanModels(ctx context.Context, session *xorm.Session, planID int, modelIDs []int) error {
	if _, err := session.Context(ctx).Where("plan_id = ?", planID).Delete(new(entity.AISubscriptionPlanModel)); err != nil {
		return err
	}
	items := make([]*entity.AISubscriptionPlanModel, 0, len(modelIDs))
	for _, modelID := range modelIDs {
		items = append(items, &entity.AISubscriptionPlanModel{PlanID: planID, ModelMappingID: modelID})
	}
	if len(items) == 0 {
		return nil
	}
	_, err := session.Context(ctx).Insert(items)
	return err
}

func (r *aiChatConfigRepo) DeleteSubscriptionPlan(ctx context.Context, id int) error {
	_, err := r.data.DB.Transaction(func(session *xorm.Session) (any, error) {
		if _, err := session.Context(ctx).Where("plan_id = ?", id).Delete(new(entity.AISubscriptionPlanModel)); err != nil {
			return nil, err
		}
		_, err := session.Context(ctx).ID(id).Delete(new(entity.AISubscriptionPlan))
		return nil, err
	})
	return err
}

func (r *aiChatConfigRepo) ListSubscriptionPlanModels(ctx context.Context, planID int) ([]*entity.AISubscriptionPlanModel, error) {
	list := make([]*entity.AISubscriptionPlanModel, 0)
	return list, r.data.DB.Context(ctx).Where("plan_id = ?", planID).Find(&list)
}

func (r *aiChatConfigRepo) EnsureFreePlan(ctx context.Context) error {
	exist, err := r.data.DB.Context(ctx).Where("plan_id = ?", "free").Exist(new(entity.AISubscriptionPlan))
	if err != nil || exist {
		return err
	}
	_, err = r.data.DB.Context(ctx).Insert(&entity.AISubscriptionPlan{
		PlanID:          "free",
		Name:            "FREE",
		Enabled:         true,
		TaskDescription: "Default free plan",
	})
	return err
}

func (r *aiChatConfigRepo) ListRedeemCodes(ctx context.Context) ([]*entity.AISubscriptionRedeemCode, error) {
	list := make([]*entity.AISubscriptionRedeemCode, 0)
	return list, r.data.DB.Context(ctx).Desc("id").Find(&list)
}

func (r *aiChatConfigRepo) GetRedeemCodeByCode(ctx context.Context, code string) (*entity.AISubscriptionRedeemCode, bool, error) {
	redeemCode := &entity.AISubscriptionRedeemCode{}
	exist, err := r.data.DB.Context(ctx).Where("code = ?", code).Get(redeemCode)
	return redeemCode, exist, err
}

func (r *aiChatConfigRepo) CreateRedeemCodes(ctx context.Context, codes []*entity.AISubscriptionRedeemCode) error {
	if len(codes) == 0 {
		return nil
	}
	_, err := r.data.DB.Context(ctx).Insert(codes)
	return err
}

func (r *aiChatConfigRepo) UseRedeemCode(ctx context.Context, code *entity.AISubscriptionRedeemCode, user *entity.User) error {
	_, err := r.data.DB.Transaction(func(session *xorm.Session) (any, error) {
		affected, err := session.Context(ctx).
			Where("id = ? AND used = ?", code.ID, false).
			Cols("used", "used_by_user_id", "used_at").
			Update(code)
		if err != nil {
			return nil, err
		}
		if affected == 0 {
			return nil, ErrRedeemCodeUsed
		}
		_, err = session.Context(ctx).ID(user.ID).
			Cols("subscription_level", "subscription_started_at", "subscription_expires_at").
			Update(user)
		return nil, err
	})
	return err
}

func (r *aiChatConfigRepo) ListConsumeRates(ctx context.Context) ([]*entity.AIModelConsumeRate, error) {
	list := make([]*entity.AIModelConsumeRate, 0)
	return list, r.data.DB.Context(ctx).Asc("model_mapping_id", "id").Find(&list)
}

func (r *aiChatConfigRepo) GetConsumeRate(ctx context.Context, id int) (*entity.AIModelConsumeRate, bool, error) {
	rate := &entity.AIModelConsumeRate{}
	exist, err := r.data.DB.Context(ctx).ID(id).Get(rate)
	return rate, exist, err
}

func (r *aiChatConfigRepo) GetConsumeRateByModelMappingID(ctx context.Context, modelMappingID int) (*entity.AIModelConsumeRate, bool, error) {
	rate := &entity.AIModelConsumeRate{}
	exist, err := r.data.DB.Context(ctx).Where("model_mapping_id = ?", modelMappingID).Get(rate)
	return rate, exist, err
}

func (r *aiChatConfigRepo) SaveConsumeRate(ctx context.Context, rate *entity.AIModelConsumeRate) error {
	if rate.ID > 0 {
		_, err := r.data.DB.Context(ctx).ID(rate.ID).AllCols().Update(rate)
		return err
	}
	_, err := r.data.DB.Context(ctx).Insert(rate)
	return err
}

func (r *aiChatConfigRepo) CreateUsageLog(ctx context.Context, log *entity.AIChatUsageLog) error {
	_, err := r.data.DB.Context(ctx).Insert(log)
	return err
}

func (r *aiChatConfigRepo) SumUserChatUsage(ctx context.Context, userID string, startAt, endAt any) (float64, error) {
	var total struct {
		ConsumePoints float64 `xorm:"consume_points"`
	}
	ok, err := r.data.DB.Context(ctx).
		Table(new(entity.AIChatUsageLog)).
		Select("COALESCE(SUM(consume_points), 0) AS consume_points").
		Where("user_id = ?", userID).
		And("created_at >= ?", startAt).
		And("created_at < ?", endAt).
		Get(&total)
	if err != nil || !ok {
		return 0, err
	}
	return total.ConsumePoints, nil
}

func (r *aiChatConfigRepo) EnsureImageTables(ctx context.Context) error {
	if err := r.data.DB.Context(ctx).Sync(
		new(entity.AIImageProvider),
		new(entity.AIImageModel),
		new(entity.AIImageSetting),
		new(entity.AIImageGeneration),
	); err != nil {
		return err
	}
	exist, err := r.data.DB.Context(ctx).ID(1).Exist(new(entity.AIImageSetting))
	if err != nil || exist {
		return err
	}
	_, err = r.data.DB.Context(ctx).Insert(&entity.AIImageSetting{
		ID:            1,
		RetentionDays: 30,
	})
	return err
}

func (r *aiChatConfigRepo) ListImageProviders(ctx context.Context) ([]*entity.AIImageProvider, error) {
	list := make([]*entity.AIImageProvider, 0)
	return list, r.data.DB.Context(ctx).Asc("id").Find(&list)
}

func (r *aiChatConfigRepo) GetImageProvider(ctx context.Context, id int) (*entity.AIImageProvider, bool, error) {
	provider := &entity.AIImageProvider{}
	exist, err := r.data.DB.Context(ctx).ID(id).Get(provider)
	return provider, exist, err
}

func (r *aiChatConfigRepo) CreateImageProvider(ctx context.Context, provider *entity.AIImageProvider) error {
	_, err := r.data.DB.Context(ctx).Insert(provider)
	return err
}

func (r *aiChatConfigRepo) UpdateImageProvider(ctx context.Context, provider *entity.AIImageProvider, cols ...string) error {
	session := r.data.DB.Context(ctx).ID(provider.ID)
	if len(cols) > 0 {
		session = session.Cols(cols...)
	}
	_, err := session.Update(provider)
	return err
}

func (r *aiChatConfigRepo) DeleteImageProvider(ctx context.Context, id int) error {
	_, err := r.data.DB.Transaction(func(session *xorm.Session) (any, error) {
		if _, err := session.Context(ctx).Where("provider_id = ?", id).Delete(new(entity.AIImageModel)); err != nil {
			return nil, err
		}
		_, err := session.Context(ctx).ID(id).Delete(new(entity.AIImageProvider))
		return nil, err
	})
	return err
}

func (r *aiChatConfigRepo) ListImageModels(ctx context.Context, onlyEnabled bool) ([]*entity.AIImageModel, error) {
	list := make([]*entity.AIImageModel, 0)
	session := r.data.DB.Context(ctx).Asc("sort_order", "id")
	if onlyEnabled {
		session = session.Where("enabled = ?", true)
	}
	return list, session.Find(&list)
}

func (r *aiChatConfigRepo) GetImageModel(ctx context.Context, id int) (*entity.AIImageModel, bool, error) {
	model := &entity.AIImageModel{}
	exist, err := r.data.DB.Context(ctx).ID(id).Get(model)
	return model, exist, err
}

func (r *aiChatConfigRepo) GetImageModelBySiteModelID(ctx context.Context, siteModelID string) (*entity.AIImageModel, bool, error) {
	model := &entity.AIImageModel{}
	exist, err := r.data.DB.Context(ctx).Where("site_model_id = ?", siteModelID).Get(model)
	return model, exist, err
}

func (r *aiChatConfigRepo) SaveImageModel(ctx context.Context, model *entity.AIImageModel) error {
	if model.ID > 0 {
		_, err := r.data.DB.Context(ctx).ID(model.ID).AllCols().Update(model)
		return err
	}
	_, err := r.data.DB.Context(ctx).Insert(model)
	return err
}

func (r *aiChatConfigRepo) DeleteImageModel(ctx context.Context, id int) error {
	_, err := r.data.DB.Context(ctx).ID(id).Delete(new(entity.AIImageModel))
	return err
}

func (r *aiChatConfigRepo) GetImageSetting(ctx context.Context) (*entity.AIImageSetting, bool, error) {
	setting := &entity.AIImageSetting{}
	exist, err := r.data.DB.Context(ctx).ID(1).Get(setting)
	return setting, exist, err
}

func (r *aiChatConfigRepo) SaveImageSetting(ctx context.Context, setting *entity.AIImageSetting) error {
	setting.ID = 1
	exist, err := r.data.DB.Context(ctx).ID(1).Exist(new(entity.AIImageSetting))
	if err != nil {
		return err
	}
	if exist {
		_, err = r.data.DB.Context(ctx).ID(1).Cols("retention_days").Update(setting)
		return err
	}
	_, err = r.data.DB.Context(ctx).Insert(setting)
	return err
}

func (r *aiChatConfigRepo) CreateImageGeneration(ctx context.Context, generation *entity.AIImageGeneration) error {
	_, err := r.data.DB.Context(ctx).Insert(generation)
	return err
}

func (r *aiChatConfigRepo) UpdateImageGeneration(ctx context.Context, generationID string, generation *entity.AIImageGeneration, cols ...string) error {
	_, err := r.data.DB.Context(ctx).
		Where("generation_id = ?", generationID).
		Cols(cols...).
		Update(generation)
	return err
}

func (r *aiChatConfigRepo) ListUserImageGenerations(ctx context.Context, userID string, limit int) ([]*entity.AIImageGeneration, error) {
	if limit <= 0 || limit > 100 {
		limit = 30
	}
	list := make([]*entity.AIImageGeneration, 0)
	return list, r.data.DB.Context(ctx).
		Where("user_id = ?", userID).
		Desc("id").
		Limit(limit).
		Find(&list)
}

func (r *aiChatConfigRepo) CountUserImageGenerations(ctx context.Context, userID string, startAt, endAt any) (int, error) {
	var total struct {
		Count int `xorm:"count"`
	}
	ok, err := r.data.DB.Context(ctx).
		Table(new(entity.AIImageGeneration)).
		Select("COALESCE(SUM(count), 0) AS count").
		Where("user_id = ?", userID).
		And("status = ?", "completed").
		And("created_at >= ?", startAt).
		And("created_at < ?", endAt).
		Get(&total)
	if err != nil || !ok {
		return 0, err
	}
	return total.Count, nil
}
