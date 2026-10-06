package tests

import (
	"context"
	"time"

	pbcard "github.com/MamangRust/monolith-payment-gateway-pb/card"
	pbmerchant "github.com/MamangRust/monolith-payment-gateway-pb/merchant"
	pbsaldo "github.com/MamangRust/monolith-payment-gateway-pb/saldo"
	cardrepo "github.com/MamangRust/monolith-payment-gateway-card/repository"
	merchantrepo "github.com/MamangRust/monolith-payment-gateway-merchant/repository"
	saldorepo "github.com/MamangRust/monolith-payment-gateway-saldo/repository"
	"github.com/MamangRust/monolith-payment-gateway-pkg/database/models"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
	"gorm.io/gorm"
)

// These GORM-backed gRPC client shims let integration suites that exercise the
// topup/transaction/transfer/withdraw repositories build the real
// adapter.CardAdapter / adapter.SaldoAdapter / adapter.MerchantAdapter without a
// network hop. They wrap the production repositories and translate the adapter
// request/response types to/from the protobuf messages, exactly like the real
// gRPC services do. Methods the adapters never call are stubbed with
// errShimNotImplemented.

func formatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(time.RFC3339)
}

func cardAllFieldsToPB(c *models.CardAllFieldsRow) *pbcard.CardResponse {
	if c == nil {
		return nil
	}
	return &pbcard.CardResponse{
		Id:                 c.CardID,
		UserId:             c.UserID,
		CardNumber:         c.CardNumber,
		CardType:           c.CardType,
		ExpireDate:         formatTime(c.ExpireDate),
		Cvv:                c.Cvv,
		CardProvider:       c.CardProvider,
		CreatedAt:          formatTime(c.CreatedAt),
		UpdatedAt:          formatTime(c.UpdatedAt),
		Status:             c.Status,
		CreditLimit:        c.CreditLimit,
		OutstandingBalance: c.OutstandingBalance,
		RewardPoints:       c.RewardPoints,
	}
}

func cardByEmailToPB(c *models.CardByEmailRow) *pbcard.CardWithEmailResponse {
	if c == nil {
		return nil
	}
	return &pbcard.CardWithEmailResponse{
		Id:           c.CardID,
		UserId:       c.UserID,
		Email:        c.Email,
		CardNumber:   c.CardNumber,
		CardType:     c.CardType,
		ExpireDate:   formatTime(c.ExpireDate),
		Cvv:          c.Cvv,
		CardProvider: c.CardProvider,
		CreatedAt:    formatTime(c.CreatedAt),
		UpdatedAt:    formatTime(c.UpdatedAt),
	}
}

func cardUpdateToPB(c *models.CardUpdateRow) *pbcard.CardResponse {
	if c == nil {
		return nil
	}
	return &pbcard.CardResponse{
		Id:           c.CardID,
		UserId:       c.UserID,
		CardNumber:   c.CardNumber,
		CardType:     c.CardType,
		ExpireDate:   formatTime(c.ExpireDate),
		Cvv:          c.Cvv,
		CardProvider: c.CardProvider,
		CreatedAt:    formatTime(c.CreatedAt),
		UpdatedAt:    formatTime(c.UpdatedAt),
	}
}

func saldoRowToPB(saldoID int32, cardNumber string, totalBalance int32, withdrawAmount *int32, withdrawTime *time.Time, createdAt, updatedAt time.Time) *pbsaldo.SaldoResponse {
	var wa int32
	if withdrawAmount != nil {
		wa = *withdrawAmount
	}
	return &pbsaldo.SaldoResponse{
		SaldoId:        saldoID,
		CardNumber:     cardNumber,
		TotalBalance:   totalBalance,
		WithdrawAmount: wa,
		WithdrawTime:   formatTimeT(withdrawTime),
		CreatedAt:      formatTime(createdAt),
		UpdatedAt:      formatTime(updatedAt),
	}
}

func formatTimeT(t *time.Time) string {
	if t == nil {
		return ""
	}
	return formatTime(*t)
}

func merchantToPB(m *models.MerchantAllFieldsRow) *pbmerchant.MerchantResponse {
	if m == nil {
		return nil
	}
	return &pbmerchant.MerchantResponse{
		Id:        m.MerchantID,
		Name:      m.Name,
		ApiKey:    m.ApiKey,
		Status:    m.Status,
		UserId:    m.UserID,
		CreatedAt: formatTime(m.CreatedAt),
		UpdatedAt: formatTime(m.UpdatedAt),
	}
}

// --- Card query ---

type gormCardQueryClient struct {
	repo cardrepo.CardQueryRepository
}

func NewCardQueryClient(db *gorm.DB) pbcard.CardQueryServiceClient {
	return &gormCardQueryClient{repo: cardrepo.NewCardQueryRepository(db)}
}

func (c *gormCardQueryClient) FindAllCard(context.Context, *pbcard.FindAllCardRequest, ...grpc.CallOption) (*pbcard.ApiResponsePaginationCard, error) {
	return nil, errShimNotImplemented
}
func (c *gormCardQueryClient) FindByIdCard(context.Context, *pbcard.FindByIdCardRequest, ...grpc.CallOption) (*pbcard.ApiResponseCard, error) {
	return nil, errShimNotImplemented
}
func (c *gormCardQueryClient) FindByUserIdCard(ctx context.Context, in *pbcard.FindByUserIdCardRequest, _ ...grpc.CallOption) (*pbcard.ApiResponseCard, error) {
	row, err := c.repo.FindCardByUserId(ctx, int(in.UserId))
	if err != nil {
		return nil, err
	}
	return &pbcard.ApiResponseCard{Data: cardAllFieldsToPB(row)}, nil
}
func (c *gormCardQueryClient) FindByActiveCard(context.Context, *pbcard.FindAllCardRequest, ...grpc.CallOption) (*pbcard.ApiResponsePaginationCardDeleteAt, error) {
	return nil, errShimNotImplemented
}
func (c *gormCardQueryClient) FindByTrashedCard(context.Context, *pbcard.FindAllCardRequest, ...grpc.CallOption) (*pbcard.ApiResponsePaginationCardDeleteAt, error) {
	return nil, errShimNotImplemented
}
func (c *gormCardQueryClient) FindByCardNumber(ctx context.Context, in *pbcard.FindByCardNumberRequest, _ ...grpc.CallOption) (*pbcard.ApiResponseCard, error) {
	row, err := c.repo.FindCardByCardNumber(ctx, in.CardNumber)
	if err != nil {
		return nil, err
	}
	return &pbcard.ApiResponseCard{Data: cardAllFieldsToPB(row)}, nil
}
func (c *gormCardQueryClient) FindUserCardByCardNumber(ctx context.Context, in *pbcard.FindByCardNumberRequest, _ ...grpc.CallOption) (*pbcard.CardWithEmailResponse, error) {
	row, err := c.repo.FindUserCardByCardNumber(ctx, in.CardNumber)
	if err != nil {
		return nil, err
	}
	return cardByEmailToPB(row), nil
}

// --- Card command ---

type gormCardCommandClient struct {
	repo cardrepo.CardCommandRepository
}

func NewCardCommandClient(db *gorm.DB) pbcard.CardCommandServiceClient {
	return &gormCardCommandClient{repo: cardrepo.NewCardCommandRepository(db)}
}

func (c *gormCardCommandClient) CreateCard(context.Context, *pbcard.CreateCardRequest, ...grpc.CallOption) (*pbcard.ApiResponseCard, error) {
	return nil, errShimNotImplemented
}
func (c *gormCardCommandClient) UpdateCard(ctx context.Context, in *pbcard.UpdateCardRequest, _ ...grpc.CallOption) (*pbcard.ApiResponseCard, error) {
	var expire time.Time
	if in.ExpireDate != nil {
		expire = in.ExpireDate.AsTime()
	}
	row, err := c.repo.UpdateCard(ctx, &requests.UpdateCardRequest{
		CardID:       int(in.CardId),
		UserID:       int(in.UserId),
		CardType:     in.CardType,
		ExpireDate:   expire,
		CVV:          in.Cvv,
		CardProvider: in.CardProvider,
	})
	if err != nil {
		return nil, err
	}
	return &pbcard.ApiResponseCard{Data: cardUpdateToPB(row)}, nil
}
func (c *gormCardCommandClient) TrashedCard(context.Context, *pbcard.FindByIdCardRequest, ...grpc.CallOption) (*pbcard.ApiResponseCardDeleteAt, error) {
	return nil, errShimNotImplemented
}
func (c *gormCardCommandClient) RestoreCard(context.Context, *pbcard.FindByIdCardRequest, ...grpc.CallOption) (*pbcard.ApiResponseCardDeleteAt, error) {
	return nil, errShimNotImplemented
}
func (c *gormCardCommandClient) DeleteCardPermanent(context.Context, *pbcard.FindByIdCardRequest, ...grpc.CallOption) (*pbcard.ApiResponseCardDelete, error) {
	return nil, errShimNotImplemented
}
func (c *gormCardCommandClient) RestoreAllCard(context.Context, *emptypb.Empty, ...grpc.CallOption) (*pbcard.ApiResponseCardAll, error) {
	return nil, errShimNotImplemented
}
func (c *gormCardCommandClient) DeleteAllCardPermanent(context.Context, *emptypb.Empty, ...grpc.CallOption) (*pbcard.ApiResponseCardAll, error) {
	return nil, errShimNotImplemented
}
func (c *gormCardCommandClient) ToggleCardStatus(context.Context, *pbcard.ToggleCardStatusRequest, ...grpc.CallOption) (*pbcard.ApiResponseCard, error) {
	return nil, errShimNotImplemented
}
func (c *gormCardCommandClient) UpdateCreditLimit(context.Context, *pbcard.UpdateCreditLimitRequest, ...grpc.CallOption) (*pbcard.ApiResponseCard, error) {
	return nil, errShimNotImplemented
}
func (c *gormCardCommandClient) RedeemPoints(context.Context, *pbcard.RedeemPointsRequest, ...grpc.CallOption) (*pbcard.ApiResponseCard, error) {
	return nil, errShimNotImplemented
}
func (c *gormCardCommandClient) ProcessBillingCycles(context.Context, *emptypb.Empty, ...grpc.CallOption) (*emptypb.Empty, error) {
	return nil, errShimNotImplemented
}

// --- Saldo query ---

type gormSaldoQueryClient struct {
	repo saldorepo.SaldoQueryRepository
}

func NewSaldoQueryClient(db *gorm.DB) pbsaldo.SaldoQueryServiceClient {
	return &gormSaldoQueryClient{repo: saldorepo.NewSaldoQueryRepository(db)}
}

func (c *gormSaldoQueryClient) FindAllSaldo(context.Context, *pbsaldo.FindAllSaldoRequest, ...grpc.CallOption) (*pbsaldo.ApiResponsePaginationSaldo, error) {
	return nil, errShimNotImplemented
}
func (c *gormSaldoQueryClient) FindByIdSaldo(context.Context, *pbsaldo.FindByIdSaldoRequest, ...grpc.CallOption) (*pbsaldo.ApiResponseSaldo, error) {
	return nil, errShimNotImplemented
}
func (c *gormSaldoQueryClient) FindByCardNumber(ctx context.Context, in *pbcard.FindByCardNumberRequest, _ ...grpc.CallOption) (*pbsaldo.ApiResponseSaldo, error) {
	s, err := c.repo.FindByCardNumber(ctx, in.CardNumber)
	if err != nil {
		return nil, err
	}
	return &pbsaldo.ApiResponseSaldo{Data: saldoRowToPB(s.SaldoID, s.CardNumber, s.TotalBalance, s.WithdrawAmount, s.WithdrawTime, s.CreatedAt, s.UpdatedAt)}, nil
}
func (c *gormSaldoQueryClient) FindByActive(context.Context, *pbsaldo.FindAllSaldoRequest, ...grpc.CallOption) (*pbsaldo.ApiResponsePaginationSaldoDeleteAt, error) {
	return nil, errShimNotImplemented
}
func (c *gormSaldoQueryClient) FindByTrashed(context.Context, *pbsaldo.FindAllSaldoRequest, ...grpc.CallOption) (*pbsaldo.ApiResponsePaginationSaldoDeleteAt, error) {
	return nil, errShimNotImplemented
}

// --- Saldo command ---

type gormSaldoCommandClient struct {
	repo saldorepo.SaldoCommandRepository
}

func NewSaldoCommandClient(db *gorm.DB) pbsaldo.SaldoCommandServiceClient {
	return &gormSaldoCommandClient{repo: saldorepo.NewSaldoCommandRepository(db)}
}

func (c *gormSaldoCommandClient) CreateSaldo(context.Context, *pbsaldo.CreateSaldoRequest, ...grpc.CallOption) (*pbsaldo.ApiResponseSaldo, error) {
	return nil, errShimNotImplemented
}
func (c *gormSaldoCommandClient) UpdateSaldo(context.Context, *pbsaldo.UpdateSaldoRequest, ...grpc.CallOption) (*pbsaldo.ApiResponseSaldo, error) {
	return nil, errShimNotImplemented
}
func (c *gormSaldoCommandClient) UpdateSaldoBalance(ctx context.Context, in *pbsaldo.UpdateSaldoBalanceRequest, _ ...grpc.CallOption) (*pbsaldo.ApiResponseSaldo, error) {
	row, err := c.repo.UpdateSaldoBalance(ctx, &requests.UpdateSaldoBalance{
		CardNumber:   in.CardNumber,
		TotalBalance: int(in.TotalBalance),
	})
	if err != nil {
		return nil, err
	}
	return &pbsaldo.ApiResponseSaldo{Data: saldoRowToPB(row.SaldoID, row.CardNumber, row.TotalBalance, row.WithdrawAmount, row.WithdrawTime, row.CreatedAt, row.UpdatedAt)}, nil
}
func (c *gormSaldoCommandClient) UpdateSaldoWithdraw(ctx context.Context, in *pbsaldo.UpdateSaldoWithdrawRequest, _ ...grpc.CallOption) (*pbsaldo.ApiResponseSaldo, error) {
	var wa int
	if in.WithdrawAmount != 0 {
		wa = int(in.WithdrawAmount)
	}
	var wt *time.Time
	if t, err := time.Parse(time.RFC3339, in.WithdrawTime); err == nil {
		wt = &t
	}
	row, err := c.repo.UpdateSaldoWithdraw(ctx, &requests.UpdateSaldoWithdraw{
		CardNumber:     in.CardNumber,
		TotalBalance:   int(in.TotalBalance),
		WithdrawAmount: &wa,
		WithdrawTime:   wt,
	})
	if err != nil {
		return nil, err
	}
	return &pbsaldo.ApiResponseSaldo{Data: saldoRowToPB(row.SaldoID, row.CardNumber, row.TotalBalance, row.WithdrawAmount, row.WithdrawTime, row.CreatedAt, row.UpdatedAt)}, nil
}
func (c *gormSaldoCommandClient) TrashedSaldo(context.Context, *pbsaldo.FindByIdSaldoRequest, ...grpc.CallOption) (*pbsaldo.ApiResponseSaldoDeleteAt, error) {
	return nil, errShimNotImplemented
}
func (c *gormSaldoCommandClient) RestoreSaldo(context.Context, *pbsaldo.FindByIdSaldoRequest, ...grpc.CallOption) (*pbsaldo.ApiResponseSaldoDeleteAt, error) {
	return nil, errShimNotImplemented
}
func (c *gormSaldoCommandClient) DeleteSaldoPermanent(context.Context, *pbsaldo.FindByIdSaldoRequest, ...grpc.CallOption) (*pbsaldo.ApiResponseSaldoDelete, error) {
	return nil, errShimNotImplemented
}
func (c *gormSaldoCommandClient) RestoreAllSaldo(context.Context, *emptypb.Empty, ...grpc.CallOption) (*pbsaldo.ApiResponseSaldoAll, error) {
	return nil, errShimNotImplemented
}
func (c *gormSaldoCommandClient) DeleteAllSaldoPermanent(context.Context, *emptypb.Empty, ...grpc.CallOption) (*pbsaldo.ApiResponseSaldoAll, error) {
	return nil, errShimNotImplemented
}
func (c *gormSaldoCommandClient) DebitSaldo(context.Context, *pbsaldo.DebitSaldoRequest, ...grpc.CallOption) (*pbsaldo.ApiResponseSaldo, error) {
	return nil, errShimNotImplemented
}
func (c *gormSaldoCommandClient) CreditSaldo(context.Context, *pbsaldo.CreditSaldoRequest, ...grpc.CallOption) (*pbsaldo.ApiResponseSaldo, error) {
	return nil, errShimNotImplemented
}

// --- Merchant query ---

type gormMerchantQueryClient struct {
	repo merchantrepo.MerchantQueryRepository
}

func NewMerchantQueryClient(db *gorm.DB) pbmerchant.MerchantQueryServiceClient {
	return &gormMerchantQueryClient{repo: merchantrepo.NewMerchantQueryRepository(db)}
}

func (c *gormMerchantQueryClient) FindAllMerchant(context.Context, *pbmerchant.FindAllMerchantRequest, ...grpc.CallOption) (*pbmerchant.ApiResponsePaginationMerchant, error) {
	return nil, errShimNotImplemented
}
func (c *gormMerchantQueryClient) FindByIdMerchant(context.Context, *pbmerchant.FindByIdMerchantRequest, ...grpc.CallOption) (*pbmerchant.ApiResponseMerchant, error) {
	return nil, errShimNotImplemented
}
func (c *gormMerchantQueryClient) FindByApiKey(ctx context.Context, in *pbmerchant.FindByApiKeyRequest, _ ...grpc.CallOption) (*pbmerchant.ApiResponseMerchant, error) {
	row, err := c.repo.FindByApiKey(ctx, in.ApiKey)
	if err != nil {
		return nil, err
	}
	return &pbmerchant.ApiResponseMerchant{Data: merchantToPB(row)}, nil
}
func (c *gormMerchantQueryClient) FindByMerchantUserId(context.Context, *pbmerchant.FindByMerchantUserIdRequest, ...grpc.CallOption) (*pbmerchant.ApiResponsesMerchant, error) {
	return nil, errShimNotImplemented
}
func (c *gormMerchantQueryClient) FindByActive(context.Context, *pbmerchant.FindAllMerchantRequest, ...grpc.CallOption) (*pbmerchant.ApiResponsePaginationMerchantDeleteAt, error) {
	return nil, errShimNotImplemented
}
func (c *gormMerchantQueryClient) FindByTrashed(context.Context, *pbmerchant.FindAllMerchantRequest, ...grpc.CallOption) (*pbmerchant.ApiResponsePaginationMerchantDeleteAt, error) {
	return nil, errShimNotImplemented
}
