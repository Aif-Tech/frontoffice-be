package passwordreset

import (
	"front-office/configs/application"
	"front-office/pkg/apperror"
	"front-office/pkg/helper"
	"strconv"
)

func NewService(repo Repository, cfg *application.Config) Service {
	return &service{Repo: repo, Cfg: cfg}
}

type service struct {
	Repo Repository
	Cfg  *application.Config
}

type Service interface {
	GetPasswordResetToken(token string) (*MstPasswordResetToken, error)
	CreatePasswordResetToken(userId, companyId, roleId uint) (string, error)
	DeletePasswordResetToken(id uint) error
}

func (svc *service) GetPasswordResetToken(token string) (*MstPasswordResetToken, error) {
	data, err := svc.Repo.GetPasswordResetTokenAPI(token)
	if err != nil {
		return nil, apperror.MapRepoError(err, "failed to get password reset token")
	}

	return data, nil
}

func (svc *service) CreatePasswordResetToken(userId, companyId, roleId uint) (string, error) {
	secret := svc.Cfg.Env.JwtSecretKey
	minutesToExpired, err := strconv.Atoi(svc.Cfg.Env.JwtActivationExpiresMinutes)
	if err != nil {
		return "", apperror.Internal("invalid password reset expiry application", err)
	}

	token, err := helper.GenerateToken(secret, minutesToExpired, userId, companyId, roleId, "")
	if err != nil {
		return "", apperror.Internal("generate password reset token failed", err)
	}

	req := &CreatePasswordResetTokenRequest{
		Token: token,
	}

	userIdStr := helper.ConvertUintToString(userId)
	err = svc.Repo.CreatePasswordResetTokenAPI(userIdStr, req)
	if err != nil {
		return "", apperror.MapRepoError(err, "failed to create password reset token")
	}

	return token, nil
}

func (svc *service) DeletePasswordResetToken(id uint) error {
	idStr := strconv.Itoa(int(id))
	err := svc.Repo.DeletePasswordResetTokenAPI(idStr)
	if err != nil {
		return apperror.MapRepoError(err, "failed to delete password reset token")
	}

	return nil
}
