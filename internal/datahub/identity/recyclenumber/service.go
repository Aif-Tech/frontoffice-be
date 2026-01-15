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
	RecycleNumber(authCtx *model.AuthContext, reqBody *recycleNumberRequest) (*model.ProCatAPIResponse[dataRecycleNumberAPI], error)
	BulkRecycleNumber(authCtx *model.AuthContext, fileHeader *multipart.FileHeader) error
}

func (svc *service) RecycleNumber(authCtx *model.AuthContext, reqBody *recycleNumberRequest) (*model.ProCatAPIResponse[dataRecycleNumberAPI], error) {
	reqBody.Timestamp = time.Now().Format(constant.FormatYYYYMMDD)

	if err := validatePeriodByOperator(reqBody.Phone, reqBody.Period); err != nil {
		return nil, apperror.BadRequest(err.Error())
	}

	subscribedResp, err := svc.memberRepo.GetSubscribedProducts(authCtx.CompanyIdStr(), constant.SlugRecycleNumber)
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
	if err := svc.dummyLogTrans(&recycleNumberContext{
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
	}); err != nil {
		return nil, apperror.MapRepoError(err, constant.FailedCreateJob)
	}

	result, err := svc.repo.RecycleNumberAPI(authCtx.APIKey, jobIdStr, authCtx.UserIdStr(), authCtx.CompanyIdStr(), reqBody)
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
		Action:    constant.EventRecycleNumberSingleReq,
	}); err != nil {
		log.Warn().
			Err(err).
			Str("action", constant.EventRecycleNumberSingleReq).
			Msg("failed to add operation log")
	}

	return result, nil
}

func (svc *service) BulkRecycleNumber(authCtx *model.AuthContext, file *multipart.FileHeader) error {
	records, err := helper.ParseCSVFile(file, constant.CSVTemplateHeaderRecycleNumber)
	if err != nil {
		return apperror.BadRequest(err.Error())
	}

	subscribedResp, err := svc.memberRepo.GetSubscribedProducts(authCtx.CompanyIdStr(), constant.SlugRecycleNumber)
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

	var requests []*recycleNumberRequest
	for i, rec := range records {
		if i == 0 {
			continue
		}

		requests = append(requests, &recycleNumberRequest{
			Phone:  rec[0],
			LoanNo: rec[1],
			Period: rec[2],
		})
	}

	var (
		wg         sync.WaitGroup
		errChan    = make(chan error, len(requests))
		batchCount = 0
	)

	for _, req := range requests {
		wg.Add(1)

		go func(recycleNumberReq *recycleNumberRequest) {
			defer wg.Done()

			if err := svc.processSingleRecycleNumber(&recycleNumberContext{
				APIKey:         authCtx.APIKey,
				JobIdStr:       jobIdStr,
				MemberIdStr:    authCtx.UserIdStr(),
				CompanyIdStr:   authCtx.CompanyIdStr(),
				MemberId:       authCtx.UserId,
				CompanyId:      authCtx.CompanyId,
				ProductId:      subscribedResp.Data.ProductId,
				ProductGroupId: subscribedResp.Data.Product.ProductGroupId,
				JobId:          jobRes.JobId,
				Request:        recycleNumberReq,
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
		log.Error().Err(err).Str("job_id", jobIdStr).Msg("error during bulk recycle number processing")
	}

	if err := svc.jobService.FinalizeJob(jobIdStr); err != nil {
		return err
	}

	if err := svc.operationRepo.AddLogOperation(&operation.AddLogRequest{
		MemberId:  authCtx.UserId,
		CompanyId: authCtx.CompanyId,
		Action:    constant.EventRecycleNumberBulkReq,
	}); err != nil {
		log.Warn().
			Err(err).
			Str("action", constant.EventRecycleNumberBulkReq).
			Msg("failed to add operation log")
	}

	return nil
}

func (svc *service) processSingleRecycleNumber(params *recycleNumberContext) error {
	trxId := helper.GenerateTrx(constant.TrxIdLoanRecord)
	params.Request.Timestamp = time.Now().Format(constant.FormatYYYYMMDD)
	if err := validator.ValidateStruct(params.Request); err != nil {
		_ = svc.logFailedTransaction(params, trxId, err.Error(), http.StatusBadRequest)

		return apperror.BadRequest(err.Error())
	}

	if err := validatePeriodByOperator(params.Request.Phone, params.Request.Period); err != nil {
		_ = svc.logFailedTransaction(params, trxId, err.Error(), http.StatusBadRequest)

		return apperror.BadRequest(err.Error())
	}

	if err := svc.dummyLogTrans(params); err != nil {
		return apperror.MapRepoError(err, constant.FailedCreateJob)
	}

	_, err := svc.repo.RecycleNumberAPI(params.APIKey, params.JobIdStr, params.MemberIdStr, params.CompanyIdStr, params.Request)
	if err != nil {
		_ = svc.logFailedTransaction(params, trxId, err.Error(), http.StatusBadGateway)
		_ = svc.jobService.FinalizeFailedJob(params.JobIdStr)

		return apperror.Internal("failed to process recycle number", err)
	}

	return nil
}

func (svc *service) logFailedTransaction(params *recycleNumberContext, trxId, msg string, status int) error {
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

func validatePeriodByOperator(phone, period string) error {
	operator := detectOperator(phone)

	if operator != constant.OperatorIsat {
		return nil
	}

	if period == "" {
		return errors.New("period is required for Isat numbers")
	}

	if period != "30" && period != "90" {
		return errors.New("period must be either 30 or 90 for Isat numbers")
	}

	return nil
}

func detectOperator(phone string) string {
	if len(phone) < 4 {
		return "unknown"
	}

	prefix := phone[:4]
	if op, ok := constant.OperatorByPrefix[prefix]; ok {
		return op
	}

	return "unknown"
}

// todo: remove
func (svc *service) dummyLogTrans(params *recycleNumberContext) error {
	dummyTrxId := helper.GenerateTrx(constant.TrxIdRecycleNumber)

	status := "phone number has never been recycled"
	if params.Request.Phone == "085700000001" {
		status = "phone number never happens recycled"
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
			Data: dataRecycleNumberAPI{
				Status: status,
			},
			Input:    params.Request,
			DateTime: time.Now().Format(constant.FormatDateAndTime),
		},
		RequestBody:  params.Request,
		RequestTime:  time.Now(),
		ResponseTime: time.Now(),
	})
}
