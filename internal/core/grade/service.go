package grade

import (
	"fmt"
	"front-office/pkg/apperror"
)

func NewService(repo Repository) Service {
	return &service{Repo: repo}
}

type service struct {
	Repo Repository
}

type Service interface {
	SaveGrading(payload *createGradePayload) error
	GetGrades(productSlug, companyId string) (*gradesResponseData, error)
}

func (svc *service) SaveGrading(payload *createGradePayload) error {
	for i := 0; i < len(payload.Request.Grades); i++ {
		for j := i + 1; j < len(payload.Request.Grades); j++ {
			if payload.Request.Grades[i].Grade == payload.Request.Grades[j].Grade {
				return apperror.BadRequest(fmt.Sprintf("duplicate grade: %s", payload.Request.Grades[i].Grade))
			}
			if !(payload.Request.Grades[i].End <= payload.Request.Grades[j].Start || payload.Request.Grades[j].End <= payload.Request.Grades[i].Start) {
				return apperror.BadRequest(fmt.Sprintf("overlapping grade range between %s and %s", payload.Request.Grades[i].Grade, payload.Request.Grades[j].Grade))
			}
		}
	}

	if err := svc.Repo.SaveGradingAPI(payload); err != nil {
		return apperror.MapRepoError(err, "failed to save grading")
	}

	return nil
}

func (svc *service) GetGrades(productSlug, companyId string) (*gradesResponseData, error) {
	result, err := svc.Repo.GetGradesAPI(productSlug, companyId)
	if err != nil {
		return nil, apperror.MapRepoError(err, "failed to get grades")
	}

	return result, nil
}
