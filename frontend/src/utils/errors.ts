import axios from 'axios';
import type { FormInstance } from 'antd';

export function getRawApiError(error: unknown): string {
  if (axios.isAxiosError(error)) {
    const data = error.response?.data;
    if (!data) {
      return error.response?.statusText || '';
    }
    if (typeof data === 'string') {
      return data;
    }
    if (typeof data === 'object') {
      if ('error' in data) {
        if (typeof data.error === 'string') return data.error;
        if (
          data.error &&
          typeof data.error === 'object' &&
          'message' in data.error &&
          typeof data.error.message === 'string'
        ) {
          return data.error.message;
        }
      }
      if ('message' in data && typeof data.message === 'string') {
        return data.message;
      }
      if ('msg' in data && typeof data.msg === 'string') {
        return data.msg;
      }
      if ('detail' in data && typeof data.detail === 'string') {
        return data.detail;
      }
    }
  }
  return '';
}

export function formatApiError(rawError: string): string {
  const normalized = rawError.toLowerCase().trim();

  if (normalized.includes('value is required')) {
    if (normalized.includes('amount')) return 'Укажите сумму расхода';
    if (normalized.includes('spent_at')) return 'Укажите дату расхода';
    if (normalized.includes('user_id')) return 'Укажите Telegram ID пользователя';
    if (normalized.includes('category')) return 'Выберите категорию расхода';
    return 'Обязательное поле не заполнено';
  }
  if (normalized.includes('amount')) {
    return 'Некорректная сумма расхода (должна быть от 1.00 до 1 000 000.00 ₽)';
  }
  if (
    normalized.includes('spent_at cannot be after created_at') ||
    (normalized.includes('spent_at') && normalized.includes('future'))
  ) {
    return 'Дата и время расхода не могут быть в будущем';
  }
  if (normalized.includes('spent_at')) {
    return 'Некорректная дата расхода (не может быть в будущем и не старше 30 дней)';
  }
  if (normalized.includes('user_id')) {
    return 'Некорректный Telegram ID пользователя (должен быть положительным числом)';
  }
  if (normalized.includes('category')) {
    return 'Некорректная категория расхода (выберите из предложенного списка)';
  }
  if (normalized.includes('comment')) {
    return 'Комментарий должен содержать от 2 до 256 символов';
  }
  if (normalized.includes('expense not found') || normalized.includes('object not found')) {
    return 'Расход не найден (возможно, запись уже была удалена)';
  }
  if (normalized.includes('access forbidden') || normalized.includes('forbidden')) {
    return 'Доступ к данному расходу запрещён';
  }
  if (normalized.includes('invalid json')) {
    return 'Некорректный формат данных запроса';
  }
  if (normalized.includes('invalid identifier')) {
    return 'Некорректный формат идентификатора (UUID)';
  }
  if (normalized.includes('internal error') || normalized.includes('internal server error')) {
    return 'Внутренняя ошибка сервера. Пожалуйста, повторите попытку позже.';
  }
  if (normalized.includes('invalid input')) {
    return 'Некорректно заполнены поля формы. Проверьте введенные данные.';
  }

  return rawError;
}

export function getErrorMessage(
  error: unknown,
  fallbackMessage = 'Произошла непредвиденная ошибка'
): string {
  if (axios.isAxiosError(error)) {
    const rawError = getRawApiError(error);
    if (rawError) {
      const formatted = formatApiError(rawError);
      if (formatted && formatted.trim().length > 0) {
        return formatted;
      }
    }
    if (error.response?.status === 400) {
      return 'Некорректные параметры запроса (400 Bad Request)';
    }
    if (error.response?.status === 404) {
      return 'Запрашиваемый ресурс не найден (404 Not Found)';
    }
    if (error.response?.status === 409) {
      return 'Конфликт данных: запись уже существует';
    }
    if (error.response?.status === 500) {
      return 'Внутренняя ошибка сервера (500 Internal Server Error)';
    }
    if (error.code === 'ERR_NETWORK') {
      return 'Не удалось связаться с сервером. Проверьте подключение к сети.';
    }
    if (error.message) {
      return error.message;
    }
  }

  if (error instanceof Error) {
    return error.message;
  }

  return fallbackMessage;
}

export function attachFieldError(
  form: FormInstance,
  error: unknown,
  messageText: string
): void {
  const rawMsg = getRawApiError(error).toLowerCase();
  if (!rawMsg) return;

  if (rawMsg.includes('amount')) {
    form.setFields([{ name: 'amount', errors: [messageText] }]);
  } else if (rawMsg.includes('spent_at') || rawMsg.includes('created_at')) {
    form.setFields([{ name: 'spent_at', errors: [messageText] }]);
  } else if (rawMsg.includes('user_id')) {
    form.setFields([{ name: 'user_id', errors: [messageText] }]);
  } else if (rawMsg.includes('category')) {
    form.setFields([{ name: 'category', errors: [messageText] }]);
  } else if (rawMsg.includes('comment')) {
    form.setFields([{ name: 'comment', errors: [messageText] }]);
  }
}
