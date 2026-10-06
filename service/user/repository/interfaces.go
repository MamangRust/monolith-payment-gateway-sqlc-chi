package repository

import (
	"context"

	"github.com/MamangRust/monolith-payment-gateway-pkg/adapter"
	"github.com/MamangRust/monolith-payment-gateway-pkg/database/models"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
)

type UserQueryRepository interface {
	FindAllUsers(ctx context.Context, req *requests.FindAllUsers) ([]*models.UserRow, error)
	FindByActive(ctx context.Context, req *requests.FindAllUsers) ([]*models.UserActiveRow, error)
	FindByTrashed(ctx context.Context, req *requests.FindAllUsers) ([]*models.UserTrashedRow, error)
	FindById(ctx context.Context, user_id int) (*models.UserByIDRow, error)
	FindByEmail(ctx context.Context, email string) (*models.UserByEmailWithPasswordRow, error)
	FindByVerificationCode(ctx context.Context, code string) (*models.UserByVerificationCodeRow, error)
}

type UserCommandRepository interface {
	CreateUser(ctx context.Context, request *requests.CreateUserRequest) (*models.CreateUserRow, error)
	UpdateUser(ctx context.Context, request *requests.UpdateUserRequest) (*models.UpdateUserRow, error)
	UpdateUserIsVerified(ctx context.Context, user_id int, is_verified bool) error
	UpdateUserPassword(ctx context.Context, user_id int, password string) error
	TrashedUser(ctx context.Context, user_id int) (*models.TrashUserRow, error)
	RestoreUser(ctx context.Context, user_id int) (*models.RestoreUserRow, error)
	DeleteUserPermanent(ctx context.Context, user_id int) (bool, error)
	RestoreAllUser(ctx context.Context) (bool, error)
	DeleteAllUserPermanent(ctx context.Context) (bool, error)
}

// RoleRepository is the role read contract, backed by the shared role adapter.
type RoleRepository = adapter.QueryRepository

// UserRoleRepository is the role-assignment contract, backed by the shared
// user-role adapter. The default role is assigned through the role service.
type UserRoleRepository interface {
	AssignRoleToUser(ctx context.Context, request *requests.CreateUserRoleRequest) (*models.UserRole, error)
}

// Compile-time assertions: the shared adapters satisfy the contracts above.
var (
	_ RoleRepository     = adapter.RoleAdapter(nil)
	_ UserRoleRepository = adapter.UserRoleAdapter(nil)
)
