package adapter

import (
	"context"
	"fmt"

	pbcard "github.com/MamangRust/monolith-payment-gateway-pb/card"
	"github.com/MamangRust/monolith-payment-gateway-pkg/database/models"
	"github.com/MamangRust/monolith-payment-gateway-pkg/resilience"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type CardAdapter interface {
	FindCardByUserId(ctx context.Context, user_id int) (*models.CardAllFieldsRow, error)
	FindUserCardByCardNumber(ctx context.Context, card_number string) (*models.CardByEmailRow, error)
	FindCardByCardNumber(ctx context.Context, card_number string) (*models.CardAllFieldsRow, error)
	UpdateCard(ctx context.Context, request *requests.UpdateCardRequest) (*models.CardUpdateRow, error)
}

type cardGRPCAdapter struct {
	QueryClient   pbcard.CardQueryServiceClient
	CommandClient pbcard.CardCommandServiceClient
	guard         *resilience.DependencyGuard
}

func (a *cardGRPCAdapter) setGuard(g *resilience.DependencyGuard) {
	a.guard = g
}

func NewCardAdapter(queryClient pbcard.CardQueryServiceClient, commandClient pbcard.CardCommandServiceClient, opts ...GuardOption) CardAdapter {
	a := &cardGRPCAdapter{
		QueryClient:   queryClient,
		CommandClient: commandClient,
	}
	for _, opt := range opts {
		opt(a)
	}
	return a
}

func mapCardResponse(resp *pbcard.CardResponse) *models.CardAllFieldsRow {
	if resp == nil {
		return nil
	}
	return &models.CardAllFieldsRow{
		CardID:             resp.Id,
		UserID:             resp.UserId,
		CardNumber:         resp.CardNumber,
		CardType:           resp.CardType,
		ExpireDate:         parseTime(resp.ExpireDate),
		Cvv:                resp.Cvv,
		CardProvider:       resp.CardProvider,
		Status:             resp.Status,
		CreditLimit:        resp.CreditLimit,
		OutstandingBalance: resp.OutstandingBalance,
		RewardPoints:       resp.RewardPoints,
		CreatedAt:          parseTime(resp.CreatedAt),
		UpdatedAt:          parseTime(resp.UpdatedAt),
	}
}

func (a *cardGRPCAdapter) FindCardByUserId(ctx context.Context, user_id int) (*models.CardAllFieldsRow, error) {
	var resp *pbcard.ApiResponseCard
	err := a.guard.Call(ctx, func(callCtx context.Context) error {
		var callErr error
		resp, callErr = a.QueryClient.FindByUserIdCard(callCtx, &pbcard.FindByUserIdCardRequest{
			UserId: int32(user_id),
		})
		return callErr
	})
	if err != nil {
		return nil, err
	}
	return mapCardResponse(resp.Data), nil
}

func (a *cardGRPCAdapter) FindUserCardByCardNumber(ctx context.Context, card_number string) (*models.CardByEmailRow, error) {
	var resp *pbcard.CardWithEmailResponse
	err := a.guard.Call(ctx, func(callCtx context.Context) error {
		var callErr error
		resp, callErr = a.QueryClient.FindUserCardByCardNumber(callCtx, &pbcard.FindByCardNumberRequest{
			CardNumber: card_number,
		})
		return callErr
	})
	if err != nil {
		return nil, err
	}
	if resp == nil {
		return nil, fmt.Errorf("card lookup returned an empty response for card_number=%q", card_number)
	}

	return &models.CardByEmailRow{
		CardID:       resp.Id,
		Email:        resp.Email,
		UserID:       resp.UserId,
		CardNumber:   resp.CardNumber,
		CardType:     resp.CardType,
		ExpireDate:   parseTime(resp.ExpireDate),
		Cvv:          resp.Cvv,
		CardProvider: resp.CardProvider,
		CreatedAt:    parseTime(resp.CreatedAt),
		UpdatedAt:    parseTime(resp.UpdatedAt),
	}, nil
}

func (a *cardGRPCAdapter) FindCardByCardNumber(ctx context.Context, card_number string) (*models.CardAllFieldsRow, error) {
	var resp *pbcard.ApiResponseCard
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
	return mapCardResponse(resp.Data), nil
}

func (a *cardGRPCAdapter) UpdateCard(ctx context.Context, request *requests.UpdateCardRequest) (*models.CardUpdateRow, error) {
	var resp *pbcard.ApiResponseCard
	err := a.guard.Call(ctx, func(callCtx context.Context) error {
		var callErr error
		resp, callErr = a.CommandClient.UpdateCard(callCtx, &pbcard.UpdateCardRequest{
			CardId:       int32(request.CardID),
			UserId:       int32(request.UserID),
			CardType:     request.CardType,
			ExpireDate:   timestamppb.New(request.ExpireDate),
			Cvv:          request.CVV,
			CardProvider: request.CardProvider,
		})
		return callErr
	})
	if err != nil {
		return nil, err
	}
	d := resp.Data
	return &models.CardUpdateRow{
		CardID:       d.Id,
		UserID:       d.UserId,
		CardNumber:   d.CardNumber,
		CardType:     d.CardType,
		ExpireDate:   parseTime(d.ExpireDate),
		Cvv:          d.Cvv,
		CardProvider: d.CardProvider,
		CreatedAt:    parseTime(d.CreatedAt),
		UpdatedAt:    parseTime(d.UpdatedAt),
	}, nil
}
