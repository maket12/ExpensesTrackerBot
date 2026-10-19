import React, { useEffect } from 'react';
import { Modal, Form, Input, InputNumber, Select, DatePicker } from 'antd';
import dayjs from 'dayjs';
import {
  Expense,
  CreateExpensePayload,
  UpdateExpensePayload,
  ALL_CATEGORIES,
  ExpenseCategory,
} from '../types/expense';

interface ExpenseModalProps {
  visible: boolean;
  expense: Expense | null;
  onCancel: () => void;
  onSubmit: (payload: CreateExpensePayload | UpdateExpensePayload) => Promise<void>;
  loading: boolean;
}

export const ExpenseModal: React.FC<ExpenseModalProps> = ({
  visible,
  expense,
  onCancel,
  onSubmit,
  loading,
}) => {
  const [form] = Form.useForm();
  const isEditing = !!expense;

  useEffect(() => {
    if (visible) {
      if (expense) {
        form.setFieldsValue({
          user_id: expense.user_id,
          amount: expense.amount,
          category: expense.category,
          spent_at: dayjs(expense.spent_at),
          comment: expense.comment,
        });
      } else {
        form.resetFields();
        form.setFieldsValue({
          spent_at: dayjs(),
          category: 'еда',
        });
      }
    }
  }, [visible, expense, form]);

  const handleOk = async () => {
    try {
      const values = await form.validateFields();
      await onSubmit({
        user_id: Number(values.user_id),
        amount: Number(values.amount),
        category: values.category as ExpenseCategory,
        spent_at: values.spent_at.toISOString(),
        comment: values.comment || '',
      });
      form.resetFields();
    } catch {
      // Form validation errors handled by antd
    }
  };

  return (
    <Modal
      title={isEditing ? 'Редактировать расход' : 'Новый расход'}
      open={visible}
      onOk={handleOk}
      onCancel={onCancel}
      confirmLoading={loading}
      okText={isEditing ? 'Сохранить' : 'Добавить'}
      cancelText="Отмена"
      destroyOnClose
    >
      <Form form={form} layout="vertical" style={{ marginTop: 16 }}>
        <Form.Item
          name="user_id"
          label="Telegram ID члена семьи"
          rules={[{ required: true, message: 'Укажите ID члена семьи' }]}
        >
          <InputNumber
            style={{ width: '100%' }}
            placeholder="Например: 12345678"
            min={1}
          />
        </Form.Item>

        <Form.Item
          name="amount"
          label="Сумма расхода (₽)"
          rules={[
            { required: true, message: 'Укажите сумму' },
            {
              type: 'number',
              min: 0.01,
              message: 'Сумма должна быть больше нуля',
            },
          ]}
        >
          <InputNumber
            style={{ width: '100%' }}
            min={0.01}
            step={0.01}
            precision={2}
            placeholder="0.00"
          />
        </Form.Item>

        <Form.Item
          name="category"
          label="Категория"
          rules={[{ required: true, message: 'Выберите категорию' }]}
        >
          <Select
            placeholder="Выберите категорию"
            options={ALL_CATEGORIES.map((cat) => ({
              label: cat.charAt(0).toUpperCase() + cat.slice(1),
              value: cat,
            }))}
          />
        </Form.Item>

        <Form.Item
          name="spent_at"
          label="Дата и время расхода"
          rules={[{ required: true, message: 'Укажите дату расхода' }]}
        >
          <DatePicker
            showTime
            format="DD.MM.YYYY HH:mm"
            style={{ width: '100%' }}
          />
        </Form.Item>

        <Form.Item name="comment" label="Комментарий">
          <Input.TextArea
            rows={2}
            placeholder="Необязательное описание (кофе, бензин, продукты...)"
          />
        </Form.Item>
      </Form>
    </Modal>
  );
};
