package adapter

import (
	"context"
	"time"

	pbcard "github.com/MamangRust/monolith-payment-gateway-pb/card"
	pbsaldo "github.com/MamangRust/monolith-payment-gateway-pb/saldo"
	"github.com/MamangRust/monolith-payment-gateway-pkg/database/models"
	"github.com/MamangRust/monolith-payment-gateway-pkg/resilience"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
	sharedErrors "github.com/MamangRust/monolith-payment-gateway-shared/errors"
)

type SaldoAdapter interface {
	FindByCardNumber(ctx context.Context, card_number string) (*models.Saldo, error)
	UpdateSaldoBalance(ctx context.Context, request *requests.UpdateSaldoBalance) (*models.UpdateSaldoBalanceRow, error)
	UpdateSaldoWithdraw(ctx context.Context, request *requests.UpdateSaldoWithdraw) (*models.UpdateSaldoWithdrawRow, error)
	DebitSaldo(ctx context.Context, request *requests.DebitSaldoRequest) (*models.SaldoMutationResult, error)
	CreditSaldo(ctx context.Context, request *requests.CreditSaldoRequest) (*models.SaldoMutationResult, error)
}

type saldoGRPCAdapter struct {
	QueryClient   pbsaldo.SaldoQueryServiceClient
	CommandClient pbsaldo.SaldoCommandServiceClient
	guard         *resilience.DependencyGuard
}

func (a *saldoGRPCAdapter) setGuard(g *resilience.DependencyGuard) {
	a.guard = g
}

func NewSaldoAdapter(queryClient pbsaldo.SaldoQueryServiceClient, commandClient pbsaldo.SaldoCommandServiceClient, opts ...GuardOption) SaldoAdapter {
	a := &saldoGRPCAdapter{
		QueryClient:   queryClient,
		CommandClient: commandClient,
	}
	for _, opt := range opts {
		opt(a)
	}
	return a
}

func (a *saldoGRPCAdapter) FindByCardNumber(ctx context.Context, card_number string) (*models.Saldo, error) {
	var resp *pbsaldo.ApiResponseSaldo
	err := a.guard.Call(ctx, func(callCtx context.Context) error {
		var callErr error
		resp, callErr = a.QueryClient.FindByCardNumber(callCtx, &pbcard.FindByCardNumberRequest{
			CardNumber: card_number,
		})
		return callErr
	})
	if err != nil {
		return nil, err
	}
	return mapSaldoResponse(resp.Data)
}

func mapSaldoResponse(s *pbsaldo.SaldoResponse) (*models.Saldo, error) {
	if s == nil {
		return nil, sharedErrors.NewBadRequestError("saldo response is required")
	}
	saldo := &models.Saldo{
		SaldoID:      s.SaldoId,
		CardNumber:   s.CardNumber,
		TotalBalance: s.TotalBalance,
	}
	if s.WithdrawAmount != 0 {
		amount := s.WithdrawAmount
		saldo.WithdrawAmount = &amount
	}
	saldo.WithdrawTime = parseTimePtr(s.WithdrawTime)
	saldo.CreatedAt = parseTime(s.CreatedAt)
	saldo.UpdatedAt = parseTime(s.UpdatedAt)
	return saldo, nil
}

func mapSaldoMutationResponse(resp *pbsaldo.SaldoResponse) *models.SaldoMutationResult {
	if resp == nil {
		return nil
	}
	return &models.SaldoMutationResult{
		SaldoID:      resp.SaldoId,
		CardNumber:   resp.CardNumber,
		TotalBalance: float64(resp.TotalBalance),
	}
}

func mapSaldoUpdateBalanceResponse(resp *pbsaldo.SaldoResponse) *models.UpdateSaldoBalanceRow {
	if resp == nil {
		return nil
	}
	return &models.UpdateSaldoBalanceRow{
		SaldoID:    resp.SaldoId,
		CardNumber: resp.CardNumber,
		TotalBalance: resp.TotalBalance,
		CreatedAt:  parseTime(resp.CreatedAt),
		UpdatedAt:  parseTime(resp.UpdatedAt),
	}
}

func mapSaldoUpdateWithdrawResponse(resp *pbsaldo.SaldoResponse) *models.UpdateSaldoWithdrawRow {
	if resp == nil {
		return nil
	}
	row := &models.UpdateSaldoWithdrawRow{
		SaldoID:      resp.SaldoId,
		CardNumber:   resp.CardNumber,
		TotalBalance: resp.TotalBalance,
		CreatedAt:    parseTime(resp.CreatedAt),
		UpdatedAt:    parseTime(resp.UpdatedAt),
	}
	if resp.WithdrawAmount != 0 {
		amount := resp.WithdrawAmount
		row.WithdrawAmount = &amount
	}
	row.WithdrawTime = parseTimePtr(resp.WithdrawTime)
	return row
}

func (a *saldoGRPCAdapter) UpdateSaldoBalance(ctx context.Context, request *requests.UpdateSaldoBalance) (*models.UpdateSaldoBalanceRow, error) {
	var resp *pbsaldo.ApiResponseSaldo
	err := a.guard.Call(ctx, func(callCtx context.Context) error {
		var callErr error
		resp, callErr = a.CommandClient.UpdateSaldoBalance(callCtx, &pbsaldo.UpdateSaldoBalanceRequest{
			CardNumber:   request.CardNumber,
			TotalBalance: int32(request.TotalBalance),
		})
		return callErr
	})
	if err != nil {
		return nil, err
	}
	return mapSaldoUpdateBalanceResponse(resp.Data), nil
}

func (a *saldoGRPCAdapter) UpdateSaldoWithdraw(ctx context.Context, request *requests.UpdateSaldoWithdraw) (*models.UpdateSaldoWithdrawRow, error) {
	var resp *pbsaldo.ApiResponseSaldo
	err := a.guard.Call(ctx, func(callCtx context.Context) error {
		var callErr error
		resp, callErr = a.CommandClient.UpdateSaldoWithdraw(callCtx, &pbsaldo.UpdateSaldoWithdrawRequest{
			CardNumber:     request.CardNumber,
			TotalBalance:   int32(request.TotalBalance),
			WithdrawAmount: int32(*request.WithdrawAmount),
			WithdrawTime:   request.WithdrawTime.Format(time.RFC3339),
		})
		return callErr
	})
	if err != nil {
		return nil, err
	}
	return mapSaldoUpdateWithdrawResponse(resp.Data), nil
}

func (a *saldoGRPCAdapter) DebitSaldo(ctx context.Context, request *requests.DebitSaldoRequest) (*models.SaldoMutationResult, error) {
	var resp *pbsaldo.ApiResponseSaldo
	err := a.guard.Call(ctx, func(callCtx context.Context) error {
		var callErr error
		resp, callErr = a.CommandClient.DebitSaldo(callCtx, &pbsaldo.DebitSaldoRequest{
			CardNumber:  request.CardNumber,
			Amount:      int32(request.Amount),
			OperationId: request.OperationID,
			SourceType:  request.SourceType,
			SourceId:    request.SourceID,
		})
		return callErr
	})
	if err != nil {
		return nil, err
	}
	if resp == nil {
		return nil, sharedErrors.NewBadRequestError("saldo response is required")
	}
	return mapSaldoMutationResponse(resp.Data), nil
}

func (a *saldoGRPCAdapter) CreditSaldo(ctx context.Context, request *requests.CreditSaldoRequest) (*models.SaldoMutationResult, error) {
	var resp *pbsaldo.ApiResponseSaldo
	err := a.guard.Call(ctx, func(callCtx context.Context) error {
		var callErr error
		resp, callErr = a.CommandClient.CreditSaldo(callCtx, &pbsaldo.CreditSaldoRequest{
			CardNumber:  request.CardNumber,
			Amount:      int32(request.Amount),
			OperationId: request.OperationID,
			SourceType:  request.SourceType,
			SourceId:    request.SourceID,
		})
		return callErr
	})
	if err != nil {
		return nil, err
	}
	if resp == nil {
		return nil, sharedErrors.NewBadRequestError("saldo response is required")
	}
	return mapSaldoMutationResponse(resp.Data), nil
}
