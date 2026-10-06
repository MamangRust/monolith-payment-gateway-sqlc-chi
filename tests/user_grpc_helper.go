package tests

import (
	"net"

	pbuser "github.com/MamangRust/monolith-payment-gateway-pb/user"
	"github.com/MamangRust/monolith-payment-gateway-pkg/hash"
	"github.com/MamangRust/monolith-payment-gateway-pkg/logger"
	"github.com/MamangRust/monolith-payment-gateway-shared/cache"
	userhandler "github.com/MamangRust/monolith-payment-gateway-user/handler"
	userrepository "github.com/MamangRust/monolith-payment-gateway-user/repository"
	userservice "github.com/MamangRust/monolith-payment-gateway-user/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"gorm.io/gorm"
)

// StartUserService boots an in-process gRPC server exposing the user query and
// command services over the shared user database. Role lookups use the
// GORM-backed shim so no separate role service is required.
//
// It returns a client connection plus a cleanup func the caller should defer.
func StartUserService(db *gorm.DB, log logger.LoggerInterface, cacheStore *cache.CacheStore) (*grpc.ClientConn, func(), error) {
	repos := userrepository.NewRepositories(&userrepository.Deps{Db: db, Role: NewRoleRepo(db), UserRole: NewUserRoleRepo(db)})
	svc := userservice.NewService(&userservice.Deps{
		Repositories: repos,
		Logger:       log,
		Cache:        cacheStore,
		Hash:         hash.NewHashingPassword(),
	})
	h := userhandler.NewHandler(svc)

	gs := grpc.NewServer()
	pbuser.RegisterUserQueryServiceServer(gs, h)
	pbuser.RegisterUserCommandServiceServer(gs, h)

	lis, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return nil, nil, err
	}
	go func() {
		_ = gs.Serve(lis)
	}()

	conn, err := grpc.NewClient(lis.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		gs.Stop()
		return nil, nil, err
	}

	cleanup := func() {
		_ = conn.Close()
		gs.Stop()
	}
	return conn, cleanup, nil
}
