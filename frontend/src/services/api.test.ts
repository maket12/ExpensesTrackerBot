import { describe, it, expect, vi, beforeEach } from 'vitest';

const mockGet = vi.fn();
const mockPost = vi.fn();
const mockPut = vi.fn();
const mockDelete = vi.fn();

vi.mock('axios', () => {
  return {
    default: {
      create: () => ({
        get: mockGet,
        post: mockPost,
        put: mockPut,
        delete: mockDelete,
      }),
    },
  };
});

import { api } from './api';
import { CreateExpensePayload, UpdateExpensePayload } from '../types/expense';

describe('api service', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('getExpenses', () => {
    it('should query with filters and return items', async () => {
      const mockItems = [
        { id: '1', user_id: 100, amount: 50, category: 'еда', spent_at: '', comment: '', created_at: '', updated_at: '' },
      ];
      mockGet.mockResolvedValueOnce({ data: { items: mockItems } });

      const result = await api.getExpenses({
        user_id: 100,
        category: 'еда',
        from: '2026-01-01',
        to: '2026-01-31',
        limit: 10,
        offset: 0,
      });

      expect(mockGet).toHaveBeenCalledWith('/expenses', {
        params: {
          user_id: 100,
          category: 'еда',
          from: '2026-01-01',
          to: '2026-01-31',
          limit: 10,
          offset: 0,
        },
      });
      expect(result).toEqual(mockItems);
    });

    it('should omit category when set to "все"', async () => {
      mockGet.mockResolvedValueOnce({ data: { items: [] } });

      await api.getExpenses({ category: 'все' });

      expect(mockGet).toHaveBeenCalledWith('/expenses', { params: {} });
    });

    it('should fallback to empty array when response has no items', async () => {
      mockGet.mockResolvedValueOnce({ data: {} });

      const result = await api.getExpenses();

      expect(result).toEqual([]);
    });
  });

  describe('createExpense', () => {
    it('should unwrap expense object from response', async () => {
      const payload: CreateExpensePayload = {
        user_id: 100,
        amount: 250,
        category: 'еда',
        spent_at: '2026-10-22T10:00:00Z',
        comment: 'lunch',
      };
      const createdExpense = { id: 'uuid-1', ...payload, created_at: '', updated_at: '' };
      mockPost.mockResolvedValueOnce({ data: { expense: createdExpense } });

      const result = await api.createExpense(payload);

      expect(mockPost).toHaveBeenCalledWith('/expenses', payload);
      expect(result).toEqual(createdExpense);
    });
  });

  describe('updateExpense', () => {
    it('should send update request and unwrap response', async () => {
      const payload: UpdateExpensePayload = {
        user_id: 100,
        amount: 300,
        category: 'транспорт',
        spent_at: '2026-10-22T12:00:00Z',
        comment: 'taxi',
      };
      const updatedExpense = { id: 'uuid-1', ...payload, created_at: '', updated_at: '' };
      mockPut.mockResolvedValueOnce({ data: { expense: updatedExpense } });

      const result = await api.updateExpense('uuid-1', payload);

      expect(mockPut).toHaveBeenCalledWith('/expenses/uuid-1', payload);
      expect(result).toEqual(updatedExpense);
    });
  });

  describe('deleteExpense', () => {
    it('should send delete request with id', async () => {
      mockDelete.mockResolvedValueOnce({ status: 204 });

      await api.deleteExpense('uuid-1');

      expect(mockDelete).toHaveBeenCalledWith('/expenses/uuid-1');
    });
  });

  describe('getDistinctUsers', () => {
    it('should return user ids list', async () => {
      mockGet.mockResolvedValueOnce({ data: { user_ids: [100, 200, 300] } });

      const result = await api.getDistinctUsers();

      expect(mockGet).toHaveBeenCalledWith('/users');
      expect(result).toEqual([100, 200, 300]);
    });

    it('should fallback to empty array if user_ids is undefined', async () => {
      mockGet.mockResolvedValueOnce({ data: {} });

      const result = await api.getDistinctUsers();

      expect(result).toEqual([]);
    });
  });
});
