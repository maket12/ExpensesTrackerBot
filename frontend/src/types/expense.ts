export type ExpenseCategory =
  | 'еда'
  | 'транспорт'
  | 'жильё'
  | 'развлечения'
  | 'здоровье'
  | 'покупки'
  | 'другое';

export const ALL_CATEGORIES: ExpenseCategory[] = [
  'еда',
  'транспорт',
  'жильё',
  'развлечения',
  'здоровье',
  'покупки',
  'другое',
];

export interface Expense {
  id: string;
  user_id: number;
  amount: number;
  category: ExpenseCategory;
  spent_at: string;
  comment: string;
  created_at: string;
  updated_at: string;
}

export interface CreateExpensePayload {
  user_id: number;
  amount: number;
  category: ExpenseCategory;
  spent_at: string;
  comment: string;
}

export interface UpdateExpensePayload {
  user_id: number;
  amount: number;
  category: ExpenseCategory;
  spent_at: string;
  comment: string;
}

export interface ExpenseFilters {
  user_id?: number;
  category?: string;
  from?: string;
  to?: string;
  limit?: number;
  offset?: number;
}
