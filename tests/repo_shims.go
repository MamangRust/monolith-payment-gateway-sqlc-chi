package tests

import (
	"context"
	"errors"

	pbrole "github.com/MamangRust/monolith-payment-gateway-pb/role"
	pbuserrole "github.com/MamangRust/monolith-payment-gateway-pb/user_role"
	"github.com/MamangRust/monolith-payment-gateway-pkg/database/models"
	saldorepository "github.com/MamangRust/monolith-payment-gateway-saldo/repository"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
	"gorm.io/gorm"
)

// SeedRoleAdmin ensures the default ROLE_ADMIN exists. The user service assigns
// this role when creating a user, so any suite that creates users through the
// user service must seed it first (migrations do not).
func SeedRoleAdmin(db *gorm.DB) error {
	return db.WithContext(context.Background()).
		Exec("INSERT INTO roles (role_name) VALUES ('ROLE_ADMIN') ON CONFLICT (role_name) DO NOTHING").
		Error
}

// The production user service reaches role/user-role through the shared gRPC
// adapters, which are built from the generated role/user-role clients.
// Integration suites that only need the data (and not the network hop) can use
// these GORM-backed implementations of the client interfaces instead.

var errShimNotImplemented = errors.New("gorm shim: method not implemented")

// --- Card ---

type gormCardRepo struct {
	db *gorm.DB
}

// NewCardRepo returns a GORM-backed saldo CardRepository.
func NewCardRepo(db *gorm.DB) saldorepository.CardRepository {
	return &gormCardRepo{db: db}
}

func (r *gormCardRepo) FindCardByCardNumber(ctx context.Context, cardNumber string) (*models.CardAllFieldsRow, error) {
	var card models.Card
	if err := r.db.WithContext(ctx).
		Where("card_number = ?", cardNumber).
		First(&card).Error; err != nil {
		return nil, err
	}
	return &models.CardAllFieldsRow{
		CardID:             card.CardID,
		UserID:             card.UserID,
		CardNumber:         card.CardNumber,
		CardType:           card.CardType,
		ExpireDate:         card.ExpireDate,
		Cvv:                card.Cvv,
		CardProvider:       card.CardProvider,
		Status:             card.Status,
		CreditLimit:        card.CreditLimit,
		OutstandingBalance: card.OutstandingBalance,
		RewardPoints:       card.RewardPoints,
		CreatedAt:          card.CreatedAt,
		UpdatedAt:          card.UpdatedAt,
	}, nil
}

// --- Role (gRPC client shim) ---

type gormRoleClient struct {
	db *gorm.DB
}

// NewRoleRepo returns a GORM-backed role.RoleServiceClient so integration suites
// can build the shared role adapter without a network hop.
func NewRoleRepo(db *gorm.DB) pbrole.RoleServiceClient {
	return &gormRoleClient{db: db}
}

func (c *gormRoleClient) FindByIdRole(ctx context.Context, in *pbrole.FindByIdRoleRequest, _ ...grpc.CallOption) (*pbrole.ApiResponseRole, error) {
	var role models.Role
	if err := c.db.WithContext(ctx).Where("role_id = ?", in.RoleId).First(&role).Error; err != nil {
		return nil, err
	}
	return &pbrole.ApiResponseRole{Data: &pbrole.RoleResponse{Id: role.RoleID, Name: role.RoleName}}, nil
}

func (c *gormRoleClient) FindByNameRole(ctx context.Context, in *pbrole.FindByNameRoleRequest, _ ...grpc.CallOption) (*pbrole.ApiResponseRole, error) {
	var role models.Role
	if err := c.db.WithContext(ctx).Where("role_name = ?", in.Name).First(&role).Error; err != nil {
		return nil, err
	}
	return &pbrole.ApiResponseRole{Data: &pbrole.RoleResponse{Id: role.RoleID, Name: role.RoleName}}, nil
}

// The remaining RPCs are unused by the user repository; they exist only to
// satisfy the generated client interface.

func (c *gormRoleClient) FindAllRole(context.Context, *pbrole.FindAllRoleRequest, ...grpc.CallOption) (*pbrole.ApiResponsePaginationRole, error) {
	return nil, errShimNotImplemented
}

func (c *gormRoleClient) FindByActive(context.Context, *pbrole.FindAllRoleRequest, ...grpc.CallOption) (*pbrole.ApiResponsePaginationRoleDeleteAt, error) {
	return nil, errShimNotImplemented
}

func (c *gormRoleClient) FindByTrashed(context.Context, *pbrole.FindAllRoleRequest, ...grpc.CallOption) (*pbrole.ApiResponsePaginationRoleDeleteAt, error) {
	return nil, errShimNotImplemented
}

func (c *gormRoleClient) FindByUserId(context.Context, *pbrole.FindByIdUserRoleRequest, ...grpc.CallOption) (*pbrole.ApiResponsesRole, error) {
	return nil, errShimNotImplemented
}

// --- User role (gRPC client shim) ---

type gormUserRoleClient struct {
	db *gorm.DB
}

// NewUserRoleRepo returns a GORM-backed user_role.UserRoleServiceClient so
// integration suites can build the shared user-role adapter without a network
// hop.
func NewUserRoleRepo(db *gorm.DB) pbuserrole.UserRoleServiceClient {
	return &gormUserRoleClient{db: db}
}

func (c *gormUserRoleClient) AssignRoleToUser(ctx context.Context, in *pbuserrole.AssignRoleToUserRequest, _ ...grpc.CallOption) (*pbuserrole.ApiResponseUserRole, error) {
	userRole := &models.UserRole{
		UserID: in.UserId,
		RoleID: in.RoleId,
	}
	if err := c.db.WithContext(ctx).
		Where("user_id = ? AND role_id = ?", in.UserId, in.RoleId).
		FirstOrCreate(userRole).Error; err != nil {
		return nil, err
	}
	return &pbuserrole.ApiResponseUserRole{Data: &pbuserrole.UserRoleResponse{
		UserRoleId: userRole.UserRoleID,
		UserId:     userRole.UserID,
		RoleId:     userRole.RoleID,
	}}, nil
}

func (c *gormUserRoleClient) FindByUserId(context.Context, *pbuserrole.FindByIdUserRoleRequest, ...grpc.CallOption) (*pbrole.ApiResponsesRole, error) {
	return nil, errShimNotImplemented
}

func (c *gormUserRoleClient) RemoveRoleFromUser(context.Context, *pbuserrole.RemoveRoleFromUserRequest, ...grpc.CallOption) (*emptypb.Empty, error) {
	return nil, errShimNotImplemented
}
