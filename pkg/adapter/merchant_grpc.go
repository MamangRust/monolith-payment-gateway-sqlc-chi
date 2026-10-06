package adapter

import (
	"context"

	pbmerchant "github.com/MamangRust/monolith-payment-gateway-pb/merchant"
	"github.com/MamangRust/monolith-payment-gateway-pkg/database/models"
	"github.com/MamangRust/monolith-payment-gateway-pkg/resilience"
)

type MerchantAdapter interface {
	FindByApiKey(ctx context.Context, api_key string) (*models.MerchantAllFieldsRow, error)
	FindByMerchantId(ctx context.Context, merchant_id int) (*models.MerchantAllFieldsRow, error)
}

type merchantGRPCAdapter struct {
	QueryClient pbmerchant.MerchantQueryServiceClient
	guard       *resilience.DependencyGuard
}

func (a *merchantGRPCAdapter) setGuard(g *resilience.DependencyGuard) {
	a.guard = g
}

func NewMerchantAdapter(queryClient pbmerchant.MerchantQueryServiceClient, opts ...GuardOption) MerchantAdapter {
	a := &merchantGRPCAdapter{
		QueryClient: queryClient,
	}
	for _, opt := range opts {
		opt(a)
	}
	return a
}

func mapMerchantResponse(resp *pbmerchant.MerchantResponse) *models.MerchantAllFieldsRow {
	if resp == nil {
		return nil
	}
	return &models.MerchantAllFieldsRow{
		MerchantID: resp.Id,
		Name:       resp.Name,
		ApiKey:     resp.ApiKey,
		UserID:     resp.UserId,
		Status:     resp.Status,
		CreatedAt:  parseTime(resp.CreatedAt),
		UpdatedAt:  parseTime(resp.UpdatedAt),
	}
}

func (a *merchantGRPCAdapter) FindByApiKey(ctx context.Context, api_key string) (*models.MerchantAllFieldsRow, error) {
	var resp *pbmerchant.ApiResponseMerchant
	err := a.guard.Call(ctx, func(callCtx context.Context) error {
		var callErr error
		resp, callErr = a.QueryClient.FindByApiKey(callCtx, &pbmerchant.FindByApiKeyRequest{
			ApiKey: api_key,
		})
		return callErr
	})
	if err != nil {
		return nil, err
	}
	return mapMerchantResponse(resp.Data), nil
}

func (a *merchantGRPCAdapter) FindByMerchantId(ctx context.Context, merchant_id int) (*models.MerchantAllFieldsRow, error) {
	var resp *pbmerchant.ApiResponseMerchant
	err := a.guard.Call(ctx, func(callCtx context.Context) error {
		var callErr error
		resp, callErr = a.QueryClient.FindByIdMerchant(callCtx, &pbmerchant.FindByIdMerchantRequest{
			MerchantId: int32(merchant_id),
		})
		return callErr
	})
	if err != nil {
		return nil, err
	}
	return mapMerchantResponse(resp.Data), nil
}
