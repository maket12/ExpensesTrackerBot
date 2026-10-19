package dto

type GetExpenseInput struct {
	ID string
}

type GetExpenseOutput struct {
	Expense ExpenseDTO
}
