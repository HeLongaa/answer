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

import { FC } from 'react';
import { ListGroup } from 'react-bootstrap';
import { NavLink, useNavigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';

import { BaseUserCard, Counts, FormatTime, Icon, Tag } from '@/components';
import { sortTagsForDisplay } from '@/utils';
import { pathFactory } from '@/router/pathFactory';

interface IProps {
  data: any[];
}

const PinList: FC<IProps> = ({ data }) => {
  const { t } = useTranslation('translation', { keyPrefix: 'question' });
  const navigate = useNavigate();
  if (!data?.length) return null;

  const handleNavigate = (item) => {
    navigate(pathFactory.questionLanding(item.id, item.url_title));
  };

  return (
    <>
      {data.map((item) => (
        <ListGroup.Item
          key={item.id}
          action
          as="li"
          onClick={() => handleNavigate(item)}
          className="question-list-item-pinned py-3 px-2 border-start-0 border-end-0 position-relative pointer">
          <span className="question-pinned-badge">
            <Icon name="pin-angle-fill" />
            {t('pinned')}
          </span>

          <div className="d-flex flex-wrap text-secondary small mb-12">
            <BaseUserCard
              data={item.operator}
              className="me-1"
              avatarClass="me-1"
            />
            •
            <FormatTime
              time={item.operated_at || item.created_at}
              className="text-secondary ms-1 flex-shrink-0"
            />
          </div>

          <h5 className="text-wrap text-break">
            <NavLink
              className="link-dark d-block"
              onClick={(event) => event.stopPropagation()}
              to={pathFactory.questionLanding(item.id, item.url_title)}>
              {item.title}
              {item.status === 2 ? ` [${t('closed')}]` : ''}
            </NavLink>
          </h5>

          <div className="question-tags mb-12">
            {Array.isArray(item.tags)
              ? sortTagsForDisplay(item.tags).map((tag, index, arr) => (
                  <Tag
                    key={tag.slug_name}
                    className={`${arr.length - 1 === index ? '' : 'me-1'}`}
                    data={tag}
                  />
                ))
              : null}
          </div>

          <div className="small text-secondary">
            <Counts
              data={{
                votes: item.vote_count,
                answers: item.answer_count,
                views: item.view_count,
              }}
              isAccepted={item.accepted_answer_id >= 1}
              className="mt-2 mt-md-0"
            />
          </div>
        </ListGroup.Item>
      ))}
    </>
  );
};

export default PinList;
