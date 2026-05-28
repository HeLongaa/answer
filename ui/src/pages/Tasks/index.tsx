import { FC, useState } from 'react';
import { Button, Badge, Form, Modal } from 'react-bootstrap';
import { Link, useSearchParams } from 'react-router-dom';

import { Empty, FormatTime, Pagination } from '@/components';
import { claimTask, submitTask, TaskItem, useTasks } from '@/services';
import { toastStore } from '@/stores';
import { usePageTags } from '@/hooks';

import './index.scss';

const PAGE_SIZE = 10;

const statusText = {
  open: '可领取',
  in_progress: '进行中',
  submitted: '待验收',
  completed: '已完成',
  failed: '已失败',
  closed: '已关闭',
  pending_review: '待审核',
  rejected: '已驳回',
};

const getErrorMessage = (err: any, fallback: string) => {
  return err?.msg || err?.message || fallback;
};

const Tasks: FC = () => {
  const [params] = useSearchParams();
  const page = Number(params.get('page')) || 1;
  const mine = params.get('mine') === '1';
  const { data, mutate } = useTasks({
    page,
    page_size: PAGE_SIZE,
    mine: mine ? 'true' : undefined,
  });
  const [submitTarget, setSubmitTarget] = useState<TaskItem | null>(null);
  const [submitContent, setSubmitContent] = useState('');
  const [submitLinks, setSubmitLinks] = useState('');

  usePageTags({ title: '任务广场' });

  const handleClaim = async (task: TaskItem) => {
    try {
      await claimTask(task.id);
      toastStore.getState().show({ msg: '领取成功', variant: 'success' });
      mutate();
    } catch (err: any) {
      toastStore.getState().show({
        msg: getErrorMessage(err, '领取失败，请稍后重试'),
        variant: 'danger',
      });
    }
  };

  const handleSubmit = async () => {
    if (!submitTarget) {
      return;
    }
    try {
      await submitTask({
        id: submitTarget.id,
        content: submitContent,
        links: submitLinks
          .split('\n')
          .map((item) => item.trim())
          .filter(Boolean),
      });
      toastStore
        .getState()
        .show({ msg: '提交成功，等待验收', variant: 'success' });
      setSubmitTarget(null);
      setSubmitContent('');
      setSubmitLinks('');
      mutate();
    } catch (err: any) {
      toastStore.getState().show({
        msg: getErrorMessage(err, '提交失败，请稍后重试'),
        variant: 'danger',
      });
    }
  };

  return (
    <div className="task-square-page">
      <div className="task-square-head">
        <div>
          <span className="task-square-kicker">Task Square</span>
          <h1>任务广场</h1>
          <p>提交需求，领取任务，验收通过后获得积分。</p>
        </div>
        <div className="task-square-actions">
          <Button
            as={Link as any}
            to="/users/settings/points"
            variant="outline-primary">
            我的积分
          </Button>
          <Button as={Link as any} to="/tasks/new">
            提出需求
          </Button>
        </div>
      </div>

      <div className="task-square-tabs">
        <Link className={!mine ? 'active' : ''} to="/tasks">
          全部任务
        </Link>
        <Link className={mine ? 'active' : ''} to="/tasks?mine=1">
          我的任务
        </Link>
      </div>

      <div className="task-square-list">
        {data?.list?.length === 0 ? <Empty /> : null}
        {data?.list?.map((task) => {
          const shouldShowReviewComment =
            (task.status === 'closed' || task.status === 'rejected') &&
            Boolean(task.review_comment);
          const requirementTitle = shouldShowReviewComment
            ? '审核说明'
            : '提交要求';
          const requirementContent = shouldShowReviewComment
            ? task.review_comment
            : task.submission_requirements;

          return (
            <article className="task-square-card" key={task.id}>
              <div className="task-square-card-main">
                <div className="task-square-card-title">
                  <h2>{task.title}</h2>
                  <Badge bg={task.status === 'open' ? 'success' : 'secondary'}>
                    {statusText[task.status] || task.status}
                  </Badge>
                </div>
                <div className="task-square-meta">
                  <span>发布人：{task.user_display_name || task.user_id}</span>
                  {task.assignee_id && task.assignee_id !== '0' ? (
                    <span>
                      领取人：{task.assignee_display_name || task.assignee_id}
                    </span>
                  ) : null}
                  <span>奖励：{task.reward_points} 积分</span>
                  {task.deadline ? (
                    <span>
                      截止：
                      <FormatTime time={task.deadline} />
                    </span>
                  ) : null}
                </div>
                {task.tags.length > 0 ? (
                  <div className="task-square-tags">
                    {task.tags.map((tag) => (
                      <span key={tag}>{tag}</span>
                    ))}
                  </div>
                ) : null}
                {requirementContent ? (
                  <div className="task-square-requirement">
                    <strong>{requirementTitle}</strong>
                    <p>{requirementContent}</p>
                  </div>
                ) : null}
              </div>
              <div className="task-square-card-actions">
                {task.status === 'open' ? (
                  <Button size="sm" onClick={() => handleClaim(task)}>
                    领取任务
                  </Button>
                ) : null}
                {task.status === 'in_progress' ? (
                  <Button
                    size="sm"
                    variant="outline-primary"
                    onClick={() => setSubmitTarget(task)}>
                    提交成果
                  </Button>
                ) : null}
              </div>
            </article>
          );
        })}
      </div>
      <Pagination
        currentPage={page}
        pageSize={PAGE_SIZE}
        totalSize={data?.count || 0}
      />

      <Modal
        show={Boolean(submitTarget)}
        onHide={() => setSubmitTarget(null)}
        centered>
        <Modal.Header closeButton>
          <Modal.Title>提交成果</Modal.Title>
        </Modal.Header>
        <Modal.Body>
          <Form.Group className="mb-3">
            <Form.Label>成果说明</Form.Label>
            <Form.Control
              as="textarea"
              rows={5}
              value={submitContent}
              onChange={(evt) => setSubmitContent(evt.target.value)}
            />
          </Form.Group>
          <Form.Group>
            <Form.Label>相关链接，每行一个</Form.Label>
            <Form.Control
              as="textarea"
              rows={3}
              value={submitLinks}
              onChange={(evt) => setSubmitLinks(evt.target.value)}
            />
          </Form.Group>
        </Modal.Body>
        <Modal.Footer>
          <Button variant="link" onClick={() => setSubmitTarget(null)}>
            取消
          </Button>
          <Button disabled={!submitContent.trim()} onClick={handleSubmit}>
            提交
          </Button>
        </Modal.Footer>
      </Modal>
    </div>
  );
};

export default Tasks;
