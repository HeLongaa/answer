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

import { FC, memo, useEffect, useState } from 'react';
import { Link, useLocation } from 'react-router-dom';

import type { AiSubscriptionOverview } from '@/common/interface';
import { getAiSubscriptionOverview } from '@/services';
import Icon from '../Icon';
import './index.scss';

interface Props {
  visible?: boolean;
}

const formatQuota = (value?: number) => {
  if (value === -1) {
    return '不限';
  }
  return String(value || 0);
};

const AiSubscriptionPill: FC<Props> = ({ visible = true }) => {
  const location = useLocation();
  const [subscription, setSubscription] =
    useState<AiSubscriptionOverview | null>(null);

  const refreshSubscription = () => {
    if (!visible) {
      return;
    }
    getAiSubscriptionOverview()
      .then(setSubscription)
      .catch(() => {
        setSubscription(null);
      });
  };

  useEffect(() => {
    refreshSubscription();
  }, [visible]);

  useEffect(() => {
    window.addEventListener('hcai-subscription-updated', refreshSubscription);
    return () => {
      window.removeEventListener(
        'hcai-subscription-updated',
        refreshSubscription,
      );
    };
  }, [visible]);

  if (!visible) {
    return null;
  }

  const handleClick = (evt) => {
    if (location.pathname !== '/') {
      return;
    }
    evt.preventDefault();
    window.dispatchEvent(new CustomEvent('hcai-open-subscription'));
  };

  return (
    <Link
      className="ai-subscription-pill"
      to="/subscription"
      onClick={handleClick}>
      <Icon name="stars" />
      <span className="ai-subscription-pill-plan">
        {subscription?.plan_name || 'AI 订阅'}
      </span>
      <span className="ai-subscription-pill-quota">
        剩余 {formatQuota(subscription?.chat_points_remaining)} 点
      </span>
    </Link>
  );
};

export default memo(AiSubscriptionPill);
