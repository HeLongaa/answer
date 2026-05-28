import { FC, useState } from 'react';
import { Button, Form } from 'react-bootstrap';
import { useNavigate } from 'react-router-dom';

import { createTask } from '@/services';
import { toastStore } from '@/stores';
import { usePageTags } from '@/hooks';

import '../index.scss';

const getErrorMessage = (err: any, fallback: string) => {
  return err?.msg || err?.message || fallback;
};

const NewTask: FC = () => {
  const navigate = useNavigate();
  const [title, setTitle] = useState('');
  const [description, setDescription] = useState('');
  const [loading, setLoading] = useState(false);

  usePageTags({ title: '提出需求' });

  const handleSubmit = async (evt) => {
    evt.preventDefault();
    setLoading(true);
    try {
      await createTask({ title, description });
      toastStore
        .getState()
        .show({ msg: '需求已提交，等待审核', variant: 'success' });
      navigate('/tasks?mine=1');
    } catch (err: any) {
      toastStore.getState().show({
        msg: getErrorMessage(err, '提交失败，请稍后重试'),
        variant: 'danger',
      });
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="task-square-page">
      <div className="task-square-head compact">
        <div>
          <span className="task-square-kicker">New Request</span>
          <h1>提出需求</h1>
          <p>奖励积分、标签、截止时间和提交要求会由管理员或版主审核时补充。</p>
        </div>
      </div>
      <Form className="task-square-form" onSubmit={handleSubmit}>
        <Form.Group className="mb-3">
          <Form.Label>标题</Form.Label>
          <Form.Control
            value={title}
            maxLength={150}
            onChange={(evt) => setTitle(evt.target.value)}
          />
        </Form.Group>
        <Form.Group className="mb-3">
          <Form.Label>描述</Form.Label>
          <Form.Control
            as="textarea"
            rows={8}
            value={description}
            onChange={(evt) => setDescription(evt.target.value)}
          />
        </Form.Group>
        <Button
          type="submit"
          disabled={loading || !title.trim() || !description.trim()}>
          提交审核
        </Button>
      </Form>
    </div>
  );
};

export default NewTask;
