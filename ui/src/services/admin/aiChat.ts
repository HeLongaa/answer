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

import request from '@/utils/request';

const base = '/answer/admin/api/ai-chat';

export const getAiChatProviders = () => {
  return request.get<any[]>(`${base}/providers`);
};

export const createAiChatProvider = (params) => {
  return request.post(`${base}/providers`, params);
};

export const updateAiChatProvider = (id: number, params) => {
  return request.put(`${base}/providers/${id}`, params);
};

export const deleteAiChatProvider = (id: number) => {
  return request.delete(`${base}/providers/${id}`);
};

export const fetchAiChatProviderModels = (id: number) => {
  return request.post(`${base}/providers/${id}/fetch-models`);
};

export const testAiChatProviderModel = (
  id: number,
  params: { provider_model_id: string },
) => {
  return request.post(`${base}/providers/${id}/test-model`, params);
};

export const getAiChatModelMappings = () => {
  return request.get<any[]>(`${base}/model-mappings`);
};

export const createAiChatModelMapping = (params) => {
  return request.post(`${base}/model-mappings`, params);
};

export const updateAiChatModelMapping = (id: number, params) => {
  return request.put(`${base}/model-mappings/${id}`, params);
};

export const deleteAiChatModelMapping = (id: number) => {
  return request.delete(`${base}/model-mappings/${id}`);
};

export const getAiChatSubscriptionPlans = () => {
  return request.get<any[]>(`${base}/subscription-plans`);
};

export const createAiChatSubscriptionPlan = (params) => {
  return request.post(`${base}/subscription-plans`, params);
};

export const updateAiChatSubscriptionPlan = (id: number, params) => {
  return request.put(`${base}/subscription-plans/${id}`, params);
};

export const deleteAiChatSubscriptionPlan = (id: number) => {
  return request.delete(`${base}/subscription-plans/${id}`);
};

export const getAiChatRedeemCodes = () => {
  return request.get<any[]>(`${base}/redeem-codes`);
};

export const generateAiChatRedeemCodes = (params) => {
  return request.post<any[]>(`${base}/redeem-codes/generate`, params);
};

export const getAiChatConsumeRates = () => {
  return request.get<any[]>(`${base}/consume-rates`);
};

export const createAiChatConsumeRate = (params) => {
  return request.post(`${base}/consume-rates`, params);
};

export const updateAiChatConsumeRate = (id: number, params) => {
  return request.put(`${base}/consume-rates/${id}`, params);
};
