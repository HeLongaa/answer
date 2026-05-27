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

package entity

import "time"

type AIProvider struct {
	ID             int       `xorm:"not null pk autoincr INT(11) id"`
	Name           string    `xorm:"not null default '' VARCHAR(255) name"`
	BaseURL        string    `xorm:"not null default '' VARCHAR(500) base_url"`
	APIKey         string    `xorm:"not null TEXT api_key"`
	Enabled        bool      `xorm:"not null default true BOOL enabled"`
	SupportsStream bool      `xorm:"not null default true BOOL supports_stream"`
	Remark         string    `xorm:"not null TEXT remark"`
	CreatedAt      time.Time `xorm:"created not null default CURRENT_TIMESTAMP TIMESTAMP created_at"`
	UpdatedAt      time.Time `xorm:"updated not null default CURRENT_TIMESTAMP TIMESTAMP updated_at"`
}

func (AIProvider) TableName() string { return "ai_providers" }

type AIProviderModel struct {
	ID              int       `xorm:"not null pk autoincr INT(11) id"`
	ProviderID      int       `xorm:"not null index INT(11) provider_id"`
	ProviderModelID string    `xorm:"not null default '' VARCHAR(255) provider_model_id"`
	ModelName       string    `xorm:"not null default '' VARCHAR(255) model_name"`
	Enabled         bool      `xorm:"not null default true BOOL enabled"`
	FetchedAt       time.Time `xorm:"TIMESTAMP fetched_at"`
	CreatedAt       time.Time `xorm:"created not null default CURRENT_TIMESTAMP TIMESTAMP created_at"`
	UpdatedAt       time.Time `xorm:"updated not null default CURRENT_TIMESTAMP TIMESTAMP updated_at"`
}

func (AIProviderModel) TableName() string { return "ai_provider_models" }

type AIModelMapping struct {
	ID                     int       `xorm:"not null pk autoincr INT(11) id"`
	SiteModelID            string    `xorm:"not null unique VARCHAR(100) site_model_id"`
	DisplayName            string    `xorm:"not null default '' VARCHAR(255) display_name"`
	Description            string    `xorm:"not null TEXT description"`
	Enabled                bool      `xorm:"not null default true BOOL enabled"`
	SortOrder              int       `xorm:"not null default 0 INT(11) sort_order"`
	SupportsVision         bool      `xorm:"not null default false BOOL supports_vision"`
	FallbackEnabled        bool      `xorm:"not null default false BOOL fallback_enabled"`
	DefaultProviderModelID string    `xorm:"not null default '' VARCHAR(255) default_provider_model_id"`
	CreatedAt              time.Time `xorm:"created not null default CURRENT_TIMESTAMP TIMESTAMP created_at"`
	UpdatedAt              time.Time `xorm:"updated not null default CURRENT_TIMESTAMP TIMESTAMP updated_at"`
}

func (AIModelMapping) TableName() string { return "ai_model_mappings" }

type AIModelMappingItem struct {
	ID              int       `xorm:"not null pk autoincr INT(11) id"`
	MappingID       int       `xorm:"not null index INT(11) mapping_id"`
	ProviderID      int       `xorm:"not null index INT(11) provider_id"`
	ProviderModelID string    `xorm:"not null default '' VARCHAR(255) provider_model_id"`
	Priority        int       `xorm:"not null default 0 INT(11) priority"`
	Enabled         bool      `xorm:"not null default true BOOL enabled"`
	CreatedAt       time.Time `xorm:"created not null default CURRENT_TIMESTAMP TIMESTAMP created_at"`
	UpdatedAt       time.Time `xorm:"updated not null default CURRENT_TIMESTAMP TIMESTAMP updated_at"`
}

func (AIModelMappingItem) TableName() string { return "ai_model_mapping_items" }

type AISubscriptionPlan struct {
	ID              int       `xorm:"not null pk autoincr INT(11) id"`
	PlanID          string    `xorm:"not null unique VARCHAR(100) plan_id"`
	Name            string    `xorm:"not null default '' VARCHAR(255) name"`
	Enabled         bool      `xorm:"not null default true BOOL enabled"`
	MonthlyPrice    float64   `xorm:"not null default 0 DOUBLE monthly_price"`
	ChatPoints      int       `xorm:"not null default 0 INT(11) chat_points"`
	ImageQuota      int       `xorm:"not null default 0 INT(11) image_quota"`
	PurchaseURL     string    `xorm:"not null default '' VARCHAR(500) purchase_url"`
	TaskDescription string    `xorm:"not null TEXT task_description"`
	SortOrder       int       `xorm:"not null default 0 INT(11) sort_order"`
	CreatedAt       time.Time `xorm:"created not null default CURRENT_TIMESTAMP TIMESTAMP created_at"`
	UpdatedAt       time.Time `xorm:"updated not null default CURRENT_TIMESTAMP TIMESTAMP updated_at"`
}

func (AISubscriptionPlan) TableName() string { return "ai_subscription_plans" }

type AISubscriptionPlanModel struct {
	ID             int       `xorm:"not null pk autoincr INT(11) id"`
	PlanID         int       `xorm:"not null index INT(11) plan_id"`
	ModelMappingID int       `xorm:"not null index INT(11) model_mapping_id"`
	CreatedAt      time.Time `xorm:"created not null default CURRENT_TIMESTAMP TIMESTAMP created_at"`
}

func (AISubscriptionPlanModel) TableName() string { return "ai_subscription_plan_models" }

type AISubscriptionRedeemCode struct {
	ID             int       `xorm:"not null pk autoincr INT(11) id"`
	Code           string    `xorm:"not null unique VARCHAR(100) code"`
	PlanID         int       `xorm:"not null index INT(11) plan_id"`
	DurationMonths int       `xorm:"not null default 1 INT(11) duration_months"`
	Enabled        bool      `xorm:"not null default true BOOL enabled"`
	Used           bool      `xorm:"not null default false BOOL used"`
	UsedByUserID   string    `xorm:"not null default '' VARCHAR(100) used_by_user_id"`
	UsedAt         time.Time `xorm:"DATETIME used_at"`
	BatchNo        string    `xorm:"not null default '' VARCHAR(100) batch_no"`
	Remark         string    `xorm:"not null TEXT remark"`
	CreatedAt      time.Time `xorm:"created not null default CURRENT_TIMESTAMP TIMESTAMP created_at"`
	UpdatedAt      time.Time `xorm:"updated not null default CURRENT_TIMESTAMP TIMESTAMP updated_at"`
}

func (AISubscriptionRedeemCode) TableName() string { return "ai_subscription_redeem_codes" }

type AIModelConsumeRate struct {
	ID             int       `xorm:"not null pk autoincr INT(11) id"`
	ModelMappingID int       `xorm:"not null index INT(11) model_mapping_id"`
	ConsumeRate    float64   `xorm:"not null default 1 DOUBLE consume_rate"`
	Enabled        bool      `xorm:"not null default true BOOL enabled"`
	Remark         string    `xorm:"not null TEXT remark"`
	CreatedAt      time.Time `xorm:"created not null default CURRENT_TIMESTAMP TIMESTAMP created_at"`
	UpdatedAt      time.Time `xorm:"updated not null default CURRENT_TIMESTAMP TIMESTAMP updated_at"`
}

func (AIModelConsumeRate) TableName() string { return "ai_model_consume_rates" }

type AIChatUsageLog struct {
	ID               int       `xorm:"not null pk autoincr INT(11) id"`
	UserID           string    `xorm:"not null index VARCHAR(100) user_id"`
	ConversationID   string    `xorm:"not null index VARCHAR(255) conversation_id"`
	ChatCompletionID string    `xorm:"not null index VARCHAR(255) chat_completion_id"`
	SiteModelID      string    `xorm:"not null default '' VARCHAR(100) site_model_id"`
	ProviderID       int       `xorm:"not null default 0 INT(11) provider_id"`
	ProviderName     string    `xorm:"not null default '' VARCHAR(255) provider_name"`
	ProviderModelID  string    `xorm:"not null default '' VARCHAR(255) provider_model_id"`
	ConsumePoints    float64   `xorm:"not null default 0 DOUBLE consume_points"`
	CreatedAt        time.Time `xorm:"created not null default CURRENT_TIMESTAMP TIMESTAMP created_at"`
}

func (AIChatUsageLog) TableName() string { return "ai_chat_usage_logs" }
