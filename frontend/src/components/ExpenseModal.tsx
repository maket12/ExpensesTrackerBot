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
      const trimmedComment = values.comment ? String(values.comment).trim() : '';
      await onSubmit({
        user_id: Number(values.user_id),
        amount: Number(values.amount),
        category: values.category as ExpenseCategory,
        spent_at: values.spent_at.toISOString(),
        comment: trimmedComment,
      });
      form.resetFields();
    } catch (err: any) {
      if (err?.errorFields) {
        return;
      }
      const msg = getErrorMessage(err, 'Не удалось сохранить расход. Проверьте введённые данные.');
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
          message="Ошибка при сохранении"
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
            { required: true, message: 'Укажите Telegram ID члена семьи' },
            {
              validator: async (_, value) => {
                if (value === undefined || value === null || value === '') {
                  return Promise.resolve();
                }
                const num = Number(value);
                if (!Number.isFinite(num) || !Number.isInteger(num)) {
                  return Promise.reject(new Error('Telegram ID должен быть целым числом'));
                }
                if (num <= 0) {
                  return Promise.reject(
                    new Error('Telegram ID должен быть положительным числом (> 0)')
                  );
                }
                if (num > Number.MAX_SAFE_INTEGER) {
                  return Promise.reject(new Error('Слишком большой ID пользователя'));
                }
                return Promise.resolve();
              },
            },
          ]}
        >
          <InputNumber
            style={{ width: '100%' }}
            placeholder="Например: 12345678"
            precision={0}
            controls={false}
          />
        </Form.Item>

        <Form.Item
          name="amount"
          label="Сумма расхода (₽)"
          rules={[
            { required: true, message: 'Укажите сумму расхода' },
            {
              validator: async (_, value) => {
                if (value === undefined || value === null || value === '') {
                  return Promise.resolve();
                }
                const num = Number(value);
                if (!Number.isFinite(num)) {
                  return Promise.reject(new Error('Введите корректное число'));
                }
                if (num < 1.0) {
                  return Promise.reject(
                    new Error('Сумма расхода должна быть не менее 1.00 ₽')
                  );
                }
                if (num > 1000000.0) {
                  return Promise.reject(
                    new Error('Сумма расхода не может превышать 1 000 000.00 ₽')
                  );
                }
                return Promise.resolve();
              },
            },
          ]}
        >
          <InputNumber
            style={{ width: '100%' }}
            step={0.01}
            precision={2}
            controls={false}
            placeholder="0.00"
          />
        </Form.Item>

        <Form.Item
          name="category"
          label="Категория"
          rules={[
            { required: true, message: 'Выберите категорию расхода' },
            {
              validator: async (_, value) => {
                if (!value) return Promise.resolve();
                if (!ALL_CATEGORIES.includes(value)) {
                  return Promise.reject(new Error('Недопустимая категория расхода'));
                }
                return Promise.resolve();
              },
            },
          ]}
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
          rules={[
            { required: true, message: 'Укажите дату и время расхода' },
            {
              validator: async (_, value) => {
                if (!value) return Promise.resolve();
                const now = dayjs();
                if (value.isAfter(now)) {
                  return Promise.reject(
                    new Error('Дата и время расхода не могут быть в будущем')
                  );
                }
                const earliest = now.subtract(30, 'days');
                if (value.isBefore(earliest)) {
                  return Promise.reject(
                    new Error('Дата расхода не может быть старше 30 дней')
                  );
                }
                return Promise.resolve();
              },
            },
          ]}
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
            {
              validator: async (_, value) => {
                if (!value) return Promise.resolve();
                const str = String(value);
                const trimmed = str.trim();
                if (str.length > 0 && trimmed.length === 0) {
                  return Promise.reject(
                    new Error('Комментарий не может состоять только из пробелов')
                  );
                }
                if (trimmed.length > 0 && trimmed.length < 2) {
                  return Promise.reject(
                    new Error('Комментарий должен содержать не менее 2 символов')
                  );
                }
                if (trimmed.length > 256) {
                  return Promise.reject(
                    new Error('Комментарий не может превышать 256 символов')
                  );
                }
                return Promise.resolve();
              },
            },
          ]}
        >
          <Input.TextArea
            rows={2}
            maxLength={256}
            placeholder="Необязательное описание (кофе, бензин, продукты...)"
          />
        </Form.Item>
      </Form>
    </Modal>
  );
};
