package npwpverification

import (
	"front-office/internal/core/log/transaction"
	"front-office/internal/core/member"
	"front-office/internal/datahub/job"
	"front-office/pkg/apperror"
	"front-office/pkg/common/constant"
	"front-office/pkg/helper"
)

func NewService(
	repo Repository,
	memberRepo member.Repository,
	jobRepo job.Repository,
	transactionRepo transaction.Repository,
	jobService job.Service,
) Service {
	return &service{
		repo,
		memberRepo,
		jobRepo,
		transactionRepo,
		jobService,
	}
}

type service struct {
	repo            Repository
	memberRepo      member.Repository
	jobRepo         job.Repository
	transactionRepo transaction.Repository
	jobService      job.Service
}

type Service interface {
	NPWPVerification(apiKey, memberId, companyId string, payload *npwpVerificationRequest) error
}

func (svc *service) NPWPVerification(apiKey, memberId, companyId string, payload *npwpVerificationRequest) error {
	subscribedResp, err := svc.memberRepo.GetSubscribedProducts(companyId, constant.SlugNPWPVerification)
	if err != nil {
		return apperror.MapRepoError(err, constant.ErrFetchSubscribedProduct)
	}

	jobRes, err := svc.jobRepo.CreateJobAPI(&job.CreateJobRequest{
		ProductId: subscribedResp.Data.ProductId,
		MemberId:  memberId,
		CompanyId: companyId,
		Total:     1,
	})
	if err != nil {
		return apperror.MapRepoError(err, constant.FailedCreateJob)
	}

	jobIdStr := helper.ConvertUintToString(jobRes.JobId)
	_, err = svc.repo.NPWPVerificationAPI(apiKey, jobIdStr, payload)
	if err != nil {
		if err := svc.jobService.FinalizeFailedJob(jobIdStr); err != nil {
			return err
		}

		return apperror.Internal("failed to process npwp verification", err)
	}

	return svc.jobService.FinalizeJob(jobIdStr)
}
