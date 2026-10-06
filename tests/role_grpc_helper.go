package tests

import (
	"net"

	pbrole "github.com/MamangRust/monolith-payment-gateway-pb/role"
	pbuserrole "github.com/MamangRust/monolith-payment-gateway-pb/user_role"
	"github.com/MamangRust/monolith-payment-gateway-pkg/logger"
	rolehandler "github.com/MamangRust/monolith-payment-gateway-role/handler"
	rolerepository "github.com/MamangRust/monolith-payment-gateway-role/repository"
	roleservice "github.com/MamangRust/monolith-payment-gateway-role/service"
	"github.com/MamangRust/monolith-payment-gateway-shared/cache"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"gorm.io/gorm"
)

// StartRoleService boots an in-process gRPC server that exposes the role query,
// role command and user-role services over the shared role database. Consumers
// that talk to role/user-role through the shared adapters (e.g. auth) use this
// to stand the dependency up without a real network hop.
//
// It returns a client connection plus a cleanup func the caller should defer.
func StartRoleService(db *gorm.DB, log logger.LoggerInterface, cacheStore *cache.CacheStore) (*grpc.ClientConn, func(), error) {
	repos := rolerepository.NewRepositories(db)
	svc := roleservice.NewService(&roleservice.Deps{
		Repositories: repos,
		Logger:       log,
		Cache:        cacheStore,
	})
	h := rolehandler.NewHandler(svc)

	gs := grpc.NewServer()
	pbrole.RegisterRoleServiceServer(gs, h.RoleQuery)
	pbrole.RegisterRoleCommandServiceServer(gs, h.RoleCommand)
	pbuserrole.RegisterUserRoleServiceServer(gs, h.UserRole)

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
