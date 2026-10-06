package topup_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	card_repo "github.com/MamangRust/monolith-payment-gateway-card/repository"
	"github.com/MamangRust/monolith-payment-gateway-pkg/database/models"
	"github.com/MamangRust/monolith-payment-gateway-pkg/logger"
	"github.com/MamangRust/monolith-payment-gateway-shared/cache"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
	"github.com/MamangRust/monolith-payment-gateway-shared/observability"
	tests "github.com/MamangRust/monolith-payment-gateway-test"
	"github.com/MamangRust/monolith-payment-gateway-topup/repository"
	user_repo "github.com/MamangRust/monolith-payment-gateway-user/repository"
	pbuser "github.com/MamangRust/monolith-payment-gateway-pb/user"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/suite"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type TopupRepositoryTestSuite struct {
	suite.Suite
	ts       *tests.TestSuite
	repo     repository.Repositories
	cardRepo *card_repo.Repositories
	userRepo user_repo.Repositories
	userID   int
	cleanup  func()
}

func (s *TopupRepositoryTestSuite) SetupSuite() {
	ts, err := tests.SetupTestSuite()
	s.Require().NoError(err)
	s.ts = ts

	gormDB, gormErr := gorm.Open(postgres.Open(s.ts.DBURL), &gorm.Config{})
	if gormErr != nil {
		s.Require().NoError(gormErr)
	}

	opts, err := redis.ParseURL(s.ts.RedisURL)
	s.Require().NoError(err)
	redisClient := redis.NewClient(opts)

	logger.ResetInstance()
	lp := sdklog.NewLoggerProvider()
	log, _ := logger.NewLogger("test", lp)
	cacheMetrics, _ := observability.NewCacheMetrics("test")
	cacheStore := cache.NewCacheStore(redisClient, log, cacheMetrics)

	userConn, userCleanup, err := tests.StartUserService(gormDB, log, cacheStore)
	s.Require().NoError(err)
	s.cleanup = userCleanup

	s.userRepo = user_repo.NewRepositories(&user_repo.Deps{Db: gormDB, Role: tests.NewRoleRepo(gormDB), UserRole: tests.NewUserRoleRepo(gormDB)})
	s.cardRepo = card_repo.NewRepositories(gormDB, pbuser.NewUserQueryServiceClient(userConn))
	// We don't need real adapters for repository integration tests because they are not used in repo methods
	s.repo = repository.NewRepositories(gormDB,
		tests.NewCardQueryClient(gormDB), tests.NewCardCommandClient(gormDB),
		tests.NewSaldoQueryClient(gormDB), tests.NewSaldoCommandClient(gormDB))

	// Create user
	user, err := s.userRepo.UserCommand().CreateUser(context.Background(), &requests.CreateUserRequest{
		FirstName: "Topup",
		LastName:  "Owner",
		Email:     fmt.Sprintf("topup.owner-%d@example.com", time.Now().UnixNano()),
		Password:  "password123",
	})
	s.Require().NoError(err)
	s.userID = int(user.UserID)
}

func (s *TopupRepositoryTestSuite) TearDownSuite() {
	if s.cleanup != nil {
		s.cleanup()
	}
	s.ts.Teardown()
}

func (s *TopupRepositoryTestSuite) createSeedTopup() (*models.TopupAllFieldsRow, error) {
	card, err := s.cardRepo.CardCommand.CreateCard(context.Background(), &requests.CreateCardRequest{
		UserID:       s.userID,
		CardType:     "debit",
		ExpireDate:   time.Now().AddDate(5, 0, 0),
		CVV:          "123",
		CardProvider: "Visa",
	})
	if err != nil {
		return nil, err
	}

	return s.repo.CreateTopup(context.Background(), &requests.CreateTopupRequest{
		CardNumber:  card.CardNumber,
		TopupAmount: 100000,
		TopupMethod: "bank_transfer",
	})
}

func (s *TopupRepositoryTestSuite) TestCreateTopup() {
	ctx := context.Background()

	card, _ := s.cardRepo.CardCommand.CreateCard(ctx, &requests.CreateCardRequest{
		UserID:       s.userID,
		CardType:     "debit",
		ExpireDate:   time.Now().AddDate(5, 0, 0),
		CVV:          "123",
		CardProvider: "Visa",
	})

	req := &requests.CreateTopupRequest{
		CardNumber:  card.CardNumber,
		TopupAmount: 100000,
		TopupMethod: "bank_transfer",
	}

	res, err := s.repo.CreateTopup(ctx, req)
	s.NoError(err)
	s.NotNil(res)
}

func (s *TopupRepositoryTestSuite) TestFindAllTopups() {
	_, err := s.createSeedTopup()
	s.Require().NoError(err)
	ctx := context.Background()

	res, err := s.repo.FindAllTopups(ctx, &requests.FindAllTopups{
		Page:     1,
		PageSize: 10,
		Search:   "",
	})
	s.NoError(err)
	s.GreaterOrEqual(len(res), 1)
}

func (s *TopupRepositoryTestSuite) TestFindById() {
	topup, err := s.createSeedTopup()
	s.Require().NoError(err)
	ctx := context.Background()

	found, err := s.repo.FindById(ctx, int(topup.TopupID))
	s.NoError(err)
	s.NotNil(found)
	s.Equal(topup.TopupID, found.TopupID)
}

func (s *TopupRepositoryTestSuite) TestFindByActive() {
	_, err := s.createSeedTopup()
	s.Require().NoError(err)
	ctx := context.Background()

	res, err := s.repo.FindByActive(ctx, &requests.FindAllTopups{
		Page:     1,
		PageSize: 10,
		Search:   "",
	})
	s.NoError(err)
	s.GreaterOrEqual(len(res), 1)
}

func (s *TopupRepositoryTestSuite) TestFindByTrashed() {
	topup, err := s.createSeedTopup()
	s.Require().NoError(err)
	ctx := context.Background()

	_, err = s.repo.TrashedTopup(ctx, int(topup.TopupID))
	s.Require().NoError(err)

	res, err := s.repo.FindByTrashed(ctx, &requests.FindAllTopups{
		Page:     1,
		PageSize: 10,
		Search:   "",
	})
	s.NoError(err)
	s.GreaterOrEqual(len(res), 1)
}

func (s *TopupRepositoryTestSuite) TestUpdateTopup() {
	topup, err := s.createSeedTopup()
	s.Require().NoError(err)
	ctx := context.Background()

	id := int(topup.TopupID)
	req := &requests.UpdateTopupRequest{
		TopupID:     &id,
		CardNumber:  topup.CardNumber,
		TopupAmount: 200000,
		TopupMethod: "bank_transfer",
	}

	res, err := s.repo.UpdateTopup(ctx, req)
	s.NoError(err)
	s.NotNil(res)
}

func (s *TopupRepositoryTestSuite) TestTrashTopup() {
	topup, err := s.createSeedTopup()
	s.Require().NoError(err)
	ctx := context.Background()

	trashed, err := s.repo.TrashedTopup(ctx, int(topup.TopupID))
	s.NoError(err)
	s.NotNil(trashed)
}

func (s *TopupRepositoryTestSuite) TestRestoreTopup() {
	topup, err := s.createSeedTopup()
	s.Require().NoError(err)
	ctx := context.Background()

	_, err = s.repo.TrashedTopup(ctx, int(topup.TopupID))
	s.Require().NoError(err)

	restored, err := s.repo.RestoreTopup(ctx, int(topup.TopupID))
	s.NoError(err)
	s.NotNil(restored)
}

func (s *TopupRepositoryTestSuite) TestDeleteTopupPermanent() {
	topup, err := s.createSeedTopup()
	s.Require().NoError(err)
	ctx := context.Background()

	_, err = s.repo.TrashedTopup(ctx, int(topup.TopupID))
	s.Require().NoError(err)

	success, err := s.repo.DeleteTopupPermanent(ctx, int(topup.TopupID))
	s.NoError(err)
	s.True(success)
}

func (s *TopupRepositoryTestSuite) TestRestoreAllTopup() {
	topup, err := s.createSeedTopup()
	s.Require().NoError(err)
	ctx := context.Background()

	_, err = s.repo.TrashedTopup(ctx, int(topup.TopupID))
	s.Require().NoError(err)

	success, err := s.repo.RestoreAllTopup(ctx)
	s.NoError(err)
	s.True(success)
}

func (s *TopupRepositoryTestSuite) TestDeleteAllTopupPermanent() {
	_, err := s.createSeedTopup()
	s.Require().NoError(err)
	ctx := context.Background()

	success, err := s.repo.DeleteAllTopupPermanent(ctx)
	s.NoError(err)
	s.True(success)
}

func TestTopupRepositorySuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(TopupRepositoryTestSuite))
}
