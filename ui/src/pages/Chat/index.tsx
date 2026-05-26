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

import { FC, memo } from 'react';
import { Link } from 'react-router-dom';

import { Icon } from '@/components';
import { usePageTags } from '@/hooks';
import { brandingStore, siteInfoStore } from '@/stores';

import './index.scss';

const navItems = [
  { icon: 'pencil-square', label: '新对话', active: true },
  { icon: 'search', label: '搜索' },
  { icon: 'image', label: '图片生成' },
  { icon: 'credit-card-2-front', label: '订阅管理' },
  { icon: 'stars', label: '订阅兑换' },
  { icon: 'grid', label: '工作空间' },
];

const conversations = ['随意内容小片段', '新对话', '随意内容中文问候'];

const suggestions = [
  {
    title: 'Help me study',
    description: 'vocabulary for a college entrance exam',
  },
  {
    title: 'Show me a code snippet',
    description: "of a website's sticky header",
  },
  {
    title: 'Explain options trading',
    description: "if I'm familiar with buying and selling stocks",
  },
];

const Chat: FC = () => {
  const siteInfo = siteInfoStore((state) => state.siteInfo);
  const brandingInfo = brandingStore((state) => state.branding);

  usePageTags({
    title: 'HCAI-Chat',
  });

  const handleSubmit = (evt) => {
    evt.preventDefault();
  };

  return (
    <div className="hcai-chat-page">
      <aside className="hcai-chat-sidebar" aria-label="HCAI-Chat navigation">
        <Link to="/" className="hcai-chat-site-brand">
          {brandingInfo.mobile_logo || brandingInfo.logo ? (
            <img
              className="hcai-chat-site-logo"
              src={brandingInfo.mobile_logo || brandingInfo.logo}
              alt={siteInfo.name}
            />
          ) : (
            <span className="hcai-chat-site-mark">
              {siteInfo.name.slice(0, 1)}
            </span>
          )}
          <span className="hcai-chat-site-name">{siteInfo.name}</span>
        </Link>

        <nav className="hcai-chat-nav">
          {navItems.map((item) => (
            <button
              type="button"
              className={item.active ? 'active' : ''}
              key={item.label}>
              <Icon name={item.icon} />
              <span>{item.label}</span>
            </button>
          ))}
        </nav>

        <div className="hcai-chat-sidebar-section">
          <div className="hcai-chat-sidebar-title">
            <Icon name="chevron-down" />
            <span>频道</span>
          </div>
          <button type="button" className="hcai-channel-item">
            <span>#</span>
            <span>bug-fix</span>
          </button>
        </div>

        <div className="hcai-chat-sidebar-section">
          <div className="hcai-chat-sidebar-title">
            <Icon name="chevron-down" />
            <span>对话</span>
          </div>
          <span className="hcai-chat-time">过去 7 天</span>
          {conversations.map((item) => (
            <button type="button" className="hcai-conversation-item" key={item}>
              {item}
            </button>
          ))}
        </div>
      </aside>

      <main className="hcai-chat-main">
        <div className="hcai-chat-topbar">
          <div className="hcai-chat-actions">
            <button type="button" aria-label="系统设置">
              <Icon name="sliders" />
            </button>
          </div>
        </div>

        <section className="hcai-chat-hero">
          <div className="hcai-chat-title">
            <span className="hcai-chat-logo">OI</span>
            <h1>codex-auto-review</h1>
          </div>

          <form className="hcai-prompt-card" onSubmit={handleSubmit}>
            <input placeholder="有什么我能帮您的吗?" aria-label="聊天输入" />
            <div className="hcai-prompt-tools">
              <div className="hcai-prompt-left">
                <button type="button" aria-label="添加内容">
                  <Icon name="plus-lg" />
                </button>
                <button type="button" className="hcai-tool-pill">
                  <Icon name="terminal" />
                  <span>代码解释器</span>
                </button>
              </div>
              <div className="hcai-prompt-right">
                <button type="button" className="hcai-model-select">
                  <span>codex-auto-review</span>
                  <Icon name="chevron-down" />
                </button>
                <button type="button" aria-label="语音输入">
                  <Icon name="mic" />
                </button>
                <button type="submit" aria-label="发送消息" className="send">
                  <Icon name="arrow-up" />
                </button>
              </div>
            </div>
          </form>

          <div className="hcai-suggestions" aria-label="建议">
            <div className="hcai-suggestions-label">
              <Icon name="lightning-charge" />
              <span>建议</span>
            </div>
            {suggestions.map((item) => (
              <button type="button" key={item.title}>
                <strong>{item.title}</strong>
                <span>{item.description}</span>
              </button>
            ))}
          </div>
        </section>

        <button type="button" className="hcai-help-button" aria-label="帮助">
          ?
        </button>
      </main>
    </div>
  );
};

export default memo(Chat);
