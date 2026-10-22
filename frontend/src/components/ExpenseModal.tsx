import React, { useEffect, useState } from 'react';
import { Modal, Form, Input, InputNumber, Select, DatePicker, Alert } from 'antd';
import dayjs from 'dayjs';
import {
  Expense,
  CreateExpensePayload,
  UpdateExpensePayload,
  ALL_CATEGORIES,
  ExpenseCategory,
} from '../types/expense';
import { getErrorMessage, attachFieldError } from '../utils/errors';

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
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const isEditing = !!expense;

  useEffect(() => {
    setErrorMessage(null);
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
      setErrorMessage(null);
      await onSubmit({
        user_id: Number(values.user_id),
        amount: Number(values.amount),
        category: values.category as ExpenseCategory,
        spent_at: values.spent_at.toISOString(),
        comment: values.comment || '',
      });
      form.resetFields();
    } catch (err: any) {
      if (err?.errorFields) {
        return;
      }
      const msg = getErrorMessage(err, 'Не удалось сохранить расход');
      setErrorMessage(msg);
      attachFieldError(form, err, msg);
    }
  };

  const handleCancel = () => {
    setErrorMessage(null);
    onCancel();
  };

  return (
    <Modal
      title={isEditing ? 'Редактировать расход' : 'Новый расход'}
      open={visible}
      onOk={handleOk}
      onCancel={handleCancel}
      confirmLoading={loading}
      okText={isEditing ? 'Сохранить' : 'Добавить'}
      cancelText="Отмена"
      destroyOnClose
    >
      {errorMessage && (
        <Alert
          message="Ошибка сохранения"
          description={errorMessage}
          type="error"
          showIcon
          closable
          onClose={() => setErrorMessage(null)}
          style={{ marginBottom: 16 }}
        />
      )}

      <Form
        form={form}
        layout="vertical"
        style={{ marginTop: 16 }}
        onValuesChange={() => {
          if (errorMessage) {
            setErrorMessage(null);
          }
        }}
      >
        <Form.Item
          name="user_id"
          label="Telegram ID члена семьи"
          rules={[
            { required: true, message: 'Укажите ID члена семьи' },
            {
              type: 'number',
              min: 1,
              message: 'Telegram ID должен быть положительным числом',
            },
          ]}
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
              min: 1,
              max: 1000000,
              message: 'Сумма должна быть от 1.00 до 1 000 000.00 ₽',
            },
          ]}
        >
          <InputNumber
            style={{ width: '100%' }}
            min={1}
            max={1000000}
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
            disabledDate={(current) => {
              return (
                current &&
                (current.isAfter(dayjs().endOf('day')) ||
                  current.isBefore(dayjs().subtract(30, 'days').startOf('day')))
              );
            }}
          />
        </Form.Item>

        <Form.Item
          name="comment"
          label="Комментарий"
          rules={[
            { max: 256, message: 'Комментарий не может превышать 256 символов' },
            {
              validator: async (_, value) => {
                if (value && value.trim().length > 0 && value.trim().length < 2) {
                  return Promise.reject(
                    new Error('Комментарий должен содержать не менее 2 символов')
                  );
                }
                return Promise.resolve();
              },
            },
          ]}
        >
          <Input.TextArea
            rows={2}
            placeholder="Необязательное описание (кофе, бензин, продукты...)"
          />
        </Form.Item>
      </Form>
    </Modal>
  );
};
