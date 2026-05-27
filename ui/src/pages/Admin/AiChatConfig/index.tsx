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

import { FormEvent, useEffect, useMemo, useState } from 'react';
import {
  Alert,
  Badge,
  Button,
  Card,
  Col,
  Form,
  Modal,
  Row,
  Spinner,
  Tab,
  Table,
  Tabs,
} from 'react-bootstrap';
import { useSearchParams } from 'react-router-dom';

import { useToast } from '@/hooks';
import {
  createAiChatConsumeRate,
  createAiChatModelMapping,
  createAiChatProvider,
  createAiChatSubscriptionPlan,
  deleteAiChatModelMapping,
  deleteAiChatProvider,
  deleteAiChatSubscriptionPlan,
  fetchAiChatProviderModels,
  generateAiChatRedeemCodes,
  getAiChatConsumeRates,
  getAiChatModelMappings,
  getAiChatProviders,
  getAiChatRedeemCodes,
  getAiChatSubscriptionPlans,
  testAiChatProviderModel,
  updateAiChatConsumeRate,
  updateAiChatModelMapping,
  updateAiChatProvider,
  updateAiChatSubscriptionPlan,
} from '@/services';

import './index.scss';

const providerInit = {
  id: 0,
  name: '',
  base_url: 'https://api.openai.com/v1',
  api_key: '',
  enabled: true,
  supports_stream: true,
  remark: '',
};

const newMappingItem = (priority = 1) => ({
  id: 0,
  client_id: `${Date.now()}-${Math.random().toString(36).slice(2)}`,
  provider_id: 0,
  provider_model_id: '',
  priority,
  enabled: true,
});

const mappingInit = {
  id: 0,
  site_model_id: '',
  display_name: '',
  description: '',
  enabled: true,
  sort_order: 0,
  supports_vision: false,
  fallback_enabled: true,
  default_provider_model_id: '',
  items: [newMappingItem()],
};

const planInit = {
  id: 0,
  plan_id: '',
  name: '',
  enabled: true,
  monthly_price: 0,
  chat_points: 0,
  image_quota: 0,
  purchase_url: '',
  model_mapping_ids: [] as number[],
  task_description: '',
  sort_order: 0,
};

const rateInit = {
  id: 0,
  model_mapping_id: 0,
  consume_rate: 1,
  enabled: true,
  remark: '',
};

const redeemInit = {
  plan_id: 0,
  count: 10,
  duration_months: 1,
  prefix: '',
  remark: '',
};

const tabKeys = ['providers', 'mappings', 'plans', 'redeem-codes', 'rates'];

const formatQuota = (value: number) => (value === -1 ? '无限制' : value);

const formatDateTime = (value?: number) => {
  if (!value) {
    return '-';
  }
  return new Date(value * 1000).toLocaleString();
};

const AiChatConfig = () => {
  const toast = useToast();
  const [searchParams, setSearchParams] = useSearchParams();
  const tabFromURL = searchParams.get('tab') || 'providers';
  const [providers, setProviders] = useState<any[]>([]);
  const [mappings, setMappings] = useState<any[]>([]);
  const [plans, setPlans] = useState<any[]>([]);
  const [redeemCodes, setRedeemCodes] = useState<any[]>([]);
  const [rates, setRates] = useState<any[]>([]);
  const [providerForm, setProviderForm] = useState(providerInit);
  const [mappingForm, setMappingForm] = useState(mappingInit);
  const [planForm, setPlanForm] = useState(planInit);
  const [rateForm, setRateForm] = useState(rateInit);
  const [redeemForm, setRedeemForm] = useState(redeemInit);
  const [generatedCodes, setGeneratedCodes] = useState<any[]>([]);
  const [activeTab, setActiveTab] = useState(
    tabKeys.includes(tabFromURL) ? tabFromURL : 'providers',
  );
  const [testingProvider, setTestingProvider] = useState<any>(null);
  const [testingModelID, setTestingModelID] = useState('');
  const [testingResult, setTestingResult] = useState<any>(null);
  const [testing, setTesting] = useState(false);
  const [initialLoading, setInitialLoading] = useState(false);
  const [loading, setLoading] = useState(false);
  const [fetchingProviderID, setFetchingProviderID] = useState(0);
  const [error, setError] = useState('');

  const upstreamOptions = useMemo(
    () =>
      providers.flatMap((provider) =>
        (provider.models || []).map((model) => ({
          provider_id: provider.id,
          provider_name: provider.name,
          provider_model_id: model.provider_model_id,
          label: `${provider.name} / ${model.provider_model_id}`,
        })),
      ),
    [providers],
  );

  const loadAll = async (showFullLoading = false) => {
    if (showFullLoading) {
      setInitialLoading(true);
    } else {
      setLoading(true);
    }
    setError('');
    try {
      const [providerData, mappingData, planData, redeemCodeData, rateData] =
        await Promise.all([
          getAiChatProviders(),
          getAiChatModelMappings(),
          getAiChatSubscriptionPlans(),
          getAiChatRedeemCodes(),
          getAiChatConsumeRates(),
        ]);
      setProviders(providerData || []);
      setMappings(mappingData || []);
      setPlans(planData || []);
      setRedeemCodes(redeemCodeData || []);
      setRates(rateData || []);
    } catch (err: any) {
      setError(err?.msg || '加载 AI-CHAT 配置失败');
    } finally {
      setInitialLoading(false);
      setLoading(false);
    }
  };

  useEffect(() => {
    loadAll(true);
  }, []);

  useEffect(() => {
    if (tabKeys.includes(tabFromURL) && tabFromURL !== activeTab) {
      setActiveTab(tabFromURL);
    }
  }, [activeTab, tabFromURL]);

  const showSuccess = (msg: string) => {
    toast.onShow({ msg, variant: 'success' });
  };

  const submitProvider = async (evt: FormEvent) => {
    evt.preventDefault();
    setError('');
    try {
      if (providerForm.id) {
        await updateAiChatProvider(providerForm.id, providerForm);
      } else {
        await createAiChatProvider(providerForm);
      }
      setProviderForm(providerInit);
      showSuccess('Provider 已保存');
      await loadAll();
    } catch (err: any) {
      setError(err?.msg || 'Provider 保存失败');
    }
  };

  const submitMapping = async (evt: FormEvent) => {
    evt.preventDefault();
    setError('');
    try {
      const payload = {
        ...mappingForm,
        sort_order: Number(mappingForm.sort_order),
        items: mappingForm.items.map(({ client_id, ...item }) => ({
          ...item,
          provider_id: Number(item.provider_id),
          priority: Number(item.priority),
        })),
      };
      if (mappingForm.id) {
        await updateAiChatModelMapping(mappingForm.id, payload);
      } else {
        await createAiChatModelMapping(payload);
      }
      setMappingForm(mappingInit);
      showSuccess('模型映射已保存');
      await loadAll();
    } catch (err: any) {
      setError(err?.msg || '模型映射保存失败');
    }
  };

  const submitPlan = async (evt: FormEvent) => {
    evt.preventDefault();
    setError('');
    try {
      const payload = {
        ...planForm,
        monthly_price: Number(planForm.monthly_price),
        chat_points: Number(planForm.chat_points),
        image_quota: Number(planForm.image_quota),
        sort_order: Number(planForm.sort_order),
        model_mapping_ids: planForm.model_mapping_ids.map(Number),
      };
      if (planForm.id) {
        await updateAiChatSubscriptionPlan(planForm.id, payload);
      } else {
        await createAiChatSubscriptionPlan(payload);
      }
      setPlanForm(planInit);
      showSuccess('订阅等级已保存');
      await loadAll();
    } catch (err: any) {
      setError(err?.msg || '订阅等级保存失败');
    }
  };

  const submitRate = async (evt: FormEvent) => {
    evt.preventDefault();
    setError('');
    try {
      const payload = {
        ...rateForm,
        model_mapping_id: Number(rateForm.model_mapping_id),
        consume_rate: Number(rateForm.consume_rate),
      };
      if (rateForm.id) {
        await updateAiChatConsumeRate(rateForm.id, payload);
      } else {
        await createAiChatConsumeRate(payload);
      }
      setRateForm(rateInit);
      showSuccess('消耗系数已保存');
      await loadAll();
    } catch (err: any) {
      setError(err?.msg || '消耗系数保存失败');
    }
  };

  const submitRedeemCodes = async (evt: FormEvent) => {
    evt.preventDefault();
    setError('');
    setGeneratedCodes([]);
    try {
      const payload = {
        ...redeemForm,
        plan_id: Number(redeemForm.plan_id),
        count: Number(redeemForm.count),
        duration_months: Number(redeemForm.duration_months),
      };
      const resp = await generateAiChatRedeemCodes(payload);
      setGeneratedCodes(resp || []);
      setRedeemForm({ ...redeemForm, count: 10 });
      showSuccess('兑换码已生成');
      await loadAll();
    } catch (err: any) {
      setError(err?.msg || '兑换码生成失败');
    }
  };

  const fetchModels = async (providerID: number) => {
    setFetchingProviderID(providerID);
    setError('');
    try {
      await fetchAiChatProviderModels(providerID);
      showSuccess('模型列表已更新');
      await loadAll();
    } catch (err: any) {
      setError(err?.msg || '获取模型列表失败');
    } finally {
      setFetchingProviderID(0);
    }
  };

  const openTestProvider = (provider) => {
    const firstModel = provider.models?.[0]?.provider_model_id || '';
    setTestingProvider(provider);
    setTestingModelID(firstModel);
    setTestingResult(null);
    setError('');
  };

  const closeTestProvider = () => {
    if (testing) {
      return;
    }
    setTestingProvider(null);
    setTestingModelID('');
    setTestingResult(null);
  };

  const testProviderModel = async () => {
    if (!testingProvider || !testingModelID) {
      return;
    }
    setTesting(true);
    setTestingResult(null);
    setError('');
    try {
      const resp = await testAiChatProviderModel(testingProvider.id, {
        provider_model_id: testingModelID,
      });
      setTestingResult(resp);
      showSuccess('模型测试成功');
    } catch (err: any) {
      setTestingResult({
        error: err?.msg || '模型测试失败',
      });
    } finally {
      setTesting(false);
    }
  };

  const updateMappingItem = (
    index: number,
    patch: Partial<(typeof mappingForm.items)[0]>,
  ) => {
    const items = [...mappingForm.items];
    items[index] = { ...items[index], ...patch };
    setMappingForm({ ...mappingForm, items });
  };

  const extraPlanCount = plans.filter((plan) => plan.plan_id !== 'free').length;
  const paidPlans = plans.filter((plan) => plan.plan_id !== 'free');

  if (initialLoading) {
    return <Spinner animation="border" />;
  }

  return (
    <div className="ai-chat-config-page">
      <h3 className="mb-4">
        AI-CHAT配置
        {loading ? (
          <Spinner animation="border" size="sm" className="ms-2" />
        ) : null}
      </h3>
      {error ? <Alert variant="danger">{error}</Alert> : null}
      <Tabs
        activeKey={activeTab}
        onSelect={(key) => {
          if (key) {
            setActiveTab(key);
            setSearchParams({ tab: key });
          }
        }}
        className="ai-chat-config-tabs mb-4">
        <Tab eventKey="providers" title="Provider 管理">
          <Card className="mb-4">
            <Card.Body>
              <Form onSubmit={submitProvider}>
                <Row>
                  <Col md={4}>
                    <Form.Group className="mb-3">
                      <Form.Label>Provider 名称</Form.Label>
                      <Form.Control
                        required
                        value={providerForm.name}
                        onChange={(e) =>
                          setProviderForm({
                            ...providerForm,
                            name: e.target.value,
                          })
                        }
                      />
                    </Form.Group>
                  </Col>
                  <Col md={5}>
                    <Form.Group className="mb-3">
                      <Form.Label>Base URL</Form.Label>
                      <Form.Control
                        required
                        value={providerForm.base_url}
                        onChange={(e) =>
                          setProviderForm({
                            ...providerForm,
                            base_url: e.target.value,
                          })
                        }
                      />
                    </Form.Group>
                  </Col>
                  <Col md={3}>
                    <Form.Group className="mb-3">
                      <Form.Label>API Key</Form.Label>
                      <Form.Control
                        type="password"
                        required={!providerForm.id}
                        placeholder={
                          providerForm.id ? '留空则保持原 API Key' : ''
                        }
                        value={providerForm.api_key}
                        onChange={(e) =>
                          setProviderForm({
                            ...providerForm,
                            api_key: e.target.value,
                          })
                        }
                      />
                    </Form.Group>
                  </Col>
                </Row>
                <Form.Group className="mb-3">
                  <Form.Label>备注</Form.Label>
                  <Form.Control
                    as="textarea"
                    rows={2}
                    value={providerForm.remark}
                    onChange={(e) =>
                      setProviderForm({
                        ...providerForm,
                        remark: e.target.value,
                      })
                    }
                  />
                </Form.Group>
                <Form.Check
                  className="mb-3"
                  type="switch"
                  label="启用"
                  checked={providerForm.enabled}
                  onChange={(e) =>
                    setProviderForm({
                      ...providerForm,
                      enabled: e.target.checked,
                    })
                  }
                />
                <Form.Check
                  className="mb-3"
                  type="switch"
                  label="支持流式输出"
                  checked={providerForm.supports_stream}
                  onChange={(e) =>
                    setProviderForm({
                      ...providerForm,
                      supports_stream: e.target.checked,
                    })
                  }
                />
                <Button type="submit" className="me-2">
                  保存 Provider
                </Button>
                {providerForm.id ? (
                  <Button
                    type="button"
                    variant="link"
                    onClick={() => setProviderForm(providerInit)}>
                    取消编辑
                  </Button>
                ) : null}
              </Form>
            </Card.Body>
          </Card>
          <Table responsive hover>
            <thead>
              <tr>
                <th>名称</th>
                <th>Base URL</th>
                <th>状态</th>
                <th>流式</th>
                <th>模型</th>
                <th className="ai-chat-config-action-col">操作</th>
              </tr>
            </thead>
            <tbody>
              {providers.map((provider) => (
                <tr key={provider.id}>
                  <td
                    className="ai-chat-config-text-cell"
                    title={provider.name}>
                    {provider.name}
                  </td>
                  <td
                    className="ai-chat-config-text-cell"
                    title={provider.base_url}>
                    {provider.base_url}
                  </td>
                  <td>
                    <Badge bg={provider.enabled ? 'success' : 'secondary'}>
                      {provider.enabled ? '启用' : '禁用'}
                    </Badge>
                  </td>
                  <td>
                    <Badge
                      bg={provider.supports_stream ? 'success' : 'secondary'}>
                      {provider.supports_stream ? '支持' : '不支持'}
                    </Badge>
                  </td>
                  <td>{provider.models?.length || 0}</td>
                  <td className="ai-chat-config-action-cell">
                    <Button
                      size="sm"
                      variant="outline-primary"
                      onClick={() => setProviderForm(provider)}>
                      编辑
                    </Button>
                    <Button
                      size="sm"
                      variant="outline-secondary"
                      disabled={fetchingProviderID === provider.id}
                      onClick={() => fetchModels(provider.id)}>
                      获取模型列表
                    </Button>
                    <Button
                      size="sm"
                      variant="outline-secondary"
                      disabled={!provider.models?.length}
                      onClick={() => openTestProvider(provider)}>
                      测试模型
                    </Button>
                    <Button
                      size="sm"
                      variant="outline-danger"
                      onClick={() =>
                        deleteAiChatProvider(provider.id).then(loadAll)
                      }>
                      删除
                    </Button>
                  </td>
                </tr>
              ))}
            </tbody>
          </Table>
        </Tab>

        <Tab eventKey="mappings" title="模型映射">
          <Card className="mb-4">
            <Card.Body>
              <Form onSubmit={submitMapping}>
                <Row>
                  <Col md={3}>
                    <Form.Group className="mb-3">
                      <Form.Label>本站模型 ID</Form.Label>
                      <Form.Control
                        required
                        placeholder="fast-chat"
                        value={mappingForm.site_model_id}
                        onChange={(e) =>
                          setMappingForm({
                            ...mappingForm,
                            site_model_id: e.target.value,
                          })
                        }
                      />
                    </Form.Group>
                  </Col>
                  <Col md={3}>
                    <Form.Group className="mb-3">
                      <Form.Label>显示名称</Form.Label>
                      <Form.Control
                        value={mappingForm.display_name}
                        onChange={(e) =>
                          setMappingForm({
                            ...mappingForm,
                            display_name: e.target.value,
                          })
                        }
                      />
                    </Form.Group>
                  </Col>
                  <Col md={2}>
                    <Form.Group className="mb-3">
                      <Form.Label>排序权重</Form.Label>
                      <Form.Control
                        type="number"
                        value={mappingForm.sort_order}
                        onChange={(e) =>
                          setMappingForm({
                            ...mappingForm,
                            sort_order: Number(e.target.value),
                          })
                        }
                      />
                    </Form.Group>
                  </Col>
                  <Col md={4}>
                    <Form.Group className="mb-3">
                      <Form.Label>默认上游模型</Form.Label>
                      <Form.Select
                        value={mappingForm.default_provider_model_id}
                        onChange={(e) =>
                          setMappingForm({
                            ...mappingForm,
                            default_provider_model_id: e.target.value,
                          })
                        }>
                        <option value="">按优先级</option>
                        {mappingForm.items.map((item) =>
                          item.provider_model_id ? (
                            <option
                              key={`${item.provider_id}-${item.provider_model_id}-${item.priority}`}
                              value={item.provider_model_id}>
                              {item.provider_model_id}
                            </option>
                          ) : null,
                        )}
                      </Form.Select>
                    </Form.Group>
                  </Col>
                </Row>
                <Form.Group className="mb-3">
                  <Form.Label>模型说明</Form.Label>
                  <Form.Control
                    as="textarea"
                    rows={2}
                    value={mappingForm.description}
                    onChange={(e) =>
                      setMappingForm({
                        ...mappingForm,
                        description: e.target.value,
                      })
                    }
                  />
                </Form.Group>
                {mappingForm.items.map((item, index) => (
                  <Row
                    key={
                      item.id ||
                      item.client_id ||
                      `${item.provider_id}-${item.provider_model_id}-${item.priority}`
                    }
                    className="align-items-end">
                    <Col md={6}>
                      <Form.Group className="mb-3">
                        <Form.Label>上游模型</Form.Label>
                        <Form.Select
                          required
                          value={`${item.provider_id}|${item.provider_model_id}`}
                          onChange={(e) => {
                            const [providerID, modelID] =
                              e.target.value.split('|');
                            updateMappingItem(index, {
                              provider_id: Number(providerID),
                              provider_model_id: modelID,
                            });
                          }}>
                          <option value="0|">请选择</option>
                          {upstreamOptions.map((option) => (
                            <option
                              key={`${option.provider_id}-${option.provider_model_id}`}
                              value={`${option.provider_id}|${option.provider_model_id}`}>
                              {option.label}
                            </option>
                          ))}
                        </Form.Select>
                      </Form.Group>
                    </Col>
                    <Col md={2}>
                      <Form.Group className="mb-3">
                        <Form.Label>优先级</Form.Label>
                        <Form.Control
                          type="number"
                          value={item.priority}
                          onChange={(e) =>
                            updateMappingItem(index, {
                              priority: Number(e.target.value),
                            })
                          }
                        />
                      </Form.Group>
                    </Col>
                    <Col md={2}>
                      <Form.Check
                        className="mb-3"
                        type="switch"
                        label="启用"
                        checked={item.enabled}
                        onChange={(e) =>
                          updateMappingItem(index, {
                            enabled: e.target.checked,
                          })
                        }
                      />
                    </Col>
                    <Col md={2}>
                      <Button
                        className="mb-3"
                        variant="outline-danger"
                        disabled={mappingForm.items.length === 1}
                        onClick={() =>
                          setMappingForm({
                            ...mappingForm,
                            items: mappingForm.items.filter(
                              (_, itemIndex) => itemIndex !== index,
                            ),
                          })
                        }>
                        删除
                      </Button>
                    </Col>
                  </Row>
                ))}
                <Button
                  type="button"
                  variant="outline-secondary"
                  className="me-2"
                  onClick={() =>
                    setMappingForm({
                      ...mappingForm,
                      items: [
                        ...mappingForm.items,
                        newMappingItem(mappingForm.items.length + 1),
                      ],
                    })
                  }>
                  添加上游模型
                </Button>
                <Form.Check
                  inline
                  type="switch"
                  label="启用映射"
                  checked={mappingForm.enabled}
                  onChange={(e) =>
                    setMappingForm({
                      ...mappingForm,
                      enabled: e.target.checked,
                    })
                  }
                />
                <Form.Check
                  inline
                  type="switch"
                  label="支持图片理解"
                  checked={mappingForm.supports_vision}
                  onChange={(e) =>
                    setMappingForm({
                      ...mappingForm,
                      supports_vision: e.target.checked,
                    })
                  }
                />
                <Form.Check
                  inline
                  type="switch"
                  label="失败自动切换备用"
                  checked={mappingForm.fallback_enabled}
                  onChange={(e) =>
                    setMappingForm({
                      ...mappingForm,
                      fallback_enabled: e.target.checked,
                    })
                  }
                />
                <div className="mt-3">
                  <Button type="submit">保存模型映射</Button>
                </div>
              </Form>
            </Card.Body>
          </Card>
          <Table responsive hover>
            <thead>
              <tr>
                <th>本站模型 ID</th>
                <th>显示名称</th>
                <th>上游模型</th>
                <th>能力</th>
                <th>状态</th>
                <th>操作</th>
              </tr>
            </thead>
            <tbody>
              {mappings.map((mapping) => (
                <tr key={mapping.id}>
                  <td>{mapping.site_model_id}</td>
                  <td>{mapping.display_name}</td>
                  <td>
                    {(mapping.items || []).map((item) => (
                      <Badge
                        bg="light"
                        text="dark"
                        className="me-1"
                        key={item.id}>
                        {item.provider_name}/{item.provider_model_id}
                      </Badge>
                    ))}
                  </td>
                  <td>
                    {mapping.supports_vision ? (
                      <Badge bg="info">图片理解</Badge>
                    ) : (
                      <span className="text-muted">文本</span>
                    )}
                  </td>
                  <td>{mapping.enabled ? '启用' : '禁用'}</td>
                  <td>
                    <Button
                      size="sm"
                      variant="outline-primary"
                      className="me-2"
                      onClick={() => setMappingForm(mapping)}>
                      编辑
                    </Button>
                    <Button
                      size="sm"
                      variant="outline-danger"
                      onClick={() =>
                        deleteAiChatModelMapping(mapping.id).then(loadAll)
                      }>
                      删除
                    </Button>
                  </td>
                </tr>
              ))}
            </tbody>
          </Table>
        </Tab>

        <Tab eventKey="plans" title="用户等级 / 订阅配置">
          {extraPlanCount >= 3 && !planForm.id ? (
            <div className="ai-chat-config-notice" role="alert">
              最多只能额外添加 3 个订阅等级。
            </div>
          ) : null}
          <Card className="mb-4">
            <Card.Body>
              <Form onSubmit={submitPlan}>
                <Row>
                  <Col md={3}>
                    <Form.Group className="mb-3">
                      <Form.Label>等级 ID</Form.Label>
                      <Form.Control
                        required
                        disabled={planForm.plan_id === 'free'}
                        value={planForm.plan_id}
                        onChange={(e) =>
                          setPlanForm({ ...planForm, plan_id: e.target.value })
                        }
                      />
                    </Form.Group>
                  </Col>
                  <Col md={3}>
                    <Form.Group className="mb-3">
                      <Form.Label>等级名称</Form.Label>
                      <Form.Control
                        required
                        value={planForm.name}
                        onChange={(e) =>
                          setPlanForm({ ...planForm, name: e.target.value })
                        }
                      />
                    </Form.Group>
                  </Col>
                  <Col md={2}>
                    <Form.Group className="mb-3">
                      <Form.Label>月费用</Form.Label>
                      <Form.Control
                        type="number"
                        min="0"
                        value={planForm.monthly_price}
                        onChange={(e) =>
                          setPlanForm({
                            ...planForm,
                            monthly_price: Number(e.target.value),
                          })
                        }
                      />
                    </Form.Group>
                  </Col>
                  <Col md={2}>
                    <Form.Group className="mb-3">
                      <Form.Label>聊天点数</Form.Label>
                      <Form.Control
                        type="number"
                        min="-1"
                        value={planForm.chat_points}
                        onChange={(e) =>
                          setPlanForm({
                            ...planForm,
                            chat_points: Number(e.target.value),
                          })
                        }
                      />
                      <Form.Text className="text-muted">
                        -1 表示无限制
                      </Form.Text>
                    </Form.Group>
                  </Col>
                  <Col md={2}>
                    <Form.Group className="mb-3">
                      <Form.Label>图片张数</Form.Label>
                      <Form.Control
                        type="number"
                        min="0"
                        value={planForm.image_quota}
                        onChange={(e) =>
                          setPlanForm({
                            ...planForm,
                            image_quota: Number(e.target.value),
                          })
                        }
                      />
                    </Form.Group>
                  </Col>
                </Row>
                <Form.Group className="mb-3">
                  <Form.Label>购买链接</Form.Label>
                  <Form.Control
                    type="url"
                    placeholder="https://..."
                    value={planForm.purchase_url}
                    onChange={(e) =>
                      setPlanForm({
                        ...planForm,
                        purchase_url: e.target.value,
                      })
                    }
                  />
                </Form.Group>
                <Form.Group className="mb-3">
                  <Form.Label>可用模型列表</Form.Label>
                  <div className="ai-chat-config-checkbox-list">
                    {mappings.map((mapping) => {
                      const checked = planForm.model_mapping_ids.includes(
                        mapping.id,
                      );
                      return (
                        <Form.Check
                          key={mapping.id}
                          type="checkbox"
                          id={`plan-model-${mapping.id}`}
                          label={mapping.site_model_id}
                          checked={checked}
                          onChange={(e) => {
                            setPlanForm({
                              ...planForm,
                              model_mapping_ids: e.target.checked
                                ? [...planForm.model_mapping_ids, mapping.id]
                                : planForm.model_mapping_ids.filter(
                                    (id) => id !== mapping.id,
                                  ),
                            });
                          }}
                        />
                      );
                    })}
                    {mappings.length === 0 ? (
                      <div className="text-muted">请先创建模型映射</div>
                    ) : null}
                  </div>
                </Form.Group>
                <Form.Group className="mb-3">
                  <Form.Label>适合的任务说明</Form.Label>
                  <Form.Control
                    as="textarea"
                    rows={2}
                    value={planForm.task_description}
                    onChange={(e) =>
                      setPlanForm({
                        ...planForm,
                        task_description: e.target.value,
                      })
                    }
                  />
                </Form.Group>
                <Form.Check
                  className="mb-3"
                  type="switch"
                  label="启用"
                  checked={planForm.enabled}
                  onChange={(e) =>
                    setPlanForm({ ...planForm, enabled: e.target.checked })
                  }
                />
                <Button
                  type="submit"
                  disabled={extraPlanCount >= 3 && !planForm.id}>
                  保存订阅等级
                </Button>
              </Form>
            </Card.Body>
          </Card>
          <Table responsive hover>
            <thead>
              <tr>
                <th>等级</th>
                <th>价格</th>
                <th>点数 / 图片</th>
                <th>购买链接</th>
                <th>可用模型</th>
                <th>状态</th>
                <th>操作</th>
              </tr>
            </thead>
            <tbody>
              {plans.map((plan) => (
                <tr key={plan.id}>
                  <td>
                    {plan.name}{' '}
                    <span className="text-muted">({plan.plan_id})</span>
                  </td>
                  <td>{plan.monthly_price}</td>
                  <td>
                    {formatQuota(plan.chat_points)} / {plan.image_quota}
                  </td>
                  <td>{plan.purchase_url ? '已配置' : '未配置'}</td>
                  <td>{(plan.available_model_ids || []).join(', ')}</td>
                  <td>{plan.enabled ? '启用' : '禁用'}</td>
                  <td>
                    <Button
                      size="sm"
                      variant="outline-primary"
                      className="me-2"
                      onClick={() => setPlanForm(plan)}>
                      编辑
                    </Button>
                    <Button
                      size="sm"
                      variant="outline-danger"
                      disabled={plan.plan_id === 'free'}
                      onClick={() =>
                        deleteAiChatSubscriptionPlan(plan.id).then(loadAll)
                      }>
                      删除
                    </Button>
                  </td>
                </tr>
              ))}
            </tbody>
          </Table>
        </Tab>

        <Tab eventKey="redeem-codes" title="订阅兑换码">
          <Card className="mb-4">
            <Card.Body>
              <Form onSubmit={submitRedeemCodes}>
                <Row>
                  <Col md={3}>
                    <Form.Group className="mb-3">
                      <Form.Label>订阅等级</Form.Label>
                      <Form.Select
                        required
                        value={redeemForm.plan_id}
                        onChange={(e) =>
                          setRedeemForm({
                            ...redeemForm,
                            plan_id: Number(e.target.value),
                          })
                        }>
                        <option value={0}>请选择</option>
                        {paidPlans.map((plan) => (
                          <option key={plan.id} value={plan.id}>
                            {plan.name} ({plan.plan_id})
                          </option>
                        ))}
                      </Form.Select>
                    </Form.Group>
                  </Col>
                  <Col md={2}>
                    <Form.Group className="mb-3">
                      <Form.Label>生成数量</Form.Label>
                      <Form.Control
                        required
                        min={1}
                        max={500}
                        type="number"
                        value={redeemForm.count}
                        onChange={(e) =>
                          setRedeemForm({
                            ...redeemForm,
                            count: Number(e.target.value),
                          })
                        }
                      />
                    </Form.Group>
                  </Col>
                  <Col md={2}>
                    <Form.Group className="mb-3">
                      <Form.Label>有效月数</Form.Label>
                      <Form.Control
                        required
                        min={1}
                        max={120}
                        type="number"
                        value={redeemForm.duration_months}
                        onChange={(e) =>
                          setRedeemForm({
                            ...redeemForm,
                            duration_months: Number(e.target.value),
                          })
                        }
                      />
                    </Form.Group>
                  </Col>
                  <Col md={2}>
                    <Form.Group className="mb-3">
                      <Form.Label>前缀</Form.Label>
                      <Form.Control
                        placeholder="PLUS"
                        value={redeemForm.prefix}
                        onChange={(e) =>
                          setRedeemForm({
                            ...redeemForm,
                            prefix: e.target.value,
                          })
                        }
                      />
                    </Form.Group>
                  </Col>
                  <Col md={3}>
                    <Form.Group className="mb-3">
                      <Form.Label>备注</Form.Label>
                      <Form.Control
                        value={redeemForm.remark}
                        onChange={(e) =>
                          setRedeemForm({
                            ...redeemForm,
                            remark: e.target.value,
                          })
                        }
                      />
                    </Form.Group>
                  </Col>
                </Row>
                <Button type="submit" disabled={paidPlans.length === 0}>
                  批量生成兑换码
                </Button>
                {paidPlans.length === 0 ? (
                  <span className="text-muted ms-3">请先创建付费订阅等级</span>
                ) : null}
              </Form>
            </Card.Body>
          </Card>

          {generatedCodes.length > 0 ? (
            <Card className="mb-4">
              <Card.Body>
                <div className="d-flex align-items-center justify-content-between mb-2">
                  <strong>本次生成</strong>
                  <Badge bg="primary">{generatedCodes.length} 个</Badge>
                </div>
                <Form.Control
                  as="textarea"
                  readOnly
                  rows={Math.min(8, generatedCodes.length)}
                  className="ai-chat-config-code-output"
                  value={generatedCodes.map((item) => item.code).join('\n')}
                />
              </Card.Body>
            </Card>
          ) : null}

          <Table responsive hover className="ai-chat-config-redeem-table">
            <thead>
              <tr>
                <th>兑换码</th>
                <th>等级</th>
                <th>月数</th>
                <th>状态</th>
                <th>使用用户</th>
                <th>使用时间</th>
                <th>批次</th>
                <th>备注</th>
              </tr>
            </thead>
            <tbody>
              {redeemCodes.map((code) => (
                <tr key={code.id}>
                  <td className="ai-chat-config-code-cell">{code.code}</td>
                  <td>
                    {code.plan_name || '-'}{' '}
                    <span className="text-muted">({code.plan_key || '-'})</span>
                  </td>
                  <td>{code.duration_months}</td>
                  <td>
                    <Badge bg={code.used ? 'secondary' : 'success'}>
                      {code.used ? '已使用' : '未使用'}
                    </Badge>
                  </td>
                  <td>{code.used_by_user_id || '-'}</td>
                  <td>{formatDateTime(code.used_at)}</td>
                  <td
                    className="ai-chat-config-text-cell"
                    title={code.batch_no || '-'}>
                    {code.batch_no || '-'}
                  </td>
                  <td
                    className="ai-chat-config-text-cell"
                    title={code.remark || '-'}>
                    {code.remark || '-'}
                  </td>
                </tr>
              ))}
              {redeemCodes.length === 0 ? (
                <tr>
                  <td colSpan={8} className="text-muted">
                    暂无兑换码
                  </td>
                </tr>
              ) : null}
            </tbody>
          </Table>
        </Tab>

        <Tab eventKey="rates" title="模型消耗系数">
          <Card className="mb-4">
            <Card.Body>
              <Form onSubmit={submitRate}>
                <Row>
                  <Col md={4}>
                    <Form.Group className="mb-3">
                      <Form.Label>本站模型 ID</Form.Label>
                      <Form.Select
                        required
                        value={rateForm.model_mapping_id}
                        onChange={(e) =>
                          setRateForm({
                            ...rateForm,
                            model_mapping_id: Number(e.target.value),
                          })
                        }>
                        <option value={0}>请选择</option>
                        {mappings
                          .filter((mapping) => mapping.enabled)
                          .map((mapping) => (
                            <option key={mapping.id} value={mapping.id}>
                              {mapping.site_model_id}
                            </option>
                          ))}
                      </Form.Select>
                    </Form.Group>
                  </Col>
                  <Col md={3}>
                    <Form.Group className="mb-3">
                      <Form.Label>消耗系数</Form.Label>
                      <Form.Control
                        required
                        min="0.01"
                        step="any"
                        type="number"
                        value={rateForm.consume_rate}
                        onChange={(e) =>
                          setRateForm({
                            ...rateForm,
                            consume_rate: Number(e.target.value),
                          })
                        }
                      />
                    </Form.Group>
                  </Col>
                  <Col md={5}>
                    <Form.Group className="mb-3">
                      <Form.Label>备注</Form.Label>
                      <Form.Control
                        value={rateForm.remark}
                        onChange={(e) =>
                          setRateForm({
                            ...rateForm,
                            remark: e.target.value,
                          })
                        }
                      />
                    </Form.Group>
                  </Col>
                </Row>
                <Form.Check
                  className="mb-3"
                  type="switch"
                  label="启用"
                  checked={rateForm.enabled}
                  onChange={(e) =>
                    setRateForm({ ...rateForm, enabled: e.target.checked })
                  }
                />
                <Button type="submit">保存消耗系数</Button>
              </Form>
            </Card.Body>
          </Card>
          <Table responsive hover>
            <thead>
              <tr>
                <th>本站模型 ID</th>
                <th>消耗系数</th>
                <th>状态</th>
                <th>备注</th>
                <th>操作</th>
              </tr>
            </thead>
            <tbody>
              {rates.map((rate) => (
                <tr key={rate.id}>
                  <td>{rate.site_model_id}</td>
                  <td>{rate.consume_rate}</td>
                  <td>{rate.enabled ? '启用' : '禁用'}</td>
                  <td>{rate.remark}</td>
                  <td>
                    <Button
                      size="sm"
                      variant="outline-primary"
                      onClick={() => setRateForm(rate)}>
                      编辑
                    </Button>
                  </td>
                </tr>
              ))}
            </tbody>
          </Table>
        </Tab>
      </Tabs>
      <Modal show={!!testingProvider} onHide={closeTestProvider} centered>
        <Modal.Header closeButton={!testing}>
          <Modal.Title>测试模型</Modal.Title>
        </Modal.Header>
        <Modal.Body>
          <Form.Group className="mb-3">
            <Form.Label>Provider</Form.Label>
            <Form.Control readOnly value={testingProvider?.name || ''} />
          </Form.Group>
          <Form.Group className="mb-3">
            <Form.Label>模型</Form.Label>
            <Form.Select
              value={testingModelID}
              disabled={testing}
              onChange={(e) => {
                setTestingModelID(e.target.value);
                setTestingResult(null);
              }}>
              {(testingProvider?.models || []).map((model) => (
                <option
                  key={model.provider_model_id}
                  value={model.provider_model_id}>
                  {model.provider_model_id}
                </option>
              ))}
            </Form.Select>
          </Form.Group>
          <div className="text-muted mb-2">测试消息：hi</div>
          {testingResult?.error ? (
            <Alert variant="danger">{testingResult.error}</Alert>
          ) : null}
          {testingResult?.message ? (
            <div className="ai-chat-config-test-result">
              {testingResult.message}
            </div>
          ) : null}
        </Modal.Body>
        <Modal.Footer>
          <Button
            type="button"
            variant="link"
            disabled={testing}
            onClick={closeTestProvider}>
            关闭
          </Button>
          <Button
            type="button"
            disabled={!testingModelID || testing}
            onClick={testProviderModel}>
            {testing ? (
              <Spinner animation="border" size="sm" className="me-2" />
            ) : null}
            开始测试
          </Button>
        </Modal.Footer>
      </Modal>
    </div>
  );
};

export default AiChatConfig;
