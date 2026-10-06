package repository

import (
	adapter "github.com/MamangRust/monolith-payment-gateway-pkg/adapter"
	pbcard "github.com/MamangRust/monolith-payment-gateway-pb/card"
	pbmerchant "github.com/MamangRust/monolith-payment-gateway-pb/merchant"
	pbsaldo "github.com/MamangRust/monolith-payment-gateway-pb/saldo"
	transactionstatsrepository "github.com/MamangRust/monolith-payment-gateway-transaction/repository/stats"
	transactionbycardrepository "github.com/MamangRust/monolith-payment-gateway-transaction/repository/statsbycard"
	"gorm.io/gorm"
)

// GuardOptions collects the dependency guards applied to the card, saldo and
// merchant gRPC adapters built inside NewRepositories. Each side gets a
// separate breaker so one failure mode cannot starve the others.
type GuardOptions struct {
	Card     []adapter.GuardOption
	Saldo    []adapter.GuardOption
	Merchant []adapter.GuardOption
}

type Repositories interface {
	SaldoRepository
	MerchantRepository
	CardRepository
	TransactionQueryRepository
	TransactionCommandRepository
	transactionstatsrepository.TransactionStatsRepository
	transactionbycardrepository.TransactionStatsByCardRepository
}

type repositories struct {
	SaldoRepository
	MerchantRepository
	CardRepository
	TransactionQueryRepository
	TransactionCommandRepository
	transactionstatsrepository.TransactionStatsRepository
	transactionbycardrepository.TransactionStatsByCardRepository
}

func NewRepositories(
	db *gorm.DB,
	cardQueryClient pbcard.CardQueryServiceClient,
	cardCommandClient pbcard.CardCommandServiceClient,
	saldoQueryClient pbsaldo.SaldoQueryServiceClient,
	saldoCommandClient pbsaldo.SaldoCommandServiceClient,
	merchantQueryClient pbmerchant.MerchantQueryServiceClient,
	guards ...GuardOptions,
) Repositories {
	var g GuardOptions
	if len(guards) > 0 {
		g = guards[0]
	}

	return &repositories{
		SaldoRepository:                  adapter.NewSaldoAdapter(saldoQueryClient, saldoCommandClient, g.Saldo...),
		MerchantRepository:               adapter.NewMerchantAdapter(merchantQueryClient, g.Merchant...),
		CardRepository:                   adapter.NewCardAdapter(cardQueryClient, cardCommandClient, g.Card...),
		TransactionQueryRepository:       NewTransactionQueryRepository(db),
		TransactionCommandRepository:     NewTransactionCommandRepository(db),
		TransactionStatsRepository:       transactionstatsrepository.NewTransactionStatsRepository(db),
		TransactionStatsByCardRepository: transactionbycardrepository.NewTransactionStatsByCardRepository(db),
	}
}
