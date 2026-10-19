import React, { useEffect, useState, useCallback } from 'react';
import { Layout, Button, message, Space, Typography } from 'antd';
import { PlusOutlined, ApiOutlined } from '@ant-design/icons';
import { api } from './services/api';
import {
  Expense,
  ExpenseFilters,
  CreateExpensePayload,
  UpdateExpensePayload,
} from './types/expense';
import { SummaryCards } from './components/SummaryCards';
import { FiltersBar } from './components/FiltersBar';
import { ExpensesTable } from './components/ExpensesTable';
import { ExpenseModal } from './components/ExpenseModal';

const { Header, Content } = Layout;
const { Title, Text } = Typography;

export const App: React.FC = () => {
  const [expenses, setExpenses] = useState<Expense[]>([]);
  const [distinctUsers, setDistinctUsers] = useState<number[]>([]);
  const [filters, setFilters] = useState<ExpenseFilters>({});
  const [loading, setLoading] = useState<boolean>(false);

  const [modalVisible, setModalVisible] = useState<boolean>(false);
  const [modalLoading, setModalLoading] = useState<boolean>(false);
  const [editingExpense, setEditingExpense] = useState<Expense | null>(null);

  const loadExpenses = useCallback(async (f: ExpenseFilters) => {
    setLoading(true);
    try {
      const data = await api.getExpenses(f);
      setExpenses(data);
    } catch {
      message.error('Ошибка загрузки списка расходов');
    } finally {
      setLoading(false);
    }
  }, []);

  const loadDistinctUsers = useCallback(async () => {
    try {
      const users = await api.getDistinctUsers();
      setDistinctUsers(users);
    } catch {
      // Non-critical, filter will stay empty
    }
  }, []);

  useEffect(() => {
    loadDistinctUsers();
  }, [loadDistinctUsers]);

  useEffect(() => {
    loadExpenses(filters);
  }, [filters, loadExpenses]);

  const handleFilterChange = (newFilters: ExpenseFilters) => {
    setFilters(newFilters);
  };

  const handleResetFilters = () => {
    setFilters({});
  };

  const handleOpenCreate = () => {
    setEditingExpense(null);
    setModalVisible(true);
  };

  const handleOpenEdit = (exp: Expense) => {
    setEditingExpense(exp);
    setModalVisible(true);
  };

  const handleModalSubmit = async (
    payload: CreateExpensePayload | UpdateExpensePayload
  ) => {
    setModalLoading(true);
    try {
      if (editingExpense) {
        await api.updateExpense(editingExpense.id, payload);
        message.success('Расход успешно обновлен');
      } else {
        await api.createExpense(payload as CreateExpensePayload);
        message.success('Расход успешно добавлен');
      }
      setModalVisible(false);
      await loadExpenses(filters);
      await loadDistinctUsers();
    } catch {
      message.error('Ошибка при сохранении расхода');
    } finally {
      setModalLoading(false);
    }
  };

  const handleDelete = async (id: string) => {
    try {
      await api.deleteExpense(id);
      message.success('Расход удален');
      await loadExpenses(filters);
      await loadDistinctUsers();
    } catch {
      message.error('Ошибка удаления расхода');
    }
  };

  return (
    <Layout style={{ minHeight: '100vh', backgroundColor: '#f0f2f5' }}>
      <Header
        style={{
          background: '#fff',
          padding: '0 32px',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
          borderBottom: '1px solid #f0f0f0',
          height: 72,
        }}
      >
        <div>
          <Title level={4} style={{ margin: 0, color: '#1f1f1f' }}>
            Семейный бюджет
          </Title>
          <Text type="secondary" style={{ fontSize: 13 }}>
            Панель управления расходами для главы семьи
          </Text>
        </div>

        <Space size="middle">
          <Button
            icon={<ApiOutlined />}
            href="/swagger/index.html"
            target="_blank"
          >
            Swagger API
          </Button>
          <Button
            type="primary"
            icon={<PlusOutlined />}
            onClick={handleOpenCreate}
            size="large"
            style={{ borderRadius: 6 }}
          >
            Добавить расход
          </Button>
        </Space>
      </Header>

      <Content style={{ padding: '24px 32px', maxWidth: 1400, margin: '0 auto', width: '100%' }}>
        <SummaryCards expenses={expenses} />

        <FiltersBar
          distinctUsers={distinctUsers}
          filters={filters}
          onChange={handleFilterChange}
          onReset={handleResetFilters}
        />

        <ExpensesTable
          expenses={expenses}
          loading={loading}
          onEdit={handleOpenEdit}
          onDelete={handleDelete}
        />

        <ExpenseModal
          visible={modalVisible}
          expense={editingExpense}
          onCancel={() => setModalVisible(false)}
          onSubmit={handleModalSubmit}
          loading={modalLoading}
        />
      </Content>
    </Layout>
  );
};
