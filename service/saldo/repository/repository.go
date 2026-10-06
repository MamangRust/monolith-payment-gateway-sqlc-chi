package repository

import (
	saldostatsrepository "github.com/MamangRust/monolith-payment-gateway-saldo/repository/stats"
	"github.com/MamangRust/monolith-payment-gateway-pkg/adapter"
	pbcard "github.com/MamangRust/monolith-payment-gateway-pb/card"
	"gorm.io/gorm"
)

// GuardOptions configures the guarded gRPC adapters used by the saldo
// repositories (e.g. the card adapter).
type GuardOptions struct {
	Card []adapter.GuardOption
}

// Repositories is a struct containing all saldo repositories.
type Repositories interface {
	SaldoQueryRepository
	SaldoCommandRepository
	saldostatsrepository.SaldoStatsRepository
	CardRepository
}

type repositories struct {
	SaldoQueryRepository
	SaldoCommandRepository
	saldostatsrepository.SaldoStatsRepository
	CardRepository
}

func NewRepositories(
	db *gorm.DB,
	cardQueryClient pbcard.CardQueryServiceClient,
	cardCommandClient pbcard.CardCommandServiceClient,
	guards ...GuardOptions,
) Repositories {
	var g GuardOptions

	if len(guards) > 0 {
		g = guards[0]
	}

	return &repositories{
		SaldoQueryRepository:   NewSaldoQueryRepository(db),
		SaldoCommandRepository: NewSaldoCommandRepository(db),
		SaldoStatsRepository:   saldostatsrepository.NewSaldoStatsRepository(db),
		CardRepository:         adapter.NewCardAdapter(cardQueryClient, cardCommandClient, g.Card...),
	}
}
