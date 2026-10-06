package requests

// DebitSaldoRequest represents a debit operation against a card's saldo.
type DebitSaldoRequest struct {
	CardNumber  string `json:"card_number" validate:"required"`
	Amount      int    `json:"amount" validate:"required,min=1"`
	OperationID string `json:"operation_id" validate:"required"`
	SourceType  string `json:"source_type" validate:"required"`
	SourceID    string `json:"source_id" validate:"required"`
}

// CreditSaldoRequest represents a credit operation to a card's saldo.
type CreditSaldoRequest struct {
	CardNumber  string `json:"card_number" validate:"required"`
	Amount      int    `json:"amount" validate:"required,min=1"`
	OperationID string `json:"operation_id" validate:"required"`
	SourceType  string `json:"source_type" validate:"required"`
	SourceID    string `json:"source_id" validate:"required"`
}
