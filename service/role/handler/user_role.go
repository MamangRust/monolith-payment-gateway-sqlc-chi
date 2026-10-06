package handler

import (
	"context"

	pb "github.com/MamangRust/monolith-payment-gateway-pb/role"
	pbuserrole "github.com/MamangRust/monolith-payment-gateway-pb/user_role"
	"github.com/MamangRust/monolith-payment-gateway-role/service"
	"github.com/MamangRust/monolith-payment-gateway-shared/domain/requests"
	"github.com/MamangRust/monolith-payment-gateway-shared/errors"
	role_errors "github.com/MamangRust/monolith-payment-gateway-shared/errors/role_errors/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

// userRoleHandleGrpc serves the UserRoleService, the split-out surface that owns
// user-role assignment and per-user role resolution.
type userRoleHandleGrpc struct {
	pbuserrole.UnimplementedUserRoleServiceServer
	roleQuery   service.RoleQueryService
	roleCommand service.RoleCommandService
}

func NewUserRoleHandleGrpc(service *service.Service) UserRoleHandlerGrpc {
	return &userRoleHandleGrpc{
		roleQuery:   service.RoleQuery,
		roleCommand: service.RoleCommand,
	}
}

func (s *userRoleHandleGrpc) FindByUserId(ctx context.Context, req *pbuserrole.FindByIdUserRoleRequest) (*pb.ApiResponsesRole, error) {
	userID := int(req.GetUserId())
	if userID <= 0 {
		return nil, role_errors.ErrGrpcRoleInvalidId
	}

	roles, err := s.roleQuery.FindByUserId(ctx, userID)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	protoRoles := make([]*pb.RoleResponse, len(roles))
	for i, role := range roles {
		protoRoles[i] = &pb.RoleResponse{
			Id:        int32(role.RoleID),
			Name:      role.RoleName,
			CreatedAt: role.CreatedAt.Format("2006-01-02"),
			UpdatedAt: role.UpdatedAt.Format("2006-01-02"),
		}
	}

	return &pb.ApiResponsesRole{
		Status:  "success",
		Message: "Successfully fetched role by user id",
		Data:    protoRoles,
	}, nil
}

func (s *userRoleHandleGrpc) AssignRoleToUser(ctx context.Context, reqPb *pbuserrole.AssignRoleToUserRequest) (*pbuserrole.ApiResponseUserRole, error) {
	if reqPb.GetUserId() <= 0 || reqPb.GetRoleId() <= 0 {
		return nil, role_errors.ErrGrpcRoleInvalidId
	}

	userRole, err := s.roleCommand.AssignRoleToUser(ctx, &requests.CreateUserRoleRequest{
		UserId: int(reqPb.GetUserId()),
		RoleId: int(reqPb.GetRoleId()),
	})
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	return &pbuserrole.ApiResponseUserRole{
		Status:  "success",
		Message: "Successfully assigned role to user",
		Data: &pbuserrole.UserRoleResponse{
			UserRoleId: userRole.UserRoleID,
			UserId:     userRole.UserID,
			RoleId:     userRole.RoleID,
		},
	}, nil
}

func (s *userRoleHandleGrpc) RemoveRoleFromUser(ctx context.Context, reqPb *pbuserrole.RemoveRoleFromUserRequest) (*emptypb.Empty, error) {
	if reqPb.GetUserId() <= 0 || reqPb.GetRoleId() <= 0 {
		return nil, role_errors.ErrGrpcRoleInvalidId
	}

	if err := s.roleCommand.RemoveRoleFromUser(ctx, &requests.RemoveUserRoleRequest{
		UserId: int(reqPb.GetUserId()),
		RoleId: int(reqPb.GetRoleId()),
	}); err != nil {
		return nil, errors.ToGrpcError(err)
	}

	return &emptypb.Empty{}, nil
}
