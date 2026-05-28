import { FC } from 'react';
import { Card, Table } from 'react-bootstrap';
import { useSearchParams } from 'react-router-dom';

import { FormatTime, Pagination } from '@/components';
import { usePointAccount, usePointTransactions } from '@/services';
import { usePageTags } from '@/hooks';

const PAGE_SIZE = 20;

const Points: FC = () => {
  const [params] = useSearchParams();
  const page = Number(params.get('page')) || 1;
  const { data: account } = usePointAccount();
  const { data } = usePointTransactions({ page, page_size: PAGE_SIZE });

  usePageTags({ title: '我的积分' });

  return (
    <>
      <h3 className="mb-4">我的积分</h3>
      <Card className="mb-3">
        <Card.Body>
          <div className="text-secondary">当前积分</div>
          <div className="display-6 fw-bold">{account?.balance || 0}</div>
        </Card.Body>
      </Card>
      <Table responsive>
        <thead>
          <tr>
            <th>时间</th>
            <th>说明</th>
            <th>变动</th>
            <th>余额</th>
          </tr>
        </thead>
        <tbody>
          {data?.list?.map((item) => (
            <tr key={item.id}>
              <td>
                <FormatTime time={item.created_at} />
              </td>
              <td>{item.description}</td>
              <td className={item.delta >= 0 ? 'text-success' : 'text-danger'}>
                {item.delta > 0 ? `+${item.delta}` : item.delta}
              </td>
              <td>{item.balance}</td>
            </tr>
          ))}
        </tbody>
      </Table>
      <Pagination
        currentPage={page}
        pageSize={PAGE_SIZE}
        totalSize={data?.count || 0}
      />
    </>
  );
};

export default Points;
