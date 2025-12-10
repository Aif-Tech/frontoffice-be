package recyclenumber

import (
	"errors"
	"front-office/internal/core/log/operation"
	"front-office/internal/core/log/transaction"
	"front-office/internal/core/member"
	"front-office/internal/datahub/job"
	"front-office/pkg/apperror"
	"front-office/pkg/common/constant"
	"front-office/pkg/common/model"
)

func NewService(
	repo Repository,
	memberRepo member.Repository,
	jobRepo job.Repository,
	operationRepo operation.Repository,
	transactionRepo transaction.Repository,
	jobService job.Service) Service {
	return &service{
		repo,
		memberRepo,
		jobRepo,
		operationRepo,
		transactionRepo,
		jobService,
	}
}

type service struct {
	repo            Repository
	memberRepo      member.Repository
	jobRepo         job.Repository
	operationRepo   operation.Repository
	transactionRepo transaction.Repository
	jobService      job.Service
}

type Service interface {
	RecycleNumber(authCtx *model.AuthContext, reqBody *recycleNumberRequest) (*model.ProCatAPIResponse[dataRecycleNumberAPI], error)
}

func (svc *service) RecycleNumber(authCtx *model.AuthContext, reqBody *recycleNumberRequest) (*model.ProCatAPIResponse[dataRecycleNumberAPI], error) {
	_, err := svc.memberRepo.GetSubscribedProducts(authCtx.CompanyIdStr(), constant.SlugRecycleNumber)
	if err != nil {
		return nil, apperror.MapRepoError(err, constant.ErrFetchSubscribedProduct)
	}

	// jobRes, err := svc.jobRepo.CreateJobAPI(&job.CreateJobRequest{
	// 	ProductId: subscribedResp.Data.ProductId,
	// 	MemberId:  authCtx.UserIdStr(),
	// 	CompanyId: authCtx.CompanyIdStr(),
	// 	Total:     1,
	// })
	// if err != nil {
	// 	return nil, apperror.MapRepoError(err, constant.FailedCreateJob)
	// }
	// jobIdStr := helper.ConvertUintToString(jobRes.JobId)

	jobIdStr := "dummy"

	result, err := svc.repo.RecycleNumberAPI(authCtx.APIKey, jobIdStr, authCtx.UserIdStr(), authCtx.CompanyIdStr(), reqBody)
	if err != nil {
		if err := svc.jobService.FinalizeFailedJob(jobIdStr); err != nil {
			return nil, err
		}

		var apiErr *apperror.ExternalAPIError
		if errors.As(err, &apiErr) {
			return nil, apperror.MapLoanError(apiErr)
		}

		return nil, apperror.Internal("failed to process loan record checker", err)
	}

	// if err := svc.jobService.FinalizeJob(jobIdStr); err != nil {
	// 	return nil, err
	// }

	// if err := svc.operationRepo.AddLogOperation(&operation.AddLogRequest{
	// 	MemberId:  authCtx.UserId,
	// 	CompanyId: authCtx.CompanyId,
	// 	Action:    constant.EventRecycleNumberSingleReq,
	// }); err != nil {
	// 	log.Warn().
	// 		Err(err).
	// 		Str("action", constant.EventRecycleNumberSingleReq).
	// 		Msg("failed to add operation log")
	// }

	return result, nil
}
