package models

// SaldoMutationResult represents the result of a saldo mutation operation
// (debit, credit, balance update, withdrawal update).
type SaldoMutationResult struct {
	SaldoID      int32   `json:"saldo_id"`
	CardNumber   string  `json:"card_number"`
	TotalBalance float64 `json:"total_balance"`
}
