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

import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';

import type { KeyedMutator } from 'swr';

import type * as Type from '@/common/interface';
import { Modal } from '@/components';
import {
  deleteTag,
  editCheck,
  mergeTag,
  unDeleteTag,
  useQuerySynonymsTags,
} from '@/services';
import { pathFactory } from '@/router/pathFactory';

type TagInfoMutator = KeyedMutator<Type.TagInfo>;

interface UseTagManageActionsProps {
  tagInfo?: Type.TagInfo;
  refreshTagInfo?: TagInfoMutator;
}

export const useTagManageActions = ({
  tagInfo,
  refreshTagInfo,
}: UseTagManageActionsProps) => {
  const navigate = useNavigate();
  const { t } = useTranslation('translation', { keyPrefix: 'tag_info' });
  const [showMergeModal, setShowMergeModal] = useState(false);
  const { data: synonymsData, mutate: mutateSynonymsData } =
    useQuerySynonymsTags(tagInfo?.tag_id, tagInfo?.status);

  const handleEditTag = () => {
    if (!tagInfo?.tag_id) {
      return;
    }
    editCheck(tagInfo.tag_id).then(() => {
      navigate(pathFactory.tagEdit(tagInfo.tag_id));
    });
  };

  const handleDeleteTag = () => {
    if (!tagInfo) {
      return;
    }
    if (synonymsData?.synonyms && synonymsData.synonyms.length > 0) {
      Modal.confirm({
        title: t('delete.title'),
        content: t('delete.tip_with_synonyms'),
        showConfirm: false,
        cancelText: t('delete.close'),
      });
      return;
    }
    if (tagInfo.question_count > 0) {
      Modal.confirm({
        title: t('delete.title'),
        content: t('delete.tip_with_posts'),
        showConfirm: false,
        cancelText: t('delete.close'),
      });
      return;
    }

    Modal.confirm({
      title: t('delete.title'),
      content: t('delete.tip'),
      confirmText: t('delete', { keyPrefix: 'btns' }),
      confirmBtnVariant: 'danger',
      onConfirm: () => {
        deleteTag(tagInfo.tag_id).then(() => {
          navigate('/tags', { replace: true });
        });
      },
    });
  };

  const handleMergeConfirm = (sourceTagID: string, targetTagID: string) => {
    mergeTag({ source_tag_id: sourceTagID, target_tag_id: targetTagID }).then(
      () => {
        setShowMergeModal(false);
        navigate('/tags', { replace: true });
      },
    );
  };

  const onAction = (params: Type.MemberActionItem) => {
    if (!tagInfo) {
      return;
    }
    if (params.action === 'edit') {
      handleEditTag();
    }
    if (params.action === 'delete') {
      handleDeleteTag();
    }
    if (params.action === 'merge') {
      setShowMergeModal(true);
    }
    if (params.action === 'undelete') {
      Modal.confirm({
        title: t('undelete_title', { keyPrefix: 'delete' }),
        content: t('undelete_desc', { keyPrefix: 'delete' }),
        cancelBtnVariant: 'link',
        confirmBtnVariant: 'danger',
        confirmText: t('undelete', { keyPrefix: 'btns' }),
        onConfirm: () => {
          unDeleteTag(tagInfo.tag_id).then(() => {
            refreshTagInfo?.();
          });
        },
      });
    }
  };

  return {
    showMergeModal,
    setShowMergeModal,
    synonymsData,
    mutateSynonymsData,
    onAction,
    handleMergeConfirm,
  };
};
