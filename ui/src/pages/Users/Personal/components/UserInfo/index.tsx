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
import { OverlayTrigger, Tooltip } from 'react-bootstrap';
import { useTranslation } from 'react-i18next';
import { Link } from 'react-router-dom';

import classnames from 'classnames';

import { Avatar, Icon, SvgIcon } from '@/components';
import type { AiSubscriptionOverview, UserInfoRes } from '@/common/interface';
import {
  getAiSubscriptionOverview,
  getUcBranding,
  UcBrandingEntry,
  usePointAccount,
} from '@/services';
import { userCenterStore } from '@/stores';

interface Props {
  data: UserInfoRes;
  isSelf?: boolean;
}

const formatQuota = (value?: number) => {
  if (value === -1) {
    return '不限';
  }
  return String(value || 0);
};

const PersonalAiSummary: FC<{ username: string }> = ({ username }) => {
  const [aiSubscription, setAiSubscription] =
    useState<AiSubscriptionOverview | null>(null);
  const { data: pointAccount } = usePointAccount();

  useEffect(() => {
    getAiSubscriptionOverview()
      .then(setAiSubscription)
      .catch(() => {
        setAiSubscription(null);
      });
  }, [username]);

  return (
    <div className="personal-ai-summary mb-4">
      <div className="personal-ai-summary-item personal-ai-plan">
        <span className="personal-ai-summary-label">AI 订阅等级</span>
        <strong>{aiSubscription?.plan_name || '免费版'}</strong>
        <span>
          剩余 {formatQuota(aiSubscription?.chat_points_remaining)} 点 · 视频{' '}
          {formatQuota(aiSubscription?.video_quota_remaining)} 次
        </span>
      </div>
      <div className="personal-ai-summary-item personal-points">
        <span className="personal-ai-summary-label">积分余额</span>
        <strong>{pointAccount?.balance || 0}</strong>
        <span>可用于任务与悬赏</span>
      </div>
    </div>
  );
};

const Index: FC<Props> = ({ data, isSelf = false }) => {
  const { t } = useTranslation('translation', { keyPrefix: 'personal' });
  const { agent: ucAgent } = userCenterStore();
  const [ucBranding, setUcBranding] = useState<UcBrandingEntry[]>([]);

  const initData = () => {
    if (ucAgent?.enabled && data?.username) {
      getUcBranding(data.username).then((resp) => {
        if (resp.enabled && Array.isArray(resp.personal_branding)) {
          setUcBranding(resp.personal_branding);
        }
      });
    }
  };

  useEffect(() => {
    initData();
  }, [data?.username]);

  if (!data?.username) {
    return null;
  }
  return (
    <div className="personal-user-info d-flex flex-column flex-md-row">
      {data?.status !== 'deleted' ? (
        <Link to={`/users/${data.username}`} reloadDocument>
          <Avatar
            avatar={data.avatar}
            size="160px"
            searchStr="s=256"
            alt={data.display_name}
          />
        </Link>
      ) : (
        <Avatar
          avatar={data.avatar}
          size="160px"
          searchStr="s=256"
          alt={data.display_name}
        />
      )}

      <div className="ms-0 ms-md-4 mt-4 mt-md-0">
        <div className="d-flex align-items-center mb-2">
          {data?.status !== 'deleted' ? (
            <Link
              to={`/users/${data.username}`}
              className="link-dark h3 mb-0"
              reloadDocument>
              {data.display_name}
            </Link>
          ) : (
            <span className="link-dark h3 mb-0">{data.display_name}</span>
          )}
          {data?.role_id === 2 && (
            <div className="ms-2">
              <OverlayTrigger
                placement="top"
                overlay={<Tooltip>{t('mod_long')}</Tooltip>}>
                <span className="badge text-bg-light">{t('mod_short')}</span>
              </OverlayTrigger>
            </div>
          )}
        </div>
        <div className="text-secondary mb-4">@{data.username}</div>

        {isSelf ? <PersonalAiSummary username={data.username} /> : null}

        <div className="d-flex flex-wrap mb-3">
          <div className="me-3">
            <strong className="fs-5">{data.rank || 0}</strong>
            <span className="text-secondary"> {t('x_reputation')}</span>
          </div>
          <div className="me-3">
            <strong className="fs-5">{data.answer_count || 0}</strong>
            <span className="text-secondary"> {t('x_answers')}</span>
          </div>
          <div>
            <strong className="fs-5">{data?.question_count || 0}</strong>
            <span className="text-secondary"> {t('x_questions')}</span>
          </div>
        </div>

        <div className="d-flex text-secondary">
          {!ucAgent?.enabled ? (
            <>
              {data.location && (
                <div className="d-flex align-items-center me-3">
                  <Icon name="geo-alt-fill" className="me-2" />
                  <span>{data.location}</span>
                </div>
              )}
              {data.website && (
                <div className="d-flex align-items-center">
                  <Icon name="house-door-fill" className="me-2" />
                  <a
                    className="link-secondary"
                    href={
                      data.website?.startsWith('http')
                        ? data.website
                        : `http://${data.website}`
                    }>
                    {
                      data?.website
                        .replace(/(http|https):\/\//, '')
                        .split('/')?.[0]
                    }
                  </a>
                </div>
              )}
            </>
          ) : null}
          {ucBranding.map((b, i, a) => {
            if (!b.label) {
              return null;
            }
            return (
              <div
                key={b.name}
                className={classnames('d-flex', 'align-items-center', {
                  'me-3': i < a.length - 1,
                })}>
                {b.icon ? (
                  <SvgIcon base64={b.icon} svgClassName="me-2" />
                ) : null}
                {b.url ? (
                  <a className="link-secondary" href={b.url}>
                    {b.label}
                  </a>
                ) : (
                  <span>{b.label}</span>
                )}
              </div>
            );
          })}
        </div>
      </div>
    </div>
  );
};

export default memo(Index);
