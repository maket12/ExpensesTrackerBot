import { describe, it, expect, vi } from 'vitest';
import { AxiosError, AxiosHeaders } from 'axios';
import {
  getRawApiError,
  formatApiError,
  getErrorMessage,
  attachFieldError,
} from './errors';

function createAxiosError(status?: number, data?: any, code?: string): AxiosError {
  const headers = new AxiosHeaders();
  return {
    isAxiosError: true,
    name: 'AxiosError',
    message: 'Request failed',
    code,
    config: { headers },
    response: status
      ? {
          status,
          statusText: 'Error',
          data,
          headers: {},
          config: { headers },
        }
      : undefined,
  } as unknown as AxiosError;
}

describe('errors utility', () => {
  describe('getRawApiError', () => {
    it('returns empty string for non-Axios error', () => {
      expect(getRawApiError(new Error('regular error'))).toBe('');
      expect(getRawApiError('string error')).toBe('');
      expect(getRawApiError(null)).toBe('');
    });

    it('returns error field from response data object', () => {
      const err = createAxiosError(400, { error: 'value is invalid: amount' });
      expect(getRawApiError(err)).toBe('value is invalid: amount');
    });

    it('returns response data if it is a string', () => {
      const err = createAxiosError(400, 'raw string error');
      expect(getRawApiError(err)).toBe('raw string error');
    });

    it('returns empty string if response has no recognized error field', () => {
      const err = createAxiosError(400, { details: 'something else' });
      expect(getRawApiError(err)).toBe('');
    });

    it('returns message field from response data object if present', () => {
      const err = createAxiosError(400, { message: 'some error message' });
      expect(getRawApiError(err)).toBe('some error message');
    });
  });

  describe('formatApiError', () => {
    it('translates internal error', () => {
      expect(formatApiError('internal error')).toContain('Внутренняя ошибка сервера');
    });

    it('translates invalid input general error', () => {
      expect(formatApiError('invalid input')).toContain('Некорректно заполнены поля');
    });

    it('translates amount validation error', () => {
      expect(formatApiError('value is invalid: amount')).toContain('1.00 до 1 000 000.00 ₽');
    });

    it('translates spent_at future error', () => {
      expect(formatApiError('spent_at cannot be after created_at')).toBe(
        'Дата и время расхода не могут быть в будущем'
      );
    });

    it('translates spent_at invalid error', () => {
      expect(formatApiError('value is invalid: spent_at')).toContain('не старше 30 дней');
    });

    it('translates user_id error', () => {
      expect(formatApiError('value is invalid: user_id')).toContain('Telegram ID');
    });

    it('translates category error', () => {
      expect(formatApiError('value is invalid: category')).toContain('Некорректная категория');
    });

    it('translates comment error', () => {
      expect(formatApiError('value is invalid: comment')).toContain('от 2 до 256 символов');
    });

    it('translates not found error', () => {
      expect(formatApiError('expense not found')).toContain('Расход не найден');
    });

    it('translates access forbidden error', () => {
      expect(formatApiError('access forbidden')).toContain('запрещён');
    });

    it('translates invalid json error', () => {
      expect(formatApiError('invalid json')).toContain('Некорректный формат данных');
    });

    it('translates invalid identifier error', () => {
      expect(formatApiError('invalid identifier')).toContain('UUID');
    });

    it('translates value is required errors', () => {
      expect(formatApiError('value is required: amount')).toContain('сумму');
      expect(formatApiError('value is required: spent_at')).toContain('дату');
      expect(formatApiError('value is required: user_id')).toContain('ID');
      expect(formatApiError('value is required: category')).toContain('категорию');
      expect(formatApiError('value is required: unknown_field')).toContain('Обязательное поле');
    });

    it('returns raw error as fallback for unknown message', () => {
      expect(formatApiError('Custom domain failure')).toBe('Custom domain failure');
    });
  });

  describe('getErrorMessage', () => {
    it('extracts formatted message from Axios response error', () => {
      const err = createAxiosError(400, { error: 'value is invalid: amount' });
      expect(getErrorMessage(err)).toContain('1.00 до 1 000 000.00 ₽');
    });

    it('handles status 400 without specific body error', () => {
      const err = createAxiosError(400, {});
      expect(getErrorMessage(err)).toContain('400 Bad Request');
    });

    it('handles status 404 without specific body error', () => {
      const err = createAxiosError(404, {});
      expect(getErrorMessage(err)).toContain('404 Not Found');
    });

    it('handles status 500 without specific body error', () => {
      const err = createAxiosError(500, {});
      expect(getErrorMessage(err)).toContain('500 Internal Server Error');
    });

    it('handles network error', () => {
      const err = createAxiosError(undefined, undefined, 'ERR_NETWORK');
      expect(getErrorMessage(err)).toContain('подключение к сети');
    });

    it('handles standard Error instance', () => {
      const err = new Error('Local error message');
      expect(getErrorMessage(err)).toBe('Local error message');
    });

    it('returns fallback for unknown non-error', () => {
      expect(getErrorMessage(null, 'Кастомный фолбэк')).toBe('Кастомный фолбэк');
    });
  });

  describe('attachFieldError', () => {
    it('attaches error to amount field', () => {
      const form = { setFields: vi.fn() } as any;
      const err = createAxiosError(400, { error: 'value is invalid: amount' });
      attachFieldError(form, err, 'Сумма некорректна');
      expect(form.setFields).toHaveBeenCalledWith([
        { name: 'amount', errors: ['Сумма некорректна'] },
      ]);
    });

    it('attaches error to spent_at field', () => {
      const form = { setFields: vi.fn() } as any;
      const err = createAxiosError(400, { error: 'spent_at cannot be after created_at' });
      attachFieldError(form, err, 'Дата в будущем');
      expect(form.setFields).toHaveBeenCalledWith([
        { name: 'spent_at', errors: ['Дата в будущем'] },
      ]);
    });

    it('attaches error to user_id field', () => {
      const form = { setFields: vi.fn() } as any;
      const err = createAxiosError(400, { error: 'value is invalid: user_id' });
      attachFieldError(form, err, 'Неверный ID');
      expect(form.setFields).toHaveBeenCalledWith([
        { name: 'user_id', errors: ['Неверный ID'] },
      ]);
    });

    it('attaches error to category field', () => {
      const form = { setFields: vi.fn() } as any;
      const err = createAxiosError(400, { error: 'value is invalid: category' });
      attachFieldError(form, err, 'Неверная категория');
      expect(form.setFields).toHaveBeenCalledWith([
        { name: 'category', errors: ['Неверная категория'] },
      ]);
    });

    it('attaches error to comment field', () => {
      const form = { setFields: vi.fn() } as any;
      const err = createAxiosError(400, { error: 'value is invalid: comment' });
      attachFieldError(form, err, 'Неверный комментарий');
      expect(form.setFields).toHaveBeenCalledWith([
        { name: 'comment', errors: ['Неверный комментарий'] },
      ]);
    });

    it('does not attach field error if error has no match', () => {
      const form = { setFields: vi.fn() } as any;
      const err = createAxiosError(500, { error: 'internal error' });
      attachFieldError(form, err, 'Ошибка');
      expect(form.setFields).not.toHaveBeenCalled();
    });
  });
});
