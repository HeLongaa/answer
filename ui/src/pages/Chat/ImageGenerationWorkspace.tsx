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

import {
  ChangeEvent,
  CSSProperties,
  FC,
  FormEvent,
  useCallback,
  useEffect,
  useMemo,
  useRef,
  useState,
} from 'react';
import { createPortal } from 'react-dom';

import classNames from 'classnames';
import { v4 as uuidv4 } from 'uuid';

import { Icon } from '@/components';
import { LOGGED_TOKEN_STORAGE_KEY } from '@/common/constants';
import type {
  AiImageGeneration,
  AiImageModel,
  AiSubscriptionOverview,
} from '@/common/interface';
import {
  editAiImage,
  generateAiImage,
  getAiImageGenerations,
  getAiImageModels,
} from '@/services/client/ai';
import Storage from '@/utils/storage';

import './imageGeneration.scss';

interface ReferenceImage {
  id: string;
  name: string;
  url: string;
}

type TaskStatus = 'queued' | 'generating' | 'completed' | 'failed';

interface ImageTask {
  id: string;
  prompt: string;
  negativePrompt: string;
  model: string;
  siteModelID: string;
  ratio: string;
  size: string;
  style: string;
  quality: string;
  count: number;
  status: TaskStatus;
  createdAt: number;
  progress: number;
  imageURLs: string[];
  referenceImages: string[];
}

interface IProps {
  subscription: AiSubscriptionOverview | null;
  onRefreshSubscription: () => void;
  onOpenSubscription: () => void;
}

interface EditTarget {
  taskID: string;
  imageIndex: number;
  imageURL: string;
}

interface PreviewImage {
  src: string;
  rawURL: string;
  task: ImageTask;
  imageIndex: number;
}

const sizeOptions = [
  { value: 'auto', label: '自动', meta: '模型决定', ratio: 'auto' },
  { value: '1024x1024', label: '1:1', meta: '1024 × 1024', ratio: '1:1' },
  { value: '1536x1024', label: '3:2', meta: '1536 × 1024', ratio: '3:2' },
  { value: '1024x1536', label: '2:3', meta: '1024 × 1536', ratio: '2:3' },
];

const styleOptions = ['写实', '产品摄影', '插画', '3D', '国风', '极简'];
const qualityOptions = [
  { value: 'auto', label: '自动' },
  { value: 'low', label: '快速' },
  { value: 'medium', label: '标准' },
  { value: 'high', label: '精细' },
];
const countOptions = [1, 2, 4];

const getSizeOption = (size?: string) =>
  sizeOptions.find((option) => option.value === size) || sizeOptions[1];

const readFileAsDataURL = (file: File) =>
  new Promise<string>((resolve, reject) => {
    const reader = new FileReader();
    reader.onload = () => resolve(String(reader.result || ''));
    reader.onerror = () => reject(reader.error);
    reader.readAsDataURL(file);
  });

const getModelName = (model?: AiImageModel) => {
  if (!model) {
    return '选择模型';
  }
  return model.display_name || model.site_model_id;
};

const formatQuota = (value?: number) => {
  if (value === -1) {
    return '无限制';
  }
  return Number(value || 0).toLocaleString();
};

const getTaskStatusLabel = (status: TaskStatus) => {
  if (status === 'queued') {
    return '排队中';
  }
  if (status === 'generating') {
    return '生成中';
  }
  if (status === 'failed') {
    return '失败';
  }
  return '已完成';
};

const normalizeTaskStatus = (status?: string): TaskStatus => {
  if (status === 'queued' || status === 'generating' || status === 'failed') {
    return status;
  }
  return 'completed';
};

const advanceProgress = (progress: number) => {
  if (progress >= 92) {
    return progress;
  }
  const step = progress < 58 ? 7 : progress < 78 ? 4 : 2;
  return Math.min(92, progress + step);
};

const getTaskAspectRatio = (task: Pick<ImageTask, 'ratio' | 'size'>) => {
  const sizeMatch = task.size?.match(/^(\d+)x(\d+)$/);
  if (sizeMatch) {
    return `${sizeMatch[1]} / ${sizeMatch[2]}`;
  }
  const ratioMatch = task.ratio?.match(/^(\d+):(\d+)$/);
  if (ratioMatch) {
    return `${ratioMatch[1]} / ${ratioMatch[2]}`;
  }
  return '1 / 1';
};

const getTaskAspectRatioParts = (task: Pick<ImageTask, 'ratio' | 'size'>) => {
  const sizeMatch = task.size?.match(/^(\d+)x(\d+)$/);
  if (sizeMatch) {
    return {
      width: Number(sizeMatch[1]) || 1,
      height: Number(sizeMatch[2]) || 1,
    };
  }
  const ratioMatch = task.ratio?.match(/^(\d+):(\d+)$/);
  if (ratioMatch) {
    return {
      width: Number(ratioMatch[1]) || 1,
      height: Number(ratioMatch[2]) || 1,
    };
  }
  return { width: 1, height: 1 };
};

const hasImages = (urls?: string[]) => Boolean(urls?.filter(Boolean).length);

const getTaskRetryKey = (
  task: Pick<
    ImageTask,
    'prompt' | 'siteModelID' | 'size' | 'ratio' | 'style' | 'quality' | 'count'
  >,
) =>
  [
    task.prompt,
    task.siteModelID,
    task.size || task.ratio,
    task.style,
    task.quality,
    task.count,
  ].join('|');

const removeRetrySupersededFailedTasks = (tasks: ImageTask[]) => {
  const successfulRetryKeys = new Set(
    tasks
      .filter((task) => task.status !== 'failed')
      .map((task) => getTaskRetryKey(task)),
  );
  return tasks.filter(
    (task) =>
      task.status !== 'failed' ||
      !successfulRetryKeys.has(getTaskRetryKey(task)),
  );
};

const mapGenerationToTask = (item: AiImageGeneration): ImageTask => {
  const imageURLs = item.image_urls || [];
  const taskStatus = normalizeTaskStatus(item.status);
  return {
    id: item.generation_id,
    prompt: item.prompt,
    negativePrompt: item.negative_prompt || '',
    model: item.site_model_id,
    siteModelID: item.site_model_id,
    ratio: item.aspect_ratio || getSizeOption(item.size).ratio,
    size: item.size || item.aspect_ratio,
    style: item.style || '写实',
    quality: item.quality || 'auto',
    count: item.count || imageURLs.length || 1,
    status: taskStatus,
    createdAt: item.created_at * 1000,
    progress: taskStatus === 'completed' || taskStatus === 'failed' ? 100 : 36,
    imageURLs,
    referenceImages: [],
  };
};

const resolveImageURL = (url: string) => {
  if (!url || url.startsWith('data:')) {
    return url;
  }
  try {
    const parsed = new URL(url, window.location.origin);
    const match = parsed.pathname.match(
      /^\/uploads\/ai-images\/([^/]+)\/([^/?#]+)$/,
    );
    if (match) {
      return `/answer/api/v1/ai-image/assets/${encodeURIComponent(
        match[1],
      )}/${encodeURIComponent(match[2])}`;
    }
  } catch {
    // Keep the original URL if the browser cannot parse it.
  }
  if (/^https?:\/\//i.test(url)) {
    return url;
  }
  const apiBase = process.env.REACT_APP_API_URL || '';
  if (apiBase && apiBase !== '/' && url.startsWith('/')) {
    return `${apiBase.replace(/\/$/, '')}${url}`;
  }
  return url;
};

const getDownloadFilename = (task?: ImageTask, imageIndex = 0) => {
  if (!task) {
    return 'ai-image.png';
  }
  return `${task.id || 'ai-image'}-${imageIndex + 1}.png`;
};

const getImagePreviewStyle = (task: Pick<ImageTask, 'ratio' | 'size'>) => {
  const ratioParts = getTaskAspectRatioParts(task);
  return {
    '--hcai-preview-ratio': getTaskAspectRatio(task),
    '--hcai-preview-ratio-width': ratioParts.width,
    '--hcai-preview-ratio-height': ratioParts.height,
  } as CSSProperties;
};

const ImageGenerationWorkspace: FC<IProps> = ({
  subscription,
  onRefreshSubscription,
  onOpenSubscription,
}) => {
  const fileInputRef = useRef<HTMLInputElement | null>(null);
  const [models, setModels] = useState<AiImageModel[]>([]);
  const [modelsLoading, setModelsLoading] = useState(false);
  const [selectedModelID, setSelectedModelID] = useState('');
  const [prompt, setPrompt] = useState('');
  const [negativePrompt, setNegativePrompt] = useState('');
  const [size, setSize] = useState('1024x1024');
  const [style, setStyle] = useState('写实');
  const [count, setCount] = useState(1);
  const [quality, setQuality] = useState('auto');
  const [autoInsert, setAutoInsert] = useState(true);
  const [referenceImages, setReferenceImages] = useState<ReferenceImage[]>([]);
  const [tasks, setTasks] = useState<ImageTask[]>([]);
  const [activeTaskID, setActiveTaskID] = useState('');
  const [error, setError] = useState('');
  const [actionNotice, setActionNotice] = useState('');
  const [generating, setGenerating] = useState(false);
  const [editTarget, setEditTarget] = useState<EditTarget | null>(null);
  const [editPrompt, setEditPrompt] = useState('');
  const [editLoading, setEditLoading] = useState(false);
  const [previewImage, setPreviewImage] = useState<PreviewImage | null>(null);
  const [mobileTasksOpen, setMobileTasksOpen] = useState(false);
  const [mobileTaskPortalHost, setMobileTaskPortalHost] =
    useState<Element | null>(null);
  const [sidebarTaskPortalHost, setSidebarTaskPortalHost] =
    useState<Element | null>(null);
  const [imageBlobURLs, setImageBlobURLs] = useState<Record<string, string>>(
    {},
  );
  const imageBlobURLsRef = useRef<Record<string, string>>({});
  const loadingImageURLsRef = useRef<Set<string>>(new Set());

  const selectedModel = useMemo(
    () => models.find((model) => model.site_model_id === selectedModelID),
    [models, selectedModelID],
  );
  const selectedSize = getSizeOption(size);
  const activeTask = tasks.find((task) => task.id === activeTaskID) || tasks[0];
  const quotaRemaining = subscription?.image_quota_remaining ?? 0;
  const canGenerate = Boolean(prompt.trim()) && Boolean(selectedModelID);
  const activeImageURL = activeTask?.imageURLs[0]
    ? resolveImageURL(activeTask.imageURLs[0])
    : '';
  const imageURLKey = useMemo(
    () => tasks.flatMap((task) => task.imageURLs).join('|'),
    [tasks],
  );

  const refreshImageGenerations = useCallback(async () => {
    const resp = await getAiImageGenerations();
    const rawHistoryTasks = (resp || []).map(mapGenerationToTask);
    const historyTasks = removeRetrySupersededFailedTasks(rawHistoryTasks);
    setTasks((prev) => {
      const historyTaskIDs = new Set(historyTasks.map((task) => task.id));
      const localPendingTasks = prev.filter(
        (task) =>
          (task.status === 'queued' || task.status === 'generating') &&
          !historyTaskIDs.has(task.id),
      );
      return removeRetrySupersededFailedTasks([
        ...localPendingTasks,
        ...historyTasks,
      ]);
    });
    setActiveTaskID((prev) => {
      if (prev && historyTasks.some((task) => task.id === prev)) {
        return prev;
      }
      return historyTasks[0]?.id || '';
    });
  }, []);

  const showActionNotice = (message: string) => {
    setActionNotice(message);
    window.setTimeout(() => {
      setActionNotice('');
    }, 1800);
  };

  useEffect(() => {
    let mounted = true;
    setModelsLoading(true);
    getAiImageModels()
      .then((resp) => {
        if (!mounted) {
          return;
        }
        const nextModels = resp || [];
        setModels(nextModels);
        if (nextModels[0]) {
          setSelectedModelID(nextModels[0].site_model_id);
        }
      })
      .catch(() => {
        if (mounted) {
          setError('生图模型加载失败，请联系管理员检查配置');
        }
      })
      .finally(() => {
        if (mounted) {
          setModelsLoading(false);
        }
      });

    getAiImageGenerations()
      .then((resp) => {
        if (!mounted || !resp?.length) {
          return;
        }
        const historyTasks = resp.map(mapGenerationToTask);
        setTasks(historyTasks);
        setActiveTaskID(historyTasks[0].id);
      })
      .catch(() => undefined);

    return () => {
      mounted = false;
    };
  }, []);

  useEffect(() => {
    if (
      !tasks.some(
        (task) => task.status === 'queued' || task.status === 'generating',
      )
    ) {
      return undefined;
    }

    const timer = window.setInterval(() => {
      refreshImageGenerations()
        .then(() => {
          onRefreshSubscription();
        })
        .catch(() => undefined);
    }, 4000);
    return () => window.clearInterval(timer);
  }, [onRefreshSubscription, refreshImageGenerations, tasks]);

  useEffect(() => {
    if (
      !tasks.some(
        (task) => task.status === 'queued' || task.status === 'generating',
      )
    ) {
      return undefined;
    }

    const timer = window.setInterval(() => {
      setTasks((prev) =>
        prev.map((task) =>
          task.status === 'queued' || task.status === 'generating'
            ? { ...task, progress: advanceProgress(task.progress) }
            : task,
        ),
      );
    }, 900);
    return () => window.clearInterval(timer);
  }, [tasks]);

  useEffect(() => {
    imageBlobURLsRef.current = imageBlobURLs;
  }, [imageBlobURLs]);

  useEffect(() => {
    if (!previewImage) {
      return undefined;
    }
    const handleKeyDown = (evt: KeyboardEvent) => {
      if (evt.key === 'Escape') {
        setPreviewImage(null);
      }
    };
    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [previewImage]);

  useEffect(() => {
    return () => {
      Array.from(new Set(Object.values(imageBlobURLsRef.current))).forEach(
        (objectURL) => {
          URL.revokeObjectURL(objectURL);
        },
      );
    };
  }, []);

  useEffect(() => {
    const handleToggleImageTasks = (evt: Event) => {
      const open = (evt as CustomEvent<{ open?: boolean }>).detail?.open;
      setMobileTasksOpen((prev) => (typeof open === 'boolean' ? open : !prev));
    };
    window.addEventListener('hcai-toggle-image-tasks', handleToggleImageTasks);
    return () => {
      window.removeEventListener(
        'hcai-toggle-image-tasks',
        handleToggleImageTasks,
      );
      window.dispatchEvent(
        new CustomEvent('hcai-image-tasks-open-change', {
          detail: { open: false },
        }),
      );
    };
  }, []);

  useEffect(() => {
    if (!mobileTasksOpen) {
      setMobileTaskPortalHost(null);
      return;
    }
    setMobileTaskPortalHost(
      document.querySelector('.hcai-mobile-conversation-menu'),
    );
  }, [mobileTasksOpen]);

  useEffect(() => {
    setSidebarTaskPortalHost(
      document.querySelector('#hcai-sidebar-image-tasks'),
    );
  }, []);

  useEffect(() => {
    const token = Storage.get(LOGGED_TOKEN_STORAGE_KEY) || '';
    const imageURLs = Array.from(
      new Set(tasks.flatMap((task) => task.imageURLs)),
    ).filter(Boolean);

    imageURLs.forEach((rawURL) => {
      if (
        rawURL.startsWith('data:') ||
        imageBlobURLsRef.current[rawURL] ||
        loadingImageURLsRef.current.has(rawURL)
      ) {
        return;
      }
      const requestURL = resolveImageURL(rawURL);
      if (!requestURL || requestURL.startsWith('data:')) {
        return;
      }
      const headers: Record<string, string> = {};
      if (token) {
        headers.Authorization = token;
      }
      loadingImageURLsRef.current.add(rawURL);
      fetch(requestURL, { credentials: 'include', headers })
        .then((resp) => {
          if (!resp.ok) {
            throw new Error(`Image load failed: ${resp.status}`);
          }
          return resp.blob();
        })
        .then((blob) => {
          const objectURL = URL.createObjectURL(blob);
          setImageBlobURLs((prev) => {
            if (prev[rawURL]) {
              URL.revokeObjectURL(objectURL);
              return prev;
            }
            const next = { ...prev, [rawURL]: objectURL };
            imageBlobURLsRef.current = next;
            return next;
          });
        })
        .catch(() => undefined)
        .finally(() => {
          loadingImageURLsRef.current.delete(rawURL);
        });
    });
  }, [imageURLKey, tasks]);

  const getDisplayImageURL = (imageURL: string) => {
    if (!imageURL || imageURL.startsWith('data:')) {
      return imageURL;
    }
    return imageBlobURLs[imageURL] || '';
  };

  const addReferenceImages = async (files: File[]) => {
    const imageFiles = files.filter((file) => file.type.startsWith('image/'));
    if (imageFiles.length === 0) {
      return;
    }
    if (referenceImages.length + imageFiles.length > 4) {
      setError('最多添加 4 张参考图');
      return;
    }
    const oversized = imageFiles.find((file) => file.size > 5 * 1024 * 1024);
    if (oversized) {
      setError('单张参考图不能超过 5MB');
      return;
    }
    const images = await Promise.all(
      imageFiles.map(async (file) => ({
        id: `${file.name}-${file.lastModified}-${Math.random()
          .toString(36)
          .slice(2)}`,
        name: file.name,
        url: await readFileAsDataURL(file),
      })),
    );
    setReferenceImages((prev) => [...prev, ...images]);
    setError('');
  };

  const handleReferenceSelect = (evt: ChangeEvent<HTMLInputElement>) => {
    const files = Array.from(evt.target.files || []);
    evt.target.value = '';
    addReferenceImages(files).catch(() => {
      setError('参考图读取失败，请重新选择');
    });
  };

  const removeReferenceImage = (id: string) => {
    setReferenceImages((prev) => prev.filter((image) => image.id !== id));
  };

  const downloadImage = async (
    imageURL: string,
    task?: ImageTask,
    imageIndex = 0,
  ) => {
    if (!imageURL) {
      showActionNotice('暂无可下载图片');
      return;
    }
    try {
      const token = Storage.get(LOGGED_TOKEN_STORAGE_KEY) || '';
      const headers: Record<string, string> = {};
      if (token) {
        headers.Authorization = token;
      }
      const resp = await fetch(imageURL, { credentials: 'include', headers });
      if (!resp.ok) {
        throw new Error(`Download failed: ${resp.status}`);
      }
      const blob = await resp.blob();
      const objectURL = URL.createObjectURL(blob);
      const link = document.createElement('a');
      link.href = objectURL;
      link.download = getDownloadFilename(task, imageIndex);
      document.body.appendChild(link);
      link.click();
      link.remove();
      window.setTimeout(() => URL.revokeObjectURL(objectURL), 0);
      showActionNotice('已开始下载');
    } catch {
      showActionNotice('下载失败，请稍后重试');
    }
  };

  const downloadActiveImage = () => {
    downloadImage(activeImageURL, activeTask, 0);
  };

  const openImagePreview = (
    task: ImageTask,
    imageIndex: number,
    src: string,
    rawURL: string,
  ) => {
    setPreviewImage({ task, imageIndex, src, rawURL });
  };

  const runImageTask = async (task: ImageTask) => {
    if (quotaRemaining !== -1 && quotaRemaining < task.count) {
      setError('剩余生图额度不足，请升级或兑换订阅');
      return;
    }

    setTasks((prev) =>
      prev.map((item) =>
        item.id === task.id
          ? { ...item, status: 'generating', progress: 36, imageURLs: [] }
          : item,
      ),
    );
    setActiveTaskID(task.id);
    setError('');
    setGenerating(true);
    try {
      const resp = await generateAiImage({
        prompt: task.prompt,
        negative_prompt: task.negativePrompt,
        model: task.siteModelID,
        aspect_ratio: task.ratio,
        size: task.size,
        style: task.style,
        quality: task.quality,
        count: task.count,
        reference_images: task.referenceImages,
      });
      const nextImageURLs = resp.image_urls || [];
      const nextTaskID = resp.generation_id || task.id;
      setTasks((prev) =>
        removeRetrySupersededFailedTasks(
          prev
            .filter((item) => item.id === task.id || item.status !== 'failed')
            .map((item) =>
              item.id === task.id
                ? {
                    ...item,
                    id: nextTaskID,
                    size: resp.size || item.size,
                    status: hasImages(nextImageURLs)
                      ? 'completed'
                      : 'generating',
                    progress: hasImages(nextImageURLs) ? 100 : 92,
                    imageURLs: nextImageURLs,
                  }
                : item,
            ),
        ),
      );
      setActiveTaskID(nextTaskID);
      if (!hasImages(nextImageURLs)) {
        refreshImageGenerations().catch(() => undefined);
      }
      onRefreshSubscription();
    } catch (err: any) {
      setTasks((prev) =>
        prev.map((item) =>
          item.id === task.id
            ? { ...item, status: 'failed', progress: 100 }
            : item,
        ),
      );
      setError(err?.msg || '生成失败，请检查模型配置或稍后重试');
    } finally {
      setGenerating(false);
    }
  };

  const retryTask = (task: ImageTask) => {
    if (generating || task.status !== 'failed') {
      return;
    }
    runImageTask(task);
  };

  const closeMobileTasks = () => {
    setMobileTasksOpen(false);
    window.dispatchEvent(
      new CustomEvent('hcai-image-tasks-open-change', {
        detail: { open: false },
      }),
    );
  };

  const selectTask = (taskID: string, closePanel = false) => {
    setActiveTaskID(taskID);
    if (closePanel) {
      closeMobileTasks();
    }
  };

  const renderTaskItem = (
    task: ImageTask,
    closePanel = false,
    showRetry = true,
  ) => (
    <div
      className={classNames('hcai-task-item', {
        active: task.id === activeTaskID,
        'has-retry': showRetry && task.status === 'failed',
      })}
      key={task.id}>
      <button
        type="button"
        className="hcai-task-select"
        onClick={() => selectTask(task.id, closePanel)}>
        <span className={`hcai-task-dot ${task.status}`} />
        <div className="hcai-task-body">
          <strong>{task.prompt}</strong>
          <span>
            {task.model} · {task.size || task.ratio} · {task.count} 张
          </span>
        </div>
        <em>{getTaskStatusLabel(task.status)}</em>
      </button>
      {showRetry && task.status === 'failed' ? (
        <button
          type="button"
          className="hcai-task-retry"
          disabled={generating}
          onClick={() => retryTask(task)}>
          <Icon name="arrow-clockwise" />
          <span>重试</span>
        </button>
      ) : null}
    </div>
  );

  const renderTaskPanel = (closePanel = false, showRetry = true) => (
    <div className="hcai-task-panel">
      <div className="hcai-task-head">
        <span>任务队列</span>
        <strong>{tasks.length}</strong>
      </div>
      <div className="hcai-task-list">
        {tasks.length > 0 ? (
          tasks.map((task) => renderTaskItem(task, closePanel, showRetry))
        ) : (
          <span className="hcai-task-empty">暂无任务</span>
        )}
      </div>
    </div>
  );

  const openImageEdit = (task: ImageTask, imageIndex: number) => {
    const imageURL = task.imageURLs[imageIndex];
    if (!imageURL) {
      showActionNotice('暂无可编辑图片');
      return;
    }
    setEditTarget({ taskID: task.id, imageIndex, imageURL });
    setEditPrompt('');
    setError('');
  };

  const cancelImageEdit = () => {
    if (editLoading) {
      return;
    }
    setEditTarget(null);
    setEditPrompt('');
  };

  const submitImageEdit = async (evt: FormEvent) => {
    evt.preventDefault();
    if (!editTarget || !editPrompt.trim()) {
      setError('请输入图片编辑提示词');
      return;
    }
    const task = tasks.find((item) => item.id === editTarget.taskID);
    if (!task) {
      setError('原图片任务不存在，请刷新后重试');
      return;
    }
    const editTaskID = uuidv4();
    const nextTask: ImageTask = {
      id: editTaskID,
      prompt: editPrompt.trim(),
      negativePrompt: task.negativePrompt,
      model: task.model,
      siteModelID: task.siteModelID,
      ratio: task.ratio,
      size: task.size,
      style: task.style,
      quality: task.quality,
      count: 1,
      status: 'generating',
      createdAt: Date.now(),
      progress: 12,
      imageURLs: [],
      referenceImages: [editTarget.imageURL],
    };
    const sourceImageURL = editTarget.imageURL;
    const editPromptValue = editPrompt.trim();
    setEditLoading(true);
    setError('');
    setTasks((prev) => [nextTask, ...prev]);
    setActiveTaskID(editTaskID);
    setEditTarget(null);
    setEditPrompt('');
    try {
      const resp = await editAiImage({
        prompt: editPromptValue,
        image_url: sourceImageURL,
        model: task.siteModelID,
        size: task.size,
        quality: task.quality,
      });
      const nextURLs = resp.image_urls || [];
      const nextTaskID = resp.generation_id || editTaskID;
      setTasks((prev) =>
        prev.map((item) =>
          item.id === editTaskID
            ? {
                ...item,
                id: nextTaskID,
                size: resp.size || item.size,
                status: hasImages(nextURLs) ? 'completed' : 'generating',
                progress: hasImages(nextURLs) ? 100 : 92,
                imageURLs: nextURLs,
                count: Math.max(1, nextURLs.length),
              }
            : item,
        ),
      );
      setActiveTaskID(nextTaskID);
      if (hasImages(nextURLs)) {
        showActionNotice('编辑完成');
      } else {
        refreshImageGenerations().catch(() => undefined);
      }
      onRefreshSubscription();
    } catch (err: any) {
      setTasks((prev) =>
        prev.map((item) =>
          item.id === editTaskID
            ? { ...item, status: 'failed', progress: 100 }
            : item,
        ),
      );
      setError(err?.msg || '图片编辑失败，请稍后重试');
    } finally {
      setEditLoading(false);
    }
  };

  const submitGeneration = async (evt: FormEvent) => {
    evt.preventDefault();
    if (!canGenerate) {
      setError(!selectedModelID ? '请先选择生图模型' : '请输入生图提示词');
      return;
    }
    if (quotaRemaining !== -1 && quotaRemaining < count) {
      setError('剩余生图额度不足，请升级或兑换订阅');
      return;
    }

    const taskID = uuidv4();
    const nextTask: ImageTask = {
      id: taskID,
      prompt: prompt.trim(),
      negativePrompt: negativePrompt.trim(),
      model: getModelName(selectedModel),
      siteModelID: selectedModelID,
      ratio: selectedSize.ratio,
      size: selectedSize.value,
      style,
      quality,
      count,
      status: 'generating',
      createdAt: Date.now(),
      progress: 36,
      imageURLs: [],
      referenceImages: referenceImages.map((image) => image.url),
    };
    setTasks((prev) => [nextTask, ...prev]);
    setActiveTaskID(taskID);
    runImageTask(nextTask).then(() => {
      setPrompt('');
    });
  };

  return (
    <div className="hcai-image-workspace">
      {sidebarTaskPortalHost
        ? createPortal(renderTaskPanel(false, false), sidebarTaskPortalHost)
        : null}
      {mobileTasksOpen && mobileTaskPortalHost
        ? createPortal(
            <div
              id="hcai-mobile-image-tasks"
              className="hcai-mobile-conversation-panel hcai-mobile-task-panel">
              {renderTaskPanel(true)}
            </div>,
            mobileTaskPortalHost,
          )
        : null}
      <section className="hcai-image-composer">
        <div className="hcai-image-head">
          <div>
            <span className="hcai-image-kicker">HCAI Image</span>
            <h1>图片生成</h1>
          </div>
          <button type="button" onClick={onOpenSubscription}>
            <Icon name="credit-card-2-front" />
            <span>剩余额度 {formatQuota(quotaRemaining)}</span>
          </button>
        </div>

        <form className="hcai-image-form" onSubmit={submitGeneration}>
          <div className="hcai-image-field">
            <label htmlFor="hcai-image-prompt">提示词</label>
            <textarea
              id="hcai-image-prompt"
              value={prompt}
              rows={6}
              placeholder="描述你想生成的画面、主体、构图、光线和风格"
              onChange={(evt) => setPrompt(evt.target.value)}
            />
          </div>

          <div className="hcai-image-field compact">
            <label htmlFor="hcai-negative-prompt">不希望出现</label>
            <input
              id="hcai-negative-prompt"
              value={negativePrompt}
              placeholder="低清晰度、变形、文字错误..."
              onChange={(evt) => setNegativePrompt(evt.target.value)}
            />
          </div>

          <div className="hcai-image-panel">
            <div className="hcai-image-panel-title">
              <Icon name="image" />
              <span>参考图</span>
            </div>
            <input
              ref={fileInputRef}
              type="file"
              accept="image/*"
              multiple
              className="hcai-image-input"
              onChange={handleReferenceSelect}
            />
            <div className="hcai-reference-grid">
              {referenceImages.map((image) => (
                <div className="hcai-reference-image" key={image.id}>
                  <img src={image.url} alt={image.name} />
                  <button
                    type="button"
                    aria-label="移除参考图"
                    onClick={() => removeReferenceImage(image.id)}>
                    <Icon name="x" />
                  </button>
                </div>
              ))}
              <button
                type="button"
                className="hcai-reference-add"
                onClick={() => fileInputRef.current?.click()}>
                <Icon name="plus-lg" />
                <span>添加</span>
              </button>
            </div>
          </div>

          <div className="hcai-image-grid-row">
            <div className="hcai-image-field compact">
              <label htmlFor="hcai-image-model">模型</label>
              <select
                id="hcai-image-model"
                value={selectedModelID}
                disabled={modelsLoading || models.length === 0}
                onChange={(evt) => setSelectedModelID(evt.target.value)}>
                {modelsLoading ? <option>加载模型...</option> : null}
                {models.map((model) => (
                  <option value={model.site_model_id} key={model.site_model_id}>
                    {getModelName(model)}
                  </option>
                ))}
              </select>
            </div>
            <div className="hcai-image-field compact">
              <label htmlFor="hcai-image-quality">质量</label>
              <select
                id="hcai-image-quality"
                value={quality}
                onChange={(evt) => setQuality(evt.target.value)}>
                {qualityOptions.map((option) => (
                  <option value={option.value} key={option.value}>
                    {option.label}
                  </option>
                ))}
              </select>
            </div>
          </div>

          <div className="hcai-image-section-label">尺寸</div>
          <div className="hcai-ratio-options">
            {sizeOptions.map((option) => (
              <button
                type="button"
                className={size === option.value ? 'active' : ''}
                key={option.value}
                onClick={() => setSize(option.value)}>
                <strong>{option.label}</strong>
                <span>{option.meta}</span>
              </button>
            ))}
          </div>

          <div className="hcai-image-section-label">风格</div>
          <div className="hcai-style-options">
            {styleOptions.map((option) => (
              <button
                type="button"
                className={style === option ? 'active' : ''}
                key={option}
                onClick={() => setStyle(option)}>
                {option}
              </button>
            ))}
          </div>

          <div className="hcai-image-action-row">
            <div className="hcai-count-control" aria-label="生成数量">
              {countOptions.map((value) => (
                <button
                  type="button"
                  className={count === value ? 'active' : ''}
                  key={value}
                  onClick={() => setCount(value)}>
                  {value}
                </button>
              ))}
            </div>
            <label className="hcai-auto-insert">
              <input
                type="checkbox"
                checked={autoInsert}
                onChange={(evt) => setAutoInsert(evt.target.checked)}
              />
              <span>生成后自动加入画布</span>
            </label>
          </div>

          {error ? <div className="hcai-image-error">{error}</div> : null}

          <button
            type="submit"
            className="hcai-image-generate"
            disabled={!canGenerate || generating}>
            <Icon name="stars" />
            <span>{generating ? '生成中' : '生成图片'}</span>
          </button>
        </form>
      </section>

      <section className="hcai-image-preview">
        <div className="hcai-preview-toolbar">
          <div>
            <span>预览</span>
            <strong>
              {activeTask ? getTaskStatusLabel(activeTask.status) : '待生成'}
            </strong>
          </div>
          <div className="hcai-preview-actions">
            {actionNotice ? (
              <span className="hcai-preview-notice">{actionNotice}</span>
            ) : null}
            <button
              type="button"
              title="下载"
              disabled={!activeImageURL}
              onClick={downloadActiveImage}>
              <Icon name="download" />
            </button>
          </div>
        </div>

        <div className="hcai-preview-canvas">
          {activeTask ? (
            <div
              className={classNames('hcai-preview-result', {
                loading: activeTask.status === 'generating',
              })}>
              {Array.from({ length: activeTask.count }, (_, tileNumber) => {
                const rawImageURL = activeTask.imageURLs[tileNumber];
                const displayImageURL = rawImageURL
                  ? getDisplayImageURL(rawImageURL)
                  : '';
                return (
                  <div
                    className={classNames('hcai-preview-tile', {
                      'has-image': Boolean(rawImageURL),
                    })}
                    style={getImagePreviewStyle(activeTask)}
                    key={`${activeTask.id}-${tileNumber + 1}`}>
                    {activeTask.status === 'generating' ? (
                      <div className="hcai-preview-progress">
                        <span style={{ width: `${activeTask.progress}%` }} />
                      </div>
                    ) : rawImageURL ? (
                      displayImageURL ? (
                        <>
                          <button
                            type="button"
                            className="hcai-preview-image-button"
                            onClick={() =>
                              openImagePreview(
                                activeTask,
                                tileNumber,
                                displayImageURL,
                                rawImageURL,
                              )
                            }>
                            <img
                              className="hcai-preview-image"
                              src={displayImageURL}
                              alt={`${activeTask.prompt}-${tileNumber + 1}`}
                            />
                          </button>
                          <div className="hcai-preview-tile-meta">
                            <div>
                              <strong>{activeTask.prompt}</strong>
                              <span>
                                尺寸 {activeTask.size || activeTask.ratio}
                              </span>
                            </div>
                            <button
                              type="button"
                              title="编辑这张图"
                              aria-label="编辑这张图"
                              disabled={editLoading}
                              onClick={() =>
                                openImageEdit(activeTask, tileNumber)
                              }>
                              <Icon name="pencil-square" />
                            </button>
                            <button
                              type="button"
                              title="下载这张图"
                              aria-label="下载这张图"
                              onClick={() => {
                                downloadImage(
                                  resolveImageURL(rawImageURL),
                                  activeTask,
                                  tileNumber,
                                );
                              }}>
                              <Icon name="download" />
                            </button>
                          </div>
                          {editTarget?.taskID === activeTask.id &&
                          editTarget.imageIndex === tileNumber ? (
                            <form
                              className="hcai-preview-edit-popover"
                              onSubmit={submitImageEdit}>
                              <input
                                value={editPrompt}
                                disabled={editLoading}
                                placeholder="输入编辑提示词"
                                onChange={(evt) =>
                                  setEditPrompt(evt.target.value)
                                }
                              />
                              <div>
                                <button
                                  type="submit"
                                  disabled={editLoading || !editPrompt.trim()}>
                                  {editLoading ? '编辑中' : '编辑'}
                                </button>
                                <button type="button" onClick={cancelImageEdit}>
                                  取消
                                </button>
                              </div>
                            </form>
                          ) : null}
                        </>
                      ) : (
                        <>
                          <Icon name="image" />
                          <span>图片加载中</span>
                        </>
                      )
                    ) : (
                      <>
                        <Icon name="image" />
                        <span>{activeTask.ratio}</span>
                      </>
                    )}
                  </div>
                );
              })}
            </div>
          ) : (
            <div className="hcai-preview-empty">
              <Icon name="image" />
              <span>生成结果会显示在这里</span>
            </div>
          )}
        </div>

        {renderTaskPanel()}
      </section>
      {previewImage ? (
        <div className="hcai-image-lightbox" role="dialog" aria-modal="true">
          <button
            type="button"
            className="hcai-image-lightbox-backdrop"
            aria-label="关闭预览"
            onClick={() => setPreviewImage(null)}
          />
          <div className="hcai-image-lightbox-bar">
            <div>
              <strong>{previewImage.task.prompt}</strong>
              <span>
                尺寸 {previewImage.task.size || previewImage.task.ratio}
              </span>
            </div>
            <button
              type="button"
              title="下载"
              aria-label="下载"
              onClick={(evt) => {
                evt.stopPropagation();
                downloadImage(
                  resolveImageURL(previewImage.rawURL),
                  previewImage.task,
                  previewImage.imageIndex,
                );
              }}>
              <Icon name="download" />
            </button>
            <button
              type="button"
              title="关闭"
              aria-label="关闭"
              onClick={(evt) => {
                evt.stopPropagation();
                setPreviewImage(null);
              }}>
              <Icon name="x-lg" />
            </button>
          </div>
          <img src={previewImage.src} alt={previewImage.task.prompt} />
        </div>
      ) : null}
    </div>
  );
};

export default ImageGenerationWorkspace;
