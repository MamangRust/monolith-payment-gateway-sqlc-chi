package repository

import (
	"context"

	"github.com/MamangRust/monolith-payment-gateway-pkg/database/models"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
	userrole_errors "github.com/MamangRust/monolith-payment-gateway-shared/errors/user_role_errors/repository"
	"gorm.io/gorm"
)

type userRoleCommandRepository struct {
	db *gorm.DB
}

func NewUserRoleCommandRepository(db *gorm.DB) UserRoleCommandRepository {
	return &userRoleCommandRepository{db: db}
}

func (r *userRoleCommandRepository) AssignRoleToUser(ctx context.Context, request *requests.CreateUserRoleRequest) (*models.UserRole, error) {
	userRole := &models.UserRole{
		UserID: int32(request.UserId),
		RoleID: int32(request.RoleId),
	}

	// Idempotent: a user may already hold the role, so reuse the existing
	// assignment instead of inserting a duplicate row.
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND role_id = ?", request.UserId, request.RoleId).
		FirstOrCreate(userRole).Error; err != nil {
		return nil, userrole_errors.ErrAssignRoleToUser.WithInternal(err)
	}

	return userRole, nil
}

func (r *userRoleCommandRepository) RemoveRoleFromUser(ctx context.Context, request *requests.RemoveUserRoleRequest) error {
	result := r.db.WithContext(ctx).
		Where("user_id = ? AND role_id = ?", request.UserId, request.RoleId).
		Delete(&models.UserRole{})

	if result.Error != nil {
		return userrole_errors.ErrRemoveRole.WithInternal(result.Error)
	}

	return nil
}
