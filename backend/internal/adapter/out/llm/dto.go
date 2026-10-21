package llm

type ToolParameterProperty struct {
	Type        string   `json:"type"`
	Description string   `json:"description,omitempty"`
	Enum        []string `json:"enum,omitempty"`
}

type ToolParameters struct {
	Type       string                           `json:"type"`
	Properties map[string]ToolParameterProperty `json:"properties"`
	Required   []string                         `json:"required,omitempty"`
}

type AddExpenseToolPayload struct {
	Amount   float64 `json:"amount"`
	Category string  `json:"category"`
	SpentAt  string  `json:"spent_at"`
	Comment  string  `json:"comment"`
}

type ListExpensesToolPayload struct {
	Limit    int     `json:"limit"`
	Category *string `json:"category"`
	From     *string `json:"from"`
	To       *string `json:"to"`
}

type UpdateExpenseToolPayload struct {
	ID       string   `json:"id"`
	Amount   *float64 `json:"amount"`
	Category *string  `json:"category"`
	SpentAt  *string  `json:"spent_at"`
	Comment  *string  `json:"comment"`
}

type DeleteExpenseToolPayload struct {
	ID string `json:"id"`
}
