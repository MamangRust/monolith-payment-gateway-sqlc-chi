package repository

import (
	pbrole "github.com/MamangRust/monolith-payment-gateway-pb/role"
	pbuserrole "github.com/MamangRust/monolith-payment-gateway-pb/user_role"
	"github.com/MamangRust/monolith-payment-gateway-pkg/adapter"
	"gorm.io/gorm"
)

// GuardOptions carries the resilience guard options for each outbound gRPC
// dependency.
type GuardOptions struct {
	Role     []adapter.GuardOption
	UserRole []adapter.GuardOption
}

// Repositories exposes the user data-access repositories. Callers depend on the
// contracts they use rather than on the concrete wiring.
type Repositories interface {
	UserQuery() UserQueryRepository
	UserCommand() UserCommandRepository
	Role() RoleRepository
	UserRole() UserRoleRepository
}

type repositories struct {
	userQuery   UserQueryRepository
	userCommand UserCommandRepository
	role        RoleRepository
	userRole    UserRoleRepository
}

func (r *repositories) UserQuery() UserQueryRepository     { return r.userQuery }
func (r *repositories) UserCommand() UserCommandRepository { return r.userCommand }
func (r *repositories) Role() RoleRepository               { return r.role }
func (r *repositories) UserRole() UserRoleRepository       { return r.userRole }

// Deps holds the DB handle plus the raw gRPC clients for the role and user-role
// services. The adapters are built here with the supplied guard options so
// resilience (timeout/circuit-breaker/bulkhead) is applied uniformly.
type Deps struct {
	Db       *gorm.DB
	Role     pbrole.RoleServiceClient
	UserRole pbuserrole.UserRoleServiceClient
	Guards   GuardOptions
}

// NewRepositories builds the GORM-backed user repositories and wraps the remote
// role/user-role clients in the shared adapters.
func NewRepositories(deps *Deps) Repositories {
	return &repositories{
		userQuery:   NewUserQueryRepository(deps.Db),
		userCommand: NewUserCommandRepository(deps.Db),
		role:        adapter.NewRoleAdapter(deps.Role, nil, deps.Guards.Role...),
		userRole:    adapter.NewUserRoleAdapter(deps.UserRole, deps.Guards.UserRole...),
	}
}
