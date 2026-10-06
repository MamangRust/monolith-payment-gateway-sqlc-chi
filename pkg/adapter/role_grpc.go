package adapter

import (
	"context"

	"github.com/MamangRust/monolith-payment-gateway-pb/role"
	"github.com/MamangRust/monolith-payment-gateway-pkg/database/models"
	"github.com/MamangRust/monolith-payment-gateway-pkg/resilience"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
)

// QueryRepository is the read path consumers use to resolve roles.
type QueryRepository interface {
	FindById(ctx context.Context, id int) (*models.Role, error)
	FindByName(ctx context.Context, name string) (*models.Role, error)
}

// RoleAdapter is the contract consumers depend on for role lookups and role
// assignment through the role service.
type RoleAdapter interface {
	QueryRepository
	AssignRoleToUser(ctx context.Context, request *requests.CreateUserRoleRequest) (*models.UserRole, error)
	RemoveRoleFromUser(ctx context.Context, request *requests.RemoveUserRoleRequest) error
}

type roleGRPCAdapter struct {
	queryClient   role.RoleServiceClient
	commandClient role.RoleCommandServiceClient
	guard         *resilience.DependencyGuard
}

func (a *roleGRPCAdapter) setGuard(g *resilience.DependencyGuard) {
	a.guard = g
}

// NewRoleAdapter wraps the generated role query/command clients. Passing nil
// for commandClient is valid for consumers that only read.
func NewRoleAdapter(queryClient role.RoleServiceClient, commandClient role.RoleCommandServiceClient, opts ...GuardOption) RoleAdapter {
	a := &roleGRPCAdapter{
		queryClient:   queryClient,
		commandClient: commandClient,
	}
	for _, opt := range opts {
		opt(a)
	}
	return a
}

func (a *roleGRPCAdapter) FindById(ctx context.Context, id int) (*models.Role, error) {
	var resp *role.ApiResponseRole
	err := a.guard.Call(ctx, func(callCtx context.Context) error {
		var callErr error
		resp, callErr = a.queryClient.FindByIdRole(callCtx, &role.FindByIdRoleRequest{RoleId: int32(id)})
		return callErr
	})
	if err != nil {
		return nil, err
	}
	if resp.Data == nil {
		return nil, nil
	}
	return &models.Role{RoleID: resp.Data.Id, RoleName: resp.Data.Name}, nil
}

func (a *roleGRPCAdapter) FindByName(ctx context.Context, name string) (*models.Role, error) {
	var resp *role.ApiResponseRole
	err := a.guard.Call(ctx, func(callCtx context.Context) error {
		var callErr error
		resp, callErr = a.queryClient.FindByNameRole(callCtx, &role.FindByNameRoleRequest{Name: name})
		return callErr
	})
	if err != nil {
		return nil, err
	}
	if resp.Data == nil {
		return nil, nil
	}
	return &models.Role{RoleID: resp.Data.Id, RoleName: resp.Data.Name}, nil
}

func (a *roleGRPCAdapter) AssignRoleToUser(ctx context.Context, request *requests.CreateUserRoleRequest) (*models.UserRole, error) {
	err := a.guard.Call(ctx, func(callCtx context.Context) error {
		_, callErr := a.commandClient.CreateUserRole(callCtx, &role.CreateUserRoleRequest{
			UserId: int32(request.UserId),
			RoleId: int32(request.RoleId),
		})
		return callErr
	})
	if err != nil {
		return nil, err
	}
	return &models.UserRole{UserID: int32(request.UserId), RoleID: int32(request.RoleId)}, nil
}

func (a *roleGRPCAdapter) RemoveRoleFromUser(ctx context.Context, request *requests.RemoveUserRoleRequest) error {
	return a.guard.Call(ctx, func(callCtx context.Context) error {
		_, callErr := a.commandClient.DeleteUserRole(callCtx, &role.DeleteUserRoleRequest{
			UserId: int32(request.UserId),
			RoleId: int32(request.RoleId),
		})
		return callErr
	})
}
