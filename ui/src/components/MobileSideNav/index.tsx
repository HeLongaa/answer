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

import { NavLink, useLocation } from 'react-router-dom';

import classNames from 'classnames';

import { SideNav, AdminSideNav, Icon } from '@/components';
import { siteInfoStore } from '@/stores';

import './index.scss';

const chatNavItems = [
  { icon: 'pencil-square', label: '新对话', active: true },
  { icon: 'search', label: '搜索' },
  { icon: 'image', label: '图片生成' },
  { icon: 'credit-card-2-front', label: '订阅管理' },
  { icon: 'stars', label: '订阅兑换' },
  { icon: 'grid', label: '工作空间' },
];

const conversations = ['随意内容小片段', '新对话', '随意内容中文问候'];

const MobileSideNav = ({ show, onHide }) => {
  const { pathname } = useLocation();
  const isAdmin = pathname.includes('/admin');
  const isChat = pathname === '/';
  const isUserSideNavPage =
    pathname === '/users' ||
    /^\/users\/[^/]+(\/(answers|questions|bookmarks|reputation|badges|votes))?$/.test(
      pathname,
    );
  const isCommunityPage =
    pathname.startsWith('/questions') ||
    pathname.startsWith('/tags') ||
    isUserSideNavPage ||
    pathname.startsWith('/badges') ||
    pathname.startsWith('/review');
  const siteInfo = siteInfoStore((state) => state.siteInfo);

  const closeSideNav = () => onHide(false);

  return (
    <>
      <button
        type="button"
        className={classNames('mobile-side-nav-scrim', { show })}
        aria-hidden={!show}
        tabIndex={show ? 0 : -1}
        onClick={closeSideNav}
      />

      <aside
        id="mobileSideNav"
        aria-hidden={!show}
        className={classNames('px-3 py-4', { show })}>
        <NavLink
          to="/"
          className="mobile-side-site-brand"
          onClick={closeSideNav}>
          {siteInfo.name}
        </NavLink>

        <div className="mobile-page-switch" aria-label="页面切换">
          <NavLink
            to="/"
            end
            onClick={closeSideNav}
            className={pathname === '/' ? 'active' : ''}>
            CHAT
          </NavLink>
          <NavLink
            to="/questions"
            onClick={closeSideNav}
            className={isCommunityPage ? 'active' : ''}>
            社区
          </NavLink>
          <button type="button" aria-disabled="true">
            支持
          </button>
        </div>
        <button
          type="button"
          className="mobile-upgrade-switch"
          aria-disabled="true">
          <Icon name="music-note-beamed" />
          <span>升级套餐</span>
        </button>

        {isChat ? (
          <div className="mobile-chat-nav" aria-label="HCAI-Chat navigation">
            <nav className="mobile-chat-nav-main">
              {chatNavItems.map((item) => (
                <button
                  type="button"
                  className={item.active ? 'active' : ''}
                  key={item.label}>
                  <Icon name={item.icon} />
                  <span>{item.label}</span>
                </button>
              ))}
            </nav>

            <div className="mobile-chat-section">
              <div className="mobile-chat-section-title">
                <Icon name="chevron-down" />
                <span>频道</span>
              </div>
              <button type="button" className="mobile-chat-item">
                <span>#</span>
                <span>bug-fix</span>
              </button>
            </div>

            <div className="mobile-chat-section">
              <div className="mobile-chat-section-title">
                <Icon name="chevron-down" />
                <span>对话</span>
              </div>
              <span className="mobile-chat-time">过去 7 天</span>
              {conversations.map((item) => (
                <button type="button" className="mobile-chat-item" key={item}>
                  {item}
                </button>
              ))}
            </div>
          </div>
        ) : isAdmin ? (
          <AdminSideNav showBrand={false} />
        ) : (
          <SideNav showBrand={false} />
        )}
      </aside>
    </>
  );
};

export default MobileSideNav;
