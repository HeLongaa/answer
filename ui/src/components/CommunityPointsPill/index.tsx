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

import { usePointAccount } from '@/services';
import { formatCount } from '@/utils';
import Icon from '../Icon';
import './index.scss';

interface Props {
  visible?: boolean;
}

const CommunityPointsPill: FC<Props> = ({ visible = true }) => {
  const { data: pointAccount } = usePointAccount();

  if (!visible) {
    return null;
  }

  return (
    <Link className="community-points-pill" to="/users/settings/points">
      <Icon name="coin" />
      <span className="community-points-pill-label">贡献积分</span>
      <span className="community-points-pill-balance">
        {formatCount(pointAccount?.balance || 0)}
      </span>
    </Link>
  );
};

export default memo(CommunityPointsPill);
