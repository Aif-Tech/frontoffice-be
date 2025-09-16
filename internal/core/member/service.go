package member

import (
	"front-office/internal/core/log/operation"
	"front-office/internal/core/role"
	"front-office/pkg/apperror"
	"front-office/pkg/common/constant"
	"front-office/pkg/common/model"
	"front-office/pkg/helper"
	"front-office/pkg/utility/mailjet"
	"time"

	"github.com/rs/zerolog/log"
)

func NewService(repo Repository, roleRepo role.Repository, operationRepo operation.Repository) Service {
	return &service{
		repo,
		roleRepo,
		operationRepo,
	}
}

type service struct {
	repo          Repository
	roleRepo      role.Repository
	operationRepo operation.Repository
}

type Service interface {
	GetMemberBy(query *FindUserQuery) (*MstMember, error)
	GetMemberList(filter *MemberFilter) ([]*MstMember, *model.Meta, error)
	UpdateProfile(userId string, currentUserRoleId uint, req *UpdateProfileRequest) (*userUpdateResponse, error)
	UploadProfileImage(id string, filename *string) (*userUpdateResponse, error)
	UpdateMemberById(currentUserId, currentUserRoleId uint, companyId, memberId string, req *UpdateUserRequest) error
	DeleteMemberById(memberId, companyId string) error
}

func (svc *service) GetMemberBy(query *FindUserQuery) (*MstMember, error) {
	member, err := svc.repo.GetMemberAPI(query)
	if err != nil {
		return nil, apperror.MapRepoError(err, "failed to get member")
	}
	if member.MemberId == 0 {
		return nil, apperror.NotFound(constant.UserNotFound)
	}

	return member, nil
}

func (svc *service) GetMemberList(filter *MemberFilter) ([]*MstMember, *model.Meta, error) {
	users, meta, err := svc.repo.GetMemberListAPI(filter)
	if err != nil {
		return nil, nil, err
	}

	return users, meta, nil
}

func (svc *service) UpdateProfile(userId string, currentUserRoleId uint, req *UpdateProfileRequest) (*userUpdateResponse, error) {
	user, err := svc.repo.GetMemberAPI(&FindUserQuery{Id: userId})
	if err != nil {
		return nil, apperror.MapRepoError(err, constant.FailedFetchMember)
	}
	if user.MemberId == 0 {
		return nil, apperror.NotFound(constant.UserNotFound)
	}

	updateFields := make(map[string]interface{})
	shouldSendEmailConfirmation := false
	var newEmail string

	if req.Name != nil {
		updateFields["name"] = *req.Name
		user.Name = *req.Name
	}

	if req.Email != nil {
		if currentUserRoleId == uint(memberRoleID) {
			return nil, apperror.Unauthorized("you are not allowed to update email")
		}

		existing, err := svc.repo.GetMemberAPI(&FindUserQuery{Email: *req.Email})
		if err != nil {
			return nil, apperror.MapRepoError(err, "failed to check existing email")
		}
		if existing.MemberId != 0 {
			return nil, apperror.Conflict(constant.EmailAlreadyExists)
		}

		updateFields["email"] = *req.Email
		shouldSendEmailConfirmation = true
		newEmail = *req.Email
	}

	updateFields["updated_at"] = time.Now()

	if err := svc.repo.UpdateMemberAPI(userId, updateFields); err != nil {
		return nil, apperror.MapRepoError(err, constant.FailedUpdateMember)
	}

	if shouldSendEmailConfirmation {
		if err := mailjet.SendConfirmationEmailUserEmailChangeSuccess(user.Name, user.Email, newEmail, helper.FormatWIB(time.Now())); err != nil {
			return nil, apperror.Internal("failed to send email confirmation", err)
		}
		user.Email = newEmail
	}

	if err := svc.operationRepo.AddLogOperation(&operation.AddLogRequest{
		MemberId:  user.MemberId,
		CompanyId: user.CompanyId,
		Action:    constant.EventUpdateProfile,
	}); err != nil {
		log.Warn().Err(err).Msg("failed to log profile update event")
	}

	return &userUpdateResponse{
		Id:        user.MemberId,
		Name:      user.Name,
		Email:     user.Email,
		Active:    user.Active,
		CompanyId: user.CompanyId,
		RoleId:    user.RoleId,
	}, nil
}

func (svc *service) UploadProfileImage(userId string, filename *string) (*userUpdateResponse, error) {
	user, err := svc.repo.GetMemberAPI(&FindUserQuery{Id: userId})
	if err != nil {
		return nil, apperror.MapRepoError(err, constant.FailedFetchMember)
	}
	if user.MemberId == 0 {
		return nil, apperror.NotFound(constant.UserNotFound)
	}

	updateFields := make(map[string]interface{})

	if filename != nil {
		updateFields["image"] = *filename
	}

	updateFields["updated_at"] = time.Now()

	if err := svc.repo.UpdateMemberAPI(userId, updateFields); err != nil {
		return nil, apperror.MapRepoError(err, constant.FailedUpdateMember)
	}

	if err := svc.operationRepo.AddLogOperation(&operation.AddLogRequest{
		MemberId:  user.MemberId,
		CompanyId: user.CompanyId,
		Action:    constant.EventUpdateProfile,
	}); err != nil {
		log.Warn().Err(err).Msg("failed to log upload profile photo event")
	}

	return &userUpdateResponse{
		Id:        user.MemberId,
		Name:      user.Name,
		Email:     user.Email,
		Active:    user.Active,
		CompanyId: user.CompanyId,
		RoleId:    user.RoleId,
	}, nil
}

func (svc *service) UpdateMemberById(currentUserId, currentUserRoleId uint, companyId, memberId string, req *UpdateUserRequest) error {
	member, err := svc.repo.GetMemberAPI(&FindUserQuery{Id: memberId, CompanyId: companyId})
	if err != nil {
		return apperror.MapRepoError(err, constant.FailedFetchMember)
	}
	if member.MemberId == 0 {
		return apperror.NotFound(constant.UserNotFound)
	}

	updateFields := make(map[string]interface{})
	var (
		sendEmailConfirmation bool
		newEmail              string
		logEvents             []string
	)

	if req.Name != nil {
		updateFields["name"] = *req.Name
		logEvents = append(logEvents, constant.EventUpdateUserData)
	}

	if req.Email != nil {
		if currentUserRoleId == uint(memberRoleID) {
			return apperror.Unauthorized("you are not allowed to update email")
		}

		existing, err := svc.repo.GetMemberAPI(&FindUserQuery{Email: *req.Email})
		if err != nil {
			return apperror.MapRepoError(err, "failed to check existing email")
		}
		if existing.MemberId != 0 {
			return apperror.Conflict(constant.EmailAlreadyExists)
		}

		updateFields["email"] = *req.Email
		newEmail = *req.Email
		sendEmailConfirmation = true
		logEvents = append(logEvents, constant.EventUpdateUserData)
	}

	if req.RoleId != nil {
		role, err := svc.roleRepo.GetRoleByIdAPI(*req.RoleId)
		if err != nil {
			return apperror.MapRepoError(err, "failed to fetch role")
		}
		if role.RoleId == 0 {
			return apperror.NotFound("role not found")
		}

		updateFields["role_id"] = *req.RoleId
		logEvents = append(logEvents, constant.EventUpdateUserData)
	}

	if req.Active != nil {
		updateFields["active"] = *req.Active

		if *req.Active {
			logEvents = append(logEvents, constant.EventActivateUser)
		} else {
			logEvents = append(logEvents, constant.EventInactivateUser)
		}
	}

	updateFields["updated_at"] = time.Now()

	if err := svc.repo.UpdateMemberAPI(memberId, updateFields); err != nil {
		return apperror.MapRepoError(err, constant.FailedUpdateMember)
	}

	if sendEmailConfirmation {
		if err := mailjet.SendConfirmationEmailUserEmailChangeSuccess(member.Name, member.Email, newEmail, helper.FormatWIB(time.Now())); err != nil {
			return apperror.Internal("failed to send email confirmation", err)
		}
	}

	for _, event := range logEvents {
		if err := svc.operationRepo.AddLogOperation(&operation.AddLogRequest{
			MemberId:  currentUserId,
			CompanyId: member.CompanyId,
			Action:    event,
		}); err != nil {
			log.Warn().Err(err).Msg("failed to log update member event")
		}
	}

	return nil
}

func (svc *service) DeleteMemberById(memberId, companyId string) error {
	member, err := svc.repo.GetMemberAPI(&FindUserQuery{Id: memberId, CompanyId: companyId})
	if err != nil {
		return apperror.MapRepoError(err, constant.FailedFetchMember)
	}
	if member.MemberId == 0 {
		return apperror.NotFound(constant.UserNotFound)
	}

	if err := svc.repo.DeleteMemberAPI(memberId); err != nil {
		return apperror.MapRepoError(err, "failed to delete member")
	}

	return nil
}
