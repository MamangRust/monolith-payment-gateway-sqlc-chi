package tests

import (
	"net"

	"github.com/MamangRust/monolith-payment-gateway-card/handler"
	cardrepository "github.com/MamangRust/monolith-payment-gateway-card/repository"
	cardservice "github.com/MamangRust/monolith-payment-gateway-card/service"
	"github.com/MamangRust/monolith-payment-gateway-pkg/logger"
	"github.com/MamangRust/monolith-payment-gateway-saldo/repository"
	"github.com/MamangRust/monolith-payment-gateway-shared/cache"
	pbcard "github.com/MamangRust/monolith-payment-gateway-pb/card"
	pbuser "github.com/MamangRust/monolith-payment-gateway-pb/user"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"gorm.io/gorm"
)

// StartCardService boots an in-process gRPC server exposing the card query and
// command services over the shared database. The card service reaches the user
// service through its own guarded gRPC adapter, which is wired here to an
// in-process user service (see StartUserService).
//
// It returns the card client connection, the user client connection (so callers
// that also need a card repository for seeding can build one without spinning up
// a second user service), a cleanup func that the caller should defer, and any
// error encountered while booting.
func StartCardService(db *gorm.DB, log logger.LoggerInterface, cacheStore *cache.CacheStore) (*grpc.ClientConn, *grpc.ClientConn, func(), error) {
	userConn, userCleanup, err := StartUserService(db, log, cacheStore)
	if err != nil {
		return nil, nil, nil, err
	}

	cardRepos := cardrepository.NewRepositories(db, pbuser.NewUserQueryServiceClient(userConn))
	cardSvc := cardservice.NewService(&cardservice.Deps{
		Cache:        cacheStore,
		Repositories: cardRepos,
		Logger:       log,
		Kafka:        nil,
	})
	cardH := handler.NewHandler(cardSvc)

	gs := grpc.NewServer()
	pbcard.RegisterCardQueryServiceServer(gs, cardH)
	pbcard.RegisterCardCommandServiceServer(gs, cardH)

	lis, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		userCleanup()
		return nil, nil, nil, err
	}
	go func() {
		_ = gs.Serve(lis)
	}()

	conn, err := grpc.NewClient(lis.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		gs.Stop()
		userCleanup()
		return nil, nil, nil, err
	}

	cleanup := func() {
		_ = conn.Close()
		gs.Stop()
		userCleanup()
	}
	return conn, userConn, cleanup, nil
}

// StartCardInfra boots the in-process user + card gRPC services and builds the
// repositories that integration suites in other services need:
//
//   - saldoRepos: the saldo repositories wired to the card adapter (so saldo
//     operations resolve cards through the card gRPC service).
//   - cardRepos:  the card repositories backed by the same database, for seeding
//     cards through the card service's data layer.
//
// Both share the single in-process user service started by StartCardService, so
// callers should defer the returned cleanup func exactly once.
func StartCardInfra(
	db *gorm.DB,
	log logger.LoggerInterface,
	cacheStore *cache.CacheStore,
) (repository.Repositories, *cardrepository.Repositories, func(), error) {
	cardConn, userConn, cleanup, err := StartCardService(db, log, cacheStore)
	if err != nil {
		return nil, nil, nil, err
	}

	saldoRepos := repository.NewRepositories(
		db,
		pbcard.NewCardQueryServiceClient(cardConn),
		pbcard.NewCardCommandServiceClient(cardConn),
	)
	cardRepos := cardrepository.NewRepositories(db, pbuser.NewUserQueryServiceClient(userConn))

	return saldoRepos, cardRepos, cleanup, nil
}
