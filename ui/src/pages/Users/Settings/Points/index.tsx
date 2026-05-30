import { FC } from 'react';
import { Table } from 'react-bootstrap';
import { useSearchParams } from 'react-router-dom';

import { FormatTime, Icon, Pagination } from '@/components';
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
    <div className="card settings-content-card">
      <div className="card-body">
        <h3 className="mb-4">我的积分</h3>
        <div className="points-balance mb-4">
          <div className="points-balance-icon">
            <Icon name="coin" />
          </div>
          <div className="points-balance-label">当前积分</div>
          <div className="points-balance-value">{account?.balance || 0}</div>
        </div>
        <Table responsive hover className="points-table mb-0">
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
                <td
                  className={item.delta >= 0 ? 'text-success' : 'text-danger'}>
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
      </div>
    </div>
  );
};

export default Points;
