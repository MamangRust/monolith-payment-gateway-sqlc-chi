package repository

import (
	adapter "github.com/MamangRust/monolith-payment-gateway-pkg/adapter"
	pbcard "github.com/MamangRust/monolith-payment-gateway-pb/card"
	pbsaldo "github.com/MamangRust/monolith-payment-gateway-pb/saldo"
	transferstatsrepository "github.com/MamangRust/monolith-payment-gateway-transfer/repository/stats"
	transferstatsbycardrepository "github.com/MamangRust/monolith-payment-gateway-transfer/repository/statsbycard"
	"gorm.io/gorm"
)

// GuardOptions collects the dependency guards applied to the card and saldo
// gRPC adapters built inside NewRepositories. The card and saldo sides get
// separate breakers so one failure mode cannot starve the other.
type GuardOptions struct {
	Card  []adapter.GuardOption
	Saldo []adapter.GuardOption
}

type Repositories interface {
	SaldoRepository
	TransferQueryRepository
	TransferCommandRepository
	CardRepository
	transferstatsrepository.TransferStatsRepository
	transferstatsbycardrepository.TransferStatsByCardRepository
}

type repositories struct {
	SaldoRepository
	TransferQueryRepository
	TransferCommandRepository
	CardRepository
	transferstatsrepository.TransferStatsRepository
	transferstatsbycardrepository.TransferStatsByCardRepository
}

func NewRepositories(
	db *gorm.DB,
	cardQueryClient pbcard.CardQueryServiceClient,
	cardCommandClient pbcard.CardCommandServiceClient,
	saldoQueryClient pbsaldo.SaldoQueryServiceClient,
	saldoCommandClient pbsaldo.SaldoCommandServiceClient,
	guards ...GuardOptions,
) Repositories {
	var g GuardOptions
	if len(guards) > 0 {
		g = guards[0]
	}

	return &repositories{
		SaldoRepository:               adapter.NewSaldoAdapter(saldoQueryClient, saldoCommandClient, g.Saldo...),
		TransferQueryRepository:       NewTransferQueryRepository(db),
		TransferCommandRepository:     NewTransferCommandRepository(db),
		TransferStatsRepository:       transferstatsrepository.NewTransferStatsRepository(db),
		TransferStatsByCardRepository: transferstatsbycardrepository.NewTransferStatsByCardRepository(db),
		CardRepository:                adapter.NewCardAdapter(cardQueryClient, cardCommandClient, g.Card...),
	}
}
