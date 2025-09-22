package passwordreset

import (
	"front-office/configs/application"
	"front-office/pkg/apperror"
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
	DeletePasswordResetToken(id uint) error
}

func (svc *service) GetPasswordResetToken(token string) (*MstPasswordResetToken, error) {
	data, err := svc.Repo.GetPasswordResetTokenAPI(token)
	if err != nil {
		return nil, apperror.MapRepoError(err, "failed to get password reset token")
	}

	return data, nil
}

func (svc *service) DeletePasswordResetToken(id uint) error {
	idStr := strconv.Itoa(int(id))
	err := svc.Repo.DeletePasswordResetTokenAPI(idStr)
	if err != nil {
		return apperror.MapRepoError(err, "failed to delete password reset token")
	}

	return nil
}
