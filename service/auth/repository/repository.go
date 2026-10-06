package repository

import (
	pbrole "github.com/MamangRust/monolith-payment-gateway-pb/role"
	pbuser "github.com/MamangRust/monolith-payment-gateway-pb/user"
	pbuserrole "github.com/MamangRust/monolith-payment-gateway-pb/user_role"
	"github.com/MamangRust/monolith-payment-gateway-pkg/adapter"
	"gorm.io/gorm"
)

type Repositories struct {
	User         UserRepository
	RefreshToken RefreshTokenRepository
	UserRole     UserRoleRepository
	Role         RoleRepository
	ResetToken   ResetTokenRepository
}

type GuardOptions struct {
	User     []adapter.GuardOption
	UserRole []adapter.GuardOption
	Role     []adapter.GuardOption
}

func NewRepositories(
	db *gorm.DB,
	userQueryClient pbuser.UserQueryServiceClient,
	userCommandClient pbuser.UserCommandServiceClient,
	roleClient pbrole.RoleServiceClient,
	userRoleClient pbuserrole.UserRoleServiceClient,
	guards ...GuardOptions,
) *Repositories {
	var g GuardOptions
	if len(guards) > 0 {
		g = guards[0]
	}

	return &Repositories{
		User:         adapter.NewAuthUserAdapter(userQueryClient, userCommandClient, g.User...),
		UserRole:     adapter.NewUserRoleAdapter(userRoleClient, g.Role...),
		RefreshToken: NewRefreshTokenRepository(db),
		Role:         adapter.NewRoleAdapter(roleClient, nil, g.Role...),
		ResetToken:   NewResetTokenRepository(db),
	}
}
