import React from 'react';
import { Card, Col, Row, Statistic } from 'antd';
import {
  DollarOutlined,
  ShoppingOutlined,
  FireOutlined,
} from '@ant-design/icons';
import { Expense } from '../types/expense';

interface SummaryCardsProps {
  expenses: Expense[];
}

export const SummaryCards: React.FC<SummaryCardsProps> = ({ expenses }) => {
  const totalAmount = expenses.reduce((acc, curr) => acc + curr.amount, 0);
  const count = expenses.length;

  const categoryTotals: Record<string, number> = {};
  for (const e of expenses) {
    categoryTotals[e.category] = (categoryTotals[e.category] || 0) + e.amount;
  }

  let topCategory = '—';
  let topCategoryAmount = 0;
  for (const [cat, amt] of Object.entries(categoryTotals)) {
    if (amt > topCategoryAmount) {
      topCategory = cat;
      topCategoryAmount = amt;
    }
  }

  return (
    <Row gutter={[16, 16]} style={{ marginBottom: 24 }}>
      <Col xs={24} sm={8}>
        <Card bordered={false} style={{ borderRadius: 8, boxShadow: '0 1px 3px rgba(0,0,0,0.08)' }}>
          <Statistic
            title="Всего потрачено"
            value={totalAmount}
            precision={2}
            suffix="₽"
            prefix={<DollarOutlined style={{ color: '#1890ff' }} />}
          />
        </Card>
      </Col>
      <Col xs={24} sm={8}>
        <Card bordered={false} style={{ borderRadius: 8, boxShadow: '0 1px 3px rgba(0,0,0,0.08)' }}>
          <Statistic
            title="Количество операций"
            value={count}
            prefix={<ShoppingOutlined style={{ color: '#52c41a' }} />}
          />
        </Card>
      </Col>
      <Col xs={24} sm={8}>
        <Card bordered={false} style={{ borderRadius: 8, boxShadow: '0 1px 3px rgba(0,0,0,0.08)' }}>
          <Statistic
            title="Главная статья расходов"
            value={topCategory}
            prefix={<FireOutlined style={{ color: '#fa8c16' }} />}
            suffix={topCategoryAmount > 0 ? `(${topCategoryAmount.toFixed(0)} ₽)` : ''}
          />
        </Card>
      </Col>
    </Row>
  );
};
