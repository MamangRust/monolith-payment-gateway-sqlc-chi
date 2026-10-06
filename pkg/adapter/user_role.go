package adapter

import (
	"context"

	pbrole "github.com/MamangRust/monolith-payment-gateway-pb/role"
	pbuserrole "github.com/MamangRust/monolith-payment-gateway-pb/user_role"
	"github.com/MamangRust/monolith-payment-gateway-pkg/database/models"
	"github.com/MamangRust/monolith-payment-gateway-pkg/resilience"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
	userrole_errors "github.com/MamangRust/monolith-payment-gateway-shared/errors/user_role_errors/repository"
)

// UserRoleAdapter is the surface consumers depend on to resolve a user's roles
// and to (un)assign roles through the dedicated user-role service.
type UserRoleAdapter interface {
	FindByUserId(ctx context.Context, userID int) ([]*models.Role, error)
	AssignRoleToUser(ctx context.Context, request *requests.CreateUserRoleRequest) (*models.UserRole, error)
	RemoveRoleFromUser(ctx context.Context, request *requests.RemoveUserRoleRequest) error
}

type userRoleGRPCAdapter struct {
	client pbuserrole.UserRoleServiceClient
	guard  *resilience.DependencyGuard
}

func (a *userRoleGRPCAdapter) setGuard(g *resilience.DependencyGuard) {
	a.guard = g
}

// NewUserRoleAdapter wraps the generated user-role service client.
func NewUserRoleAdapter(client pbuserrole.UserRoleServiceClient, opts ...GuardOption) UserRoleAdapter {
	a := &userRoleGRPCAdapter{client: client}
	for _, opt := range opts {
		opt(a)
	}
	return a
}

func (a *userRoleGRPCAdapter) FindByUserId(ctx context.Context, userID int) ([]*models.Role, error) {
	var resp *pbrole.ApiResponsesRole
	err := a.guard.Call(ctx, func(callCtx context.Context) error {
		var callErr error
		resp, callErr = a.client.FindByUserId(callCtx, &pbuserrole.FindByIdUserRoleRequest{UserId: int32(userID)})
		return callErr
	})
	if err != nil {
		return nil, err
	}
	if resp == nil {
		return nil, nil
	}

	roles := make([]*models.Role, 0, len(resp.Data))
	for _, item := range resp.Data {
		if item == nil {
			continue
		}
		roles = append(roles, &models.Role{
			RoleID:    item.Id,
			RoleName:  item.Name,
			CreatedAt: parseTime(item.CreatedAt),
			UpdatedAt: parseTime(item.UpdatedAt),
		})
	}
	return roles, nil
}

func (a *userRoleGRPCAdapter) AssignRoleToUser(ctx context.Context, request *requests.CreateUserRoleRequest) (*models.UserRole, error) {
	var resp *pbuserrole.ApiResponseUserRole
	err := a.guard.Call(ctx, func(callCtx context.Context) error {
		var callErr error
		resp, callErr = a.client.AssignRoleToUser(callCtx, &pbuserrole.AssignRoleToUserRequest{
			UserId: int32(request.UserId),
			RoleId: int32(request.RoleId),
		})
		return callErr
	})
	if err != nil {
		return nil, userrole_errors.ErrAssignRoleToUser.WithInternal(err)
	}
	if resp == nil || resp.Data == nil {
		return &models.UserRole{UserID: int32(request.UserId), RoleID: int32(request.RoleId)}, nil
	}
	return &models.UserRole{
		UserRoleID: resp.Data.UserRoleId,
		UserID:     resp.Data.UserId,
		RoleID:     resp.Data.RoleId,
	}, nil
}

func (a *userRoleGRPCAdapter) RemoveRoleFromUser(ctx context.Context, request *requests.RemoveUserRoleRequest) error {
	err := a.guard.Call(ctx, func(callCtx context.Context) error {
		_, callErr := a.client.RemoveRoleFromUser(callCtx, &pbuserrole.RemoveRoleFromUserRequest{
			UserId: int32(request.UserId),
			RoleId: int32(request.RoleId),
		})
		return callErr
	})
	if err != nil {
		return userrole_errors.ErrRemoveRole.WithInternal(err)
	}
	return nil
}
