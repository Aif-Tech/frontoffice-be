package phonenik

import (
	"errors"
	"front-office/internal/core/log/operation"
	"front-office/internal/core/log/transaction"
	"front-office/internal/core/member"
	"front-office/internal/datahub/job"
	"front-office/pkg/apperror"
	"front-office/pkg/common/constant"
	"front-office/pkg/common/model"
	"front-office/pkg/helper"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
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
	PhoneToNIK(authCtx *model.AuthContext, reqBody *phoneNIKRequest) (*model.ProCatAPIResponse[dataPhoneNIKAPI], error)
}

func (svc *service) PhoneToNIK(authCtx *model.AuthContext, reqBody *phoneNIKRequest) (*model.ProCatAPIResponse[dataPhoneNIKAPI], error) {
	subscribedResp, err := svc.memberRepo.GetSubscribedProducts(authCtx.CompanyIdStr(), constant.SlugPhoneNIKMatching)
	if err != nil {
		return nil, apperror.MapRepoError(err, constant.ErrFetchSubscribedProduct)
	}

	jobRes, err := svc.jobRepo.CreateJobAPI(&job.CreateJobRequest{
		ProductId: subscribedResp.Data.ProductId,
		MemberId:  authCtx.UserIdStr(),
		CompanyId: authCtx.CompanyIdStr(),
		Total:     1,
	})
	if err != nil {
		return nil, apperror.MapRepoError(err, constant.FailedCreateJob)
	}
	jobIdStr := helper.ConvertUintToString(jobRes.JobId)

	// todo: remove
	dummyTrxId := helper.GenerateTrx(constant.TrxIdPhoneNIK)
	if err := svc.dummyLogTrans(&phoneNIKContext{
		APIKey:         authCtx.APIKey,
		JobIdStr:       jobIdStr,
		MemberIdStr:    authCtx.UserIdStr(),
		CompanyIdStr:   authCtx.CompanyIdStr(),
		MemberId:       authCtx.UserId,
		CompanyId:      authCtx.CompanyId,
		ProductId:      subscribedResp.Data.ProductId,
		ProductGroupId: subscribedResp.Data.Product.ProductGroupId,
		JobId:          jobRes.JobId,
		Request:        reqBody,
	}, dummyTrxId); err != nil {
		return nil, apperror.MapRepoError(err, constant.FailedCreateJob)
	}

	result, err := svc.repo.PhoneToNIKAPI(authCtx.APIKey, dummyTrxId, reqBody)
	if err != nil {
		if err := svc.jobService.FinalizeFailedJob(jobIdStr); err != nil {
			return nil, err
		}

		var apiErr *apperror.ExternalAPIError
		if errors.As(err, &apiErr) {
			return nil, apperror.MapLoanError(apiErr)
		}

		return nil, apperror.Internal("failed to process recycle number", err)
	}

	if err := svc.jobService.FinalizeJob(jobIdStr); err != nil {
		return nil, err
	}

	if err := svc.operationRepo.AddLogOperation(&operation.AddLogRequest{
		MemberId:  authCtx.UserId,
		CompanyId: authCtx.CompanyId,
		Action:    constant.EventPhoneToNIKSingleReq,
	}); err != nil {
		log.Warn().
			Err(err).
			Str("action", constant.EventPhoneToNIKSingleReq).
			Msg("failed to add operation log")
	}

	return result, nil
}

// todo: remove
func (svc *service) dummyLogTrans(params *phoneNIKContext, dummyTrxId string) error {
	status := "not match"
	if params.Request.Phone == "08111111110" && params.Request.NIK == "3576014403910003" {
		status = "match"
	}

	return svc.transactionRepo.CreateLogTransAPI(&transaction.LogTransProCatRequest{
		TransactionID:  dummyTrxId,
		MemberID:       params.MemberId,
		CompanyID:      params.CompanyId,
		ProductID:      params.ProductId,
		ProductGroupID: params.ProductGroupId,
		JobID:          params.JobId,
		Message:        constant.Success,
		Status:         http.StatusOK,
		Success:        true,
		LoanNo:         params.Request.LoanNo,
		ResponseBody: &transaction.ResponseBody{
			Data: dataPhoneNIKAPI{
				Status: status,
			},
			Input:           params.Request,
			TransactionId:   dummyTrxId,
			PricingStrategy: "FREE",
			DateTime:        time.Now().Format(constant.FormatDateAndTime),
		},
		RequestBody:  params.Request,
		RequestTime:  time.Now(),
		ResponseTime: time.Now(),
	})
}
