package withdraw_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	card_repo_impl "github.com/MamangRust/monolith-payment-gateway-card/repository"
	models "github.com/MamangRust/monolith-payment-gateway-pkg/database/models"
	"github.com/MamangRust/monolith-payment-gateway-pkg/logger"
	saldo_repo_impl "github.com/MamangRust/monolith-payment-gateway-saldo/repository"
	"github.com/MamangRust/monolith-payment-gateway-shared/cache"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
	"github.com/MamangRust/monolith-payment-gateway-shared/observability"
	tests "github.com/MamangRust/monolith-payment-gateway-test"
	user_repo_impl "github.com/MamangRust/monolith-payment-gateway-user/repository"
	"github.com/MamangRust/monolith-payment-gateway-withdraw/repository"
	"github.com/MamangRust/monolith-payment-gateway-withdraw/service"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/suite"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type WithdrawServiceTestSuite struct {
	suite.Suite
	ts              *tests.TestSuite
	withdrawService service.Service
	gormDB          *gorm.DB
	withdrawID      int
	cardNumber      string
	userID          int
}

func (s *WithdrawServiceTestSuite) SetupSuite() {
	ts, err := tests.SetupTestSuite()
	s.Require().NoError(err)
	s.ts = ts

	opts, err := redis.ParseURL(s.ts.RedisURL)
	s.Require().NoError(err)
	redisClient := redis.NewClient(opts)

	gormDB, gormErr := gorm.Open(postgres.Open(s.ts.DBURL), &gorm.Config{})
	if gormErr != nil {
		s.Require().NoError(gormErr)
	}
	s.gormDB = gormDB

	repos := repository.NewRepositories(gormDB, tests.NewCardQueryClient(gormDB), tests.NewCardCommandClient(gormDB), tests.NewSaldoQueryClient(gormDB), tests.NewSaldoCommandClient(gormDB), 0)

	logger.ResetInstance()
	lp := sdklog.NewLoggerProvider()
	log, _ := logger.NewLogger("test", lp)
	cacheMetrics, _ := observability.NewCacheMetrics("test")
	cacheStore := cache.NewCacheStore(redisClient, log, cacheMetrics)

	s.withdrawService = service.NewService(&service.Deps{
		Kafka:        nil,
		Repositories: repos,
		Logger:       log,
		Cache:        cacheStore,
	})

	// Seed User
	userRepo := user_repo_impl.NewUserCommandRepository(gormDB)
	user, err := userRepo.CreateUser(context.Background(), &requests.CreateUserRequest{
		FirstName: "Withdraw",
		LastName:  "Tester",
		Email:     fmt.Sprintf("withdraw.tester.%d@example.com", time.Now().UnixNano()),
		Password:  "password123",
	})
	s.Require().NoError(err)
	s.userID = int(user.UserID)

	// Seed Card
	cardCmdRepo := card_repo_impl.NewCardCommandRepository(gormDB)
	card, err := cardCmdRepo.CreateCard(context.Background(), &requests.CreateCardRequest{
		UserID:       s.userID,
		CardType:     "debit",
		ExpireDate:   time.Now().AddDate(5, 0, 0),
		CVV:          "123",
		CardProvider: "visa",
	})
	s.Require().NoError(err)
	s.cardNumber = card.CardNumber

	// Seed Saldo
	saldoCmdRepo := saldo_repo_impl.NewSaldoCommandRepository(gormDB)
	_, err = saldoCmdRepo.CreateSaldo(context.Background(), &requests.CreateSaldoRequest{
		CardNumber:   s.cardNumber,
		TotalBalance: 1000000,
	})
	s.Require().NoError(err)
}

func (s *WithdrawServiceTestSuite) TearDownSuite() {
	s.ts.Teardown()
}

func (s *WithdrawServiceTestSuite) getSaldoBalance(ctx context.Context, cardNumber string) int32 {
	var saldo models.Saldo
	err := s.gormDB.WithContext(ctx).Raw(`SELECT saldo_id, card_number, total_balance FROM saldos WHERE card_number = ? AND deleted_at IS NULL`, cardNumber).Scan(&saldo).Error
	s.Require().NoError(err)
	return saldo.TotalBalance
}

func (s *WithdrawServiceTestSuite) Test1_WithdrawLifecycle() {
	ctx := context.Background()

	// Create Withdraw
	createReq := &requests.CreateWithdrawRequest{
		CardNumber:     s.cardNumber,
		WithdrawAmount: 100000,
		WithdrawTime:   time.Now(),
	}
	res, err := s.withdrawService.Create(ctx, createReq)
	s.NoError(err)
	s.NotNil(res)
	s.withdrawID = int(res.WithdrawID)
	s.Equal("success", res.Status)

	// FindById
	found, err := s.withdrawService.FindById(ctx, s.withdrawID)
	s.NoError(err)
	s.NotNil(found)
	s.Equal(int32(s.withdrawID), found.WithdrawID)

	// Update Withdraw
	updateReq := &requests.UpdateWithdrawRequest{
		WithdrawID:     &s.withdrawID,
		CardNumber:     s.cardNumber,
		WithdrawAmount: 200000,
		WithdrawTime:   time.Now(),
	}
	updated, err := s.withdrawService.Update(ctx, updateReq)
	s.NoError(err)
	s.NotNil(updated)
	s.Equal(int32(200000), updated.WithdrawAmount)
}

func (s *WithdrawServiceTestSuite) Test2_QueryOperations() {
	ctx := context.Background()

	// FindAll
	all, total, err := s.withdrawService.FindAll(ctx, &requests.FindAllWithdraws{
		Page:     1,
		PageSize: 10,
	})
	s.NoError(err)
	s.NotNil(all)
	s.GreaterOrEqual(*total, 1)

	// FindByActive
	active, totalActive, err := s.withdrawService.FindByActive(ctx, &requests.FindAllWithdraws{
		Page:     1,
		PageSize: 10,
	})
	s.NoError(err)
	s.NotNil(active)
	s.GreaterOrEqual(*totalActive, 1)

	// FindByCardNumber
	byCard, totalCard, err := s.withdrawService.FindAllByCardNumber(ctx, &requests.FindAllWithdrawCardNumber{
		CardNumber: s.cardNumber,
		Page:       1,
		PageSize:   10,
	})
	s.NoError(err)
	s.NotNil(byCard)
	s.GreaterOrEqual(*totalCard, 1)
}

func (s *WithdrawServiceTestSuite) Test3_TrashAndRestore() {
	ctx := context.Background()
	s.Require().NotZero(s.withdrawID)

	// Trash
	trashed, err := s.withdrawService.TrashedWithdraw(ctx, s.withdrawID)
	s.NoError(err)
	s.NotNil(trashed)

	// FindByTrashed
	trashedList, totalTrashed, err := s.withdrawService.FindByTrashed(ctx, &requests.FindAllWithdraws{
		Page:     1,
		PageSize: 10,
	})
	s.NoError(err)
	s.NotNil(trashedList)
	s.GreaterOrEqual(*totalTrashed, 1)

	// Restore
	restored, err := s.withdrawService.RestoreWithdraw(ctx, s.withdrawID)
	s.NoError(err)
	s.NotNil(restored)
}

func (s *WithdrawServiceTestSuite) Test4_BulkOperations() {
	ctx := context.Background()

	// Restore All
	ok, err := s.withdrawService.RestoreAllWithdraw(ctx)
	s.NoError(err)
	s.True(ok)

	// Delete All Permanent
	ok, err = s.withdrawService.DeleteAllWithdrawPermanent(ctx)
	s.NoError(err)
	s.True(ok)
}

func (s *WithdrawServiceTestSuite) Test5_WithdrawIdempotentReplay() {
	ctx := context.Background()
	key := fmt.Sprintf("withdraw:create:user:%d:op:%d", s.userID, time.Now().UnixNano())
	amount := 50000

	before := s.getSaldoBalance(ctx, s.cardNumber)

	req := &requests.CreateWithdrawRequest{
		CardNumber:     s.cardNumber,
		WithdrawAmount: amount,
		WithdrawTime:   time.Now(),
		IdempotencyKey: key,
	}

	// A replay with the same key must return the original withdrawal and must
	// NOT debit the balance a second time.
	first, err := s.withdrawService.Create(ctx, req)
	s.Require().NoError(err)
	s.Require().NotNil(first)

	second, err := s.withdrawService.Create(ctx, req)
	s.Require().NoError(err)
	s.Require().NotNil(second)
	s.Equal(first.WithdrawID, second.WithdrawID, "replay must return the original withdrawal")

	after := s.getSaldoBalance(ctx, s.cardNumber)
	s.Equal(before-int32(amount), after, "balance must be debited exactly once")
}

func (s *WithdrawServiceTestSuite) TestZZ_NotFound404() {
	ctx := context.Background()
	_, err := s.withdrawService.FindById(ctx, 999999999)
	s.Require().Error(err)
	s.Contains(err.Error(), "not found")
}

func (s *WithdrawServiceTestSuite) Test6_FindByIdNotFound() {
	ctx := context.Background()
	_, err := s.withdrawService.FindById(ctx, 999999999)
	s.Require().Error(err, "a non-existent withdraw must not resolve")
	s.Contains(err.Error(), "not found", "missing withdraw must surface as a 404-style not-found error")
}

func TestWithdrawServiceSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(WithdrawServiceTestSuite))
}
