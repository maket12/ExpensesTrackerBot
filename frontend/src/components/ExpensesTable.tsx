import React from 'react';
import { Table, Tag, Button, Space, Popconfirm, Tooltip } from 'antd';
import { EditOutlined, DeleteOutlined } from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';
import dayjs from 'dayjs';
import { Expense, ExpenseCategory, formatCategory } from '../types/expense';

interface ExpensesTableProps {
  expenses: Expense[];
  loading: boolean;
  onEdit: (expense: Expense) => void;
  onDelete: (id: string) => void;
}

const categoryColors: Record<ExpenseCategory, string> = {
  еда: 'orange',
  транспорт: 'blue',
  жильё: 'green',
  развлечения: 'purple',
  здоровье: 'magenta',
  покупки: 'cyan',
  другое: 'default',
};

export const ExpensesTable: React.FC<ExpensesTableProps> = ({
  expenses,
  loading,
  onEdit,
  onDelete,
}) => {
  const columns: ColumnsType<Expense> = [
    {
      title: 'Дата и время',
      dataIndex: 'spent_at',
      key: 'spent_at',
      sorter: (a, b) => dayjs(a.spent_at).valueOf() - dayjs(b.spent_at).valueOf(),
      defaultSortOrder: 'descend',
      render: (val: string) => dayjs(val).format('DD.MM.YYYY HH:mm'),
      width: 170,
    },
    {
      title: 'Член семьи',
      dataIndex: 'user_id',
      key: 'user_id',
      render: (uid: number) => <Tag color="geekblue">#{uid}</Tag>,
      width: 140,
    },
    {
      title: 'Категория',
      dataIndex: 'category',
      key: 'category',
      render: (cat: ExpenseCategory) => (
        <Tag color={categoryColors[cat] || 'default'}>
          {formatCategory(cat)}
        </Tag>
      ),
      width: 150,
    },
    {
      title: 'Сумма',
      dataIndex: 'amount',
      key: 'amount',
      sorter: (a, b) => a.amount - b.amount,
      render: (amt: number) => (
        <span style={{ fontWeight: 600, fontSize: 15 }}>
          {amt.toFixed(2)} ₽
        </span>
      ),
      width: 140,
    },
    {
      title: 'Комментарий',
      dataIndex: 'comment',
      key: 'comment',
      render: (val: string) => val || <span style={{ color: '#bfbfbf' }}>—</span>,
    },
    {
      title: 'Действия',
      key: 'actions',
      width: 110,
      render: (_, record) => (
        <Space size="small">
          <Tooltip title="Редактировать">
            <Button
              type="text"
              icon={<EditOutlined style={{ color: '#1890ff' }} />}
              onClick={() => onEdit(record)}
            />
          </Tooltip>
          <Tooltip title="Удалить">
            <Popconfirm
              title="Удалить запись?"
              description="Вы действительно хотите удалить этот расход?"
              onConfirm={() => onDelete(record.id)}
              okText="Да, удалить"
              cancelText="Отмена"
              okButtonProps={{ danger: true }}
            >
              <Button
                type="text"
                danger
                icon={<DeleteOutlined />}
              />
            </Popconfirm>
          </Tooltip>
        </Space>
      ),
    },
  ];

  return (
    <Table
      columns={columns}
      dataSource={expenses}
      rowKey="id"
      loading={loading}
      pagination={{
        defaultPageSize: 10,
        showSizeChanger: true,
        pageSizeOptions: ['10', '25', '50', '100'],
        showTotal: (total) => `Всего записей: ${total}`,
      }}
      style={{
        backgroundColor: '#fff',
        borderRadius: 8,
        boxShadow: '0 1px 3px rgba(0,0,0,0.08)',
        overflow: 'hidden',
      }}
    />
  );
};
