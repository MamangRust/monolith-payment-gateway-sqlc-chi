package adapter

import (
	"context"

	"github.com/MamangRust/monolith-payment-gateway-pb/user"
	"github.com/MamangRust/monolith-payment-gateway-pkg/database/models"
	"github.com/MamangRust/monolith-payment-gateway-pkg/resilience"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
)

// UserAdapter is the query-only surface used by consumers that just look up
// users (card, merchant, topup, ...).
type UserAdapter interface {
	FindById(ctx context.Context, userID int) (*models.User, error)
}

// AuthUserAdapter extends UserAdapter with the lookup/creation/update surface
// the auth service needs.
type AuthUserAdapter interface {
	UserAdapter
	FindByEmail(ctx context.Context, email string) (*models.User, error)
	FindByEmailAndVerify(ctx context.Context, email string) (*models.User, error)
	FindByVerificationCode(ctx context.Context, code string) (*models.User, error)
	CreateUser(ctx context.Context, request *requests.RegisterRequest) (*models.User, error)
	UpdateUserIsVerified(ctx context.Context, userID int, isVerified bool) error
	UpdateUserPassword(ctx context.Context, userID int, password string) error
}

type userGRPCAdapter struct {
	queryClient user.UserQueryServiceClient
	guard       *resilience.DependencyGuard
}

func (a *userGRPCAdapter) setGuard(g *resilience.DependencyGuard) {
	a.guard = g
}

func NewUserAdapter(queryClient user.UserQueryServiceClient, opts ...GuardOption) UserAdapter {
	a := &userGRPCAdapter{
		queryClient: queryClient,
	}
	for _, opt := range opts {
		opt(a)
	}
	return a
}

func (a *userGRPCAdapter) FindById(ctx context.Context, userID int) (*models.User, error) {
	var resp *user.ApiResponseUser
	err := a.guard.Call(ctx, func(callCtx context.Context) error {
		var callErr error
		resp, callErr = a.queryClient.FindById(callCtx, &user.FindByIdUserRequest{
			Id: int32(userID),
		})
		return callErr
	})
	if err != nil {
		return nil, err
	}

	return &models.User{
		UserID:    resp.Data.Id,
		Email:     resp.Data.Email,
		Firstname: resp.Data.Firstname,
		Lastname:  resp.Data.Lastname,
	}, nil
}

// NewAuthUserAdapter builds the full user surface needed by the auth service
func NewAuthUserAdapter(queryClient user.UserQueryServiceClient, commandClient user.UserCommandServiceClient, opts ...GuardOption) AuthUserAdapter {
	a := &authUserGRPCAdapter{
		userGRPCAdapter: userGRPCAdapter{queryClient: queryClient},
		commandClient:   commandClient,
	}
	for _, opt := range opts {
		opt(a)
	}
	return a
}

type authUserGRPCAdapter struct {
	userGRPCAdapter
	commandClient user.UserCommandServiceClient
}

func (a *authUserGRPCAdapter) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	var resp *user.ApiResponseUserWithPassword
	err := a.guard.Call(ctx, func(callCtx context.Context) error {
		var callErr error
		resp, callErr = a.queryClient.FindByEmail(callCtx, &user.FindByEmailUserRequest{Email: email})
		return callErr
	})
	if err != nil {
		return nil, err
	}
	if resp == nil || resp.Data == nil {
		return nil, nil
	}
	return &models.User{
		UserID:    resp.Data.Id,
		Firstname: resp.Data.Firstname,
		Lastname:  resp.Data.Lastname,
		Email:     resp.Data.Email,
		Password:  resp.Data.Password,
		CreatedAt: parseTime(resp.Data.CreatedAt),
		UpdatedAt: parseTime(resp.Data.UpdatedAt),
	}, nil
}

func (a *authUserGRPCAdapter) FindByEmailAndVerify(ctx context.Context, email string) (*models.User, error) {
	return a.FindByEmail(ctx, email)
}

func (a *authUserGRPCAdapter) FindByVerificationCode(ctx context.Context, code string) (*models.User, error) {
	var resp *user.ApiResponseUser
	err := a.guard.Call(ctx, func(callCtx context.Context) error {
		var callErr error
		resp, callErr = a.queryClient.FindByVerificationCode(callCtx, &user.FindByVerificationCodeUserRequest{VerificationCode: code})
		return callErr
	})
	if err != nil {
		return nil, err
	}
	return mapUserResponse(resp.Data), nil
}

func (a *authUserGRPCAdapter) CreateUser(ctx context.Context, request *requests.RegisterRequest) (*models.User, error) {
	var resp *user.ApiResponseUser
	err := a.guard.Call(ctx, func(callCtx context.Context) error {
		var callErr error
		resp, callErr = a.commandClient.Create(callCtx, &user.CreateUserRequest{
			Firstname:        request.FirstName,
			Lastname:         request.LastName,
			Email:            request.Email,
			Password:         request.Password,
			ConfirmPassword:  request.Password,
			VerificationCode: request.VerifiedCode,
			IsVerified:       request.IsVerified,
		})
		return callErr
	})
	if err != nil {
		return nil, err
	}
	return mapUserResponse(resp.Data), nil
}

func (a *authUserGRPCAdapter) UpdateUserIsVerified(ctx context.Context, userID int, isVerified bool) error {
	return a.guard.Call(ctx, func(callCtx context.Context) error {
		_, callErr := a.commandClient.UpdateIsVerified(callCtx, &user.UpdateUserIsVerifiedRequest{
			UserId:     int32(userID),
			IsVerified: isVerified,
		})
		return callErr
	})
}

func (a *authUserGRPCAdapter) UpdateUserPassword(ctx context.Context, userID int, password string) error {
	return a.guard.Call(ctx, func(callCtx context.Context) error {
		_, callErr := a.commandClient.UpdatePassword(callCtx, &user.UpdateUserPasswordRequest{
			UserId:   int32(userID),
			Password: password,
		})
		return callErr
	})
}

func mapUserResponse(u *user.UserResponse) *models.User {
	if u == nil {
		return nil
	}
	return &models.User{
		UserID:    u.Id,
		Firstname: u.Firstname,
		Lastname:  u.Lastname,
		Email:     u.Email,
		CreatedAt: parseTime(u.CreatedAt),
		UpdatedAt: parseTime(u.UpdatedAt),
	}
}
