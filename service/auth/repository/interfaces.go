package repository

import (
	"context"

	"github.com/MamangRust/monolith-payment-gateway-pkg/adapter"
	"github.com/MamangRust/monolith-payment-gateway-pkg/database/models"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
)

//go:generate mockgen -source=interfaces.go -destination=mocks/mock.go

// UserRepository is the user data access contract, backed by the shared user
// adapter. The user service owns the users table (including the password hash
// and verification code), so auth reaches it over gRPC.
type UserRepository = adapter.AuthUserAdapter

type ResetTokenRepository interface {
	FindByToken(ctx context.Context, code string) (*models.ResetTokenRow, error)
	CreateResetToken(ctx context.Context, req *requests.CreateResetTokenRequest) (*models.ResetTokenRow, error)
	DeleteResetToken(ctx context.Context, user_id int) error
}

type RefreshTokenRepository interface {
	FindByToken(ctx context.Context, token string) (*models.RefreshToken, error)
	FindByUserId(ctx context.Context, user_id int) (*models.RefreshToken, error)
	CreateRefreshToken(ctx context.Context, req *requests.CreateRefreshToken) (*models.RefreshToken, error)
	UpdateRefreshToken(ctx context.Context, req *requests.UpdateRefreshToken) (*models.RefreshToken, error)
	DeleteRefreshToken(ctx context.Context, token string) error
	DeleteRefreshTokenByUserId(ctx context.Context, user_id int) error
}

// UserRoleRepository is satisfied by the shared user-role gRPC adapter: the
// user_roles writes live behind the UserRole service.
type UserRoleRepository interface {
	AssignRoleToUser(ctx context.Context, req *requests.CreateUserRoleRequest) (*models.UserRole, error)
	RemoveRoleFromUser(ctx context.Context, req *requests.RemoveUserRoleRequest) error
}

// RoleRepository is satisfied by the shared role gRPC adapter.
type RoleRepository interface {
	FindById(ctx context.Context, id int) (*models.Role, error)
	FindByName(ctx context.Context, name string) (*models.Role, error)
}

// Compile-time assertions: the shared adapters satisfy auth's repository
// contracts directly.
var (
	_ UserRepository     = adapter.AuthUserAdapter(nil)
	_ RoleRepository     = adapter.RoleAdapter(nil)
	_ UserRoleRepository = adapter.UserRoleAdapter(nil)
)
