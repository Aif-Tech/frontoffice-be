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
	"mime/multipart"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/usepzaka/validator"
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
	PhoneNIK(authCtx *model.AuthContext, reqBody *phoneNIKRequest) (*model.ProCatAPIResponse[dataPhoneNIKAPI], error)
	BulkPhoneNIK(authCtx *model.AuthContext, fileHeader *multipart.FileHeader) error
}

func (svc *service) PhoneNIK(authCtx *model.AuthContext, reqBody *phoneNIKRequest) (*model.ProCatAPIResponse[dataPhoneNIKAPI], error) {
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

func (svc *service) BulkPhoneNIK(authCtx *model.AuthContext, file *multipart.FileHeader) error {
	records, err := helper.ParseCSVFile(file, constant.CSVTemplateHeaderPhoneNIK)
	if err != nil {
		return apperror.BadRequest(err.Error())
	}

	subscribedResp, err := svc.memberRepo.GetSubscribedProducts(authCtx.CompanyIdStr(), constant.SlugPhoneNIKMatching)
	if err != nil {
		return apperror.MapRepoError(err, constant.ErrFetchSubscribedProduct)
	}

	subscribedIdStr := strconv.Itoa(int(subscribedResp.Data.SubsribedProductID))
	quotaResp, err := svc.memberRepo.GetQuotaAPI(&member.QuotaParams{
		MemberId:     authCtx.UserIdStr(),
		CompanyId:    authCtx.CompanyIdStr(),
		SubscribedId: subscribedIdStr,
		QuotaType:    authCtx.QuotaTypeStr(),
	})
	if err != nil {
		return apperror.MapRepoError(err, constant.FailedFetchQuota)
	}

	totalRequests := len(records) - 1
	if authCtx.QuotaTypeStr() != "0" && quotaResp.Data.Quota < totalRequests {
		return apperror.Forbidden(constant.ErrQuotaExceeded)
	}

	jobRes, err := svc.jobRepo.CreateJobAPI(&job.CreateJobRequest{
		ProductId: subscribedResp.Data.ProductId,
		MemberId:  authCtx.UserIdStr(),
		CompanyId: authCtx.CompanyIdStr(),
		Total:     totalRequests,
	})
	if err != nil {
		return apperror.MapRepoError(err, constant.FailedCreateJob)
	}
	jobIdStr := helper.ConvertUintToString(jobRes.JobId)

	var requests []*phoneNIKRequest
	for i, rec := range records {
		if i == 0 {
			continue
		}

		requests = append(requests, &phoneNIKRequest{
			Phone:  rec[0],
			LoanNo: rec[1],
		})
	}

	var (
		wg         sync.WaitGroup
		errChan    = make(chan error, len(requests))
		batchCount = 0
	)

	for _, req := range requests {
		wg.Add(1)

		go func(req *phoneNIKRequest) {
			defer wg.Done()

			if err := svc.processSinglePhoneNIK(&phoneNIKContext{
				APIKey:         authCtx.APIKey,
				JobIdStr:       jobIdStr,
				MemberIdStr:    authCtx.UserIdStr(),
				CompanyIdStr:   authCtx.CompanyIdStr(),
				MemberId:       authCtx.UserId,
				CompanyId:      authCtx.CompanyId,
				ProductId:      subscribedResp.Data.ProductId,
				ProductGroupId: subscribedResp.Data.Product.ProductGroupId,
				JobId:          jobRes.JobId,
				Request:        req,
			}); err != nil {
				errChan <- err
			}
		}(req)

		batchCount++
		if batchCount == 100 {
			time.Sleep(time.Second)
			batchCount = 0
		}
	}

	wg.Wait()
	close(errChan)

	for err := range errChan {
		log.Error().Err(err).Str("job_id", jobIdStr).Msg("error during bulk phone to nik processing")
	}

	if err := svc.jobService.FinalizeJob(jobIdStr); err != nil {
		return err
	}

	if err := svc.operationRepo.AddLogOperation(&operation.AddLogRequest{
		MemberId:  authCtx.UserId,
		CompanyId: authCtx.CompanyId,
		Action:    constant.EventPhoneToNIKBulkReq,
	}); err != nil {
		log.Warn().
			Err(err).
			Str("action", constant.EventPhoneToNIKBulkReq).
			Msg("failed to add operation log")
	}

	return nil
}

func (svc *service) processSinglePhoneNIK(params *phoneNIKContext) error {
	trxId := helper.GenerateTrx(constant.TrxIdPhoneNIK)
	if err := validator.ValidateStruct(params.Request); err != nil {
		_ = svc.logFailedTransaction(params, trxId, err.Error(), http.StatusBadRequest)

		return apperror.BadRequest(err.Error())
	}

	if err := svc.dummyLogTrans(params, trxId); err != nil {
		return apperror.MapRepoError(err, constant.FailedCreateJob)
	}

	_, err := svc.repo.PhoneToNIKAPI(params.APIKey, trxId, params.Request)
	if err != nil {
		_ = svc.logFailedTransaction(params, trxId, err.Error(), http.StatusBadGateway)
		_ = svc.jobService.FinalizeFailedJob(params.JobIdStr)

		return apperror.Internal("failed to process recycle number", err)
	}

	return nil
}

func (svc *service) logFailedTransaction(params *phoneNIKContext, trxId, msg string, status int) error {
	return svc.transactionRepo.CreateLogTransAPI(&transaction.LogTransProCatRequest{
		TransactionID:  trxId,
		MemberID:       params.MemberId,
		CompanyID:      params.CompanyId,
		ProductID:      params.ProductId,
		ProductGroupID: params.ProductGroupId,
		JobID:          params.JobId,
		Message:        msg,
		Status:         status,
		Success:        false,
		LoanNo:         params.Request.LoanNo,
		ResponseBody: &transaction.ResponseBody{
			Input:    params.Request,
			DateTime: time.Now().Format(constant.FormatDateAndTime),
		},
		RequestBody:  params.Request,
		RequestTime:  time.Now(),
		ResponseTime: time.Now(),
	})
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
