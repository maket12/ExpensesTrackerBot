import axios from 'axios';
import {
  Expense,
  CreateExpensePayload,
  UpdateExpensePayload,
  ExpenseFilters,
} from '../types/expense';

const client = axios.create({
  baseURL: '/api',
  headers: {
    'Content-Type': 'application/json',
  },
});

export const api = {
  async getExpenses(filters?: ExpenseFilters): Promise<Expense[]> {
    const params: Record<string, string | number> = {};
    if (filters?.user_id !== undefined) params.user_id = filters.user_id;
    if (filters?.category && filters.category !== 'все') params.category = filters.category;
    if (filters?.from) params.from = filters.from;
    if (filters?.to) params.to = filters.to;
    if (filters?.limit !== undefined) params.limit = filters.limit;
    if (filters?.offset !== undefined) params.offset = filters.offset;

    const resp = await client.get<{ items?: Expense[]; expenses?: Expense[] }>('/expenses', { params });
    return resp.data.items || resp.data.expenses || [];
  },

  async createExpense(payload: CreateExpensePayload): Promise<Expense> {
    const resp = await client.post<{ expense?: Expense } & Expense>('/expenses', payload);
    return resp.data.expense || resp.data;
  },

  async updateExpense(id: string, payload: UpdateExpensePayload): Promise<Expense> {
    const resp = await client.put<{ expense?: Expense } & Expense>(`/expenses/${id}`, payload);
    return resp.data.expense || resp.data;
  },

  async deleteExpense(id: string): Promise<void> {
    await client.delete(`/expenses/${id}`);
  },

  async getDistinctUsers(): Promise<number[]> {
    const resp = await client.get<{ user_ids: number[] }>('/users');
    return resp.data.user_ids || [];
  },
};
