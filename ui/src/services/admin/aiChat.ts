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

export const getAdminAiImageProviders = () => {
  return request.get<any[]>(`${base}/image-providers`);
};

export const createAdminAiImageProvider = (params) => {
  return request.post(`${base}/image-providers`, params);
};

export const updateAdminAiImageProvider = (id: number, params) => {
  return request.put(`${base}/image-providers/${id}`, params);
};

export const deleteAdminAiImageProvider = (id: number) => {
  return request.delete(`${base}/image-providers/${id}`);
};

export const getAdminAiImageModels = () => {
  return request.get<any[]>(`${base}/image-models`);
};

export const createAdminAiImageModel = (params) => {
  return request.post(`${base}/image-models`, params);
};

export const updateAdminAiImageModel = (id: number, params) => {
  return request.put(`${base}/image-models/${id}`, params);
};

export const deleteAdminAiImageModel = (id: number) => {
  return request.delete(`${base}/image-models/${id}`);
};

export const getAdminAiImageSetting = () => {
  return request.get<any>(`${base}/image-setting`);
};

export const updateAdminAiImageSetting = (params) => {
  return request.put(`${base}/image-setting`, params);
};

export const getAdminAiVideoProviders = () => {
  return request.get<any[]>(`${base}/video-providers`);
};

export const createAdminAiVideoProvider = (params) => {
  return request.post(`${base}/video-providers`, params);
};

export const updateAdminAiVideoProvider = (id: number, params) => {
  return request.put(`${base}/video-providers/${id}`, params);
};

export const deleteAdminAiVideoProvider = (id: number) => {
  return request.delete(`${base}/video-providers/${id}`);
};

export const getAdminAiVideoModels = () => {
  return request.get<any[]>(`${base}/video-models`);
};

export const createAdminAiVideoModel = (params) => {
  return request.post(`${base}/video-models`, params);
};

export const updateAdminAiVideoModel = (id: number, params) => {
  return request.put(`${base}/video-models/${id}`, params);
};

export const deleteAdminAiVideoModel = (id: number) => {
  return request.delete(`${base}/video-models/${id}`);
};

export const getAdminAiVideoSetting = () => {
  return request.get<any>(`${base}/video-setting`);
};

export const updateAdminAiVideoSetting = (params) => {
  return request.put(`${base}/video-setting`, params);
};
