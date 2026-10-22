import React from 'react';
import { Card, Row, Col, Select, DatePicker, Button, Space } from 'antd';
import { ReloadOutlined } from '@ant-design/icons';
import dayjs, { Dayjs } from 'dayjs';
import { ALL_CATEGORIES, ExpenseFilters } from '../types/expense';

interface FiltersBarProps {
  distinctUsers: number[];
  filters: ExpenseFilters;
  onChange: (filters: ExpenseFilters) => void;
  onReset: () => void;
}

export const FiltersBar: React.FC<FiltersBarProps> = ({
  distinctUsers,
  filters,
  onChange,
  onReset,
}) => {
  const handleUserChange = (val?: number) => {
    onChange({ ...filters, user_id: val });
  };

  const handleCategoryChange = (val?: string) => {
    onChange({ ...filters, category: val });
  };

  const handleDateRangeChange = (dates: [Dayjs | null, Dayjs | null] | null) => {
    if (!dates || !dates[0] || !dates[1]) {
      onChange({ ...filters, from: undefined, to: undefined });
      return;
    }
    onChange({
      ...filters,
      from: dates[0].startOf('day').toISOString(),
      to: dates[1].endOf('day').toISOString(),
    });
  };

  const dateValue: [Dayjs, Dayjs] | null =
    filters.from && filters.to
      ? [dayjs(filters.from), dayjs(filters.to)]
      : null;

  return (
    <Card
      bordered={false}
      style={{ marginBottom: 24, borderRadius: 8, boxShadow: '0 1px 3px rgba(0,0,0,0.08)' }}
    >
      <Row gutter={[16, 16]} align="middle">
        <Col xs={24} sm={12} md={6}>
          <Space direction="vertical" style={{ width: '100%' }} size={2}>
            <span style={{ fontSize: 12, color: '#8c8c8c' }}>Член семьи (User ID):</span>
            <Select
              allowClear
              placeholder="Все члены семьи"
              style={{ width: '100%' }}
              value={filters.user_id}
              onChange={handleUserChange}
              options={distinctUsers.map((uid) => ({
                label: `Пользователь #${uid}`,
                value: uid,
              }))}
            />
          </Space>
        </Col>

        <Col xs={24} sm={12} md={6}>
          <Space direction="vertical" style={{ width: '100%' }} size={2}>
            <span style={{ fontSize: 12, color: '#8c8c8c' }}>Категория:</span>
            <Select
              allowClear
              placeholder="Все категории"
              style={{ width: '100%' }}
              value={filters.category}
              onChange={handleCategoryChange}
              options={[
                { label: 'Все категории', value: 'все' },
                ...ALL_CATEGORIES.map((cat) => ({
                  label: cat.charAt(0).toUpperCase() + cat.slice(1),
                  value: cat,
                })),
              ]}
            />
          </Space>
        </Col>

        <Col xs={24} sm={16} md={8}>
          <Space direction="vertical" style={{ width: '100%' }} size={2}>
            <span style={{ fontSize: 12, color: '#8c8c8c' }}>Период трат:</span>
            <DatePicker.RangePicker
              style={{ width: '100%' }}
              format="DD.MM.YYYY"
              value={dateValue}
              onChange={handleDateRangeChange}
            />
          </Space>
        </Col>

        <Col xs={24} sm={8} md={4} style={{ display: 'flex', alignItems: 'flex-end' }}>
          <Button
            icon={<ReloadOutlined />}
            onClick={onReset}
            style={{ width: '100%', marginTop: 18 }}
          >
            Сбросить
          </Button>
        </Col>
      </Row>
    </Card>
  );
};
