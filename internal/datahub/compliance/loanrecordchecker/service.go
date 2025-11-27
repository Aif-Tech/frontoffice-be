package loanrecordchecker

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
	jobService job.Service,
) Service {
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
	LoanRecordChecker(authCtx *model.AuthContext, reqBody *loanRecordCheckerRequest) (*model.ProCatAPIResponse[dataLoanRecord], error)
	BulkLoanRecordChecker(authCtx *model.AuthContext, file *multipart.FileHeader) error
}

func (svc *service) LoanRecordChecker(authCtx *model.AuthContext, reqBody *loanRecordCheckerRequest) (*model.ProCatAPIResponse[dataLoanRecord], error) {
	subscribedResp, err := svc.memberRepo.GetSubscribedProducts(authCtx.CompanyIdStr(), constant.SlugLoanRecordChecker)
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

	result, err := svc.repo.LoanRecordCheckerAPI(authCtx.APIKey, jobIdStr, authCtx.UserIdStr(), authCtx.CompanyIdStr(), reqBody)
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

	if err := svc.jobService.FinalizeJob(jobIdStr); err != nil {
		return nil, err
	}

	if err := svc.operationRepo.AddLogOperation(&operation.AddLogRequest{
		MemberId:  authCtx.UserId,
		CompanyId: authCtx.CompanyId,
		Action:    constant.EventLoanRecordSingleReq,
	}); err != nil {
		log.Warn().
			Err(err).
			Str("action", constant.EventLoanRecordSingleReq).
			Msg("failed to add operation log")
	}

	return result, nil
}

func (svc *service) BulkLoanRecordChecker(authCtx *model.AuthContext, file *multipart.FileHeader) error {
	records, err := helper.ParseCSVFile(file, constant.CSVTemplateHeaderLoanRecord)
	if err != nil {
		return apperror.BadRequest(err.Error())
	}

	subscribedResp, err := svc.memberRepo.GetSubscribedProducts(authCtx.CompanyIdStr(), constant.SlugLoanRecordChecker)
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

	var loanCheckerReqs []*loanRecordCheckerRequest
	for i, rec := range records {
		if i == 0 {
			continue
		} // skip header
		loanCheckerReqs = append(loanCheckerReqs, &loanRecordCheckerRequest{
			Name:   rec[0],
			Nik:    rec[1],
			Phone:  rec[2],
			LoanNo: rec[3],
		})
	}

	var (
		wg         sync.WaitGroup
		errChan    = make(chan error, len(loanCheckerReqs))
		batchCount = 0
	)

	for _, req := range loanCheckerReqs {
		wg.Add(1)

		go func(loanCheckerReq *loanRecordCheckerRequest) {
			defer wg.Done()

			if err := svc.processSingleLoanRecord(&loanCheckerContext{
				APIKey:         authCtx.APIKey,
				JobIdStr:       jobIdStr,
				MemberIdStr:    authCtx.UserIdStr(),
				CompanyIdStr:   authCtx.CompanyIdStr(),
				MemberId:       authCtx.UserId,
				CompanyId:      authCtx.CompanyId,
				ProductId:      subscribedResp.Data.ProductId,
				ProductGroupId: subscribedResp.Data.Product.ProductGroupId,
				JobId:          jobRes.JobId,
				Request:        loanCheckerReq,
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
		log.Error().Err(err).Msg("error during bulk loan record checker processing")
	}

	if err := svc.operationRepo.AddLogOperation(&operation.AddLogRequest{
		MemberId:  authCtx.UserId,
		CompanyId: authCtx.CompanyId,
		Action:    constant.EventLoanRecordBulkReq,
	}); err != nil {
		log.Warn().
			Err(err).
			Str("action", constant.EventLoanRecordBulkReq).
			Msg("failed to add operation log")
	}

	return svc.jobService.FinalizeJob(jobIdStr)
}

func (svc *service) processSingleLoanRecord(params *loanCheckerContext) error {
	trxId := helper.GenerateTrx(constant.TrxIdLoanRecord)
	if err := validator.ValidateStruct(params.Request); err != nil {
		_ = svc.logFailedTransaction(params, trxId, err.Error(), http.StatusBadRequest)

		return apperror.BadRequest(err.Error())
	}

	_, err := svc.repo.LoanRecordCheckerAPI(params.APIKey, params.JobIdStr, params.MemberIdStr, params.CompanyIdStr, params.Request)
	if err != nil {
		_ = svc.logFailedTransaction(params, trxId, err.Error(), http.StatusBadGateway)
		_ = svc.jobService.FinalizeFailedJob(params.JobIdStr)

		return apperror.Internal("failed to process loan record checker", err)
	}

	return nil
}

func (svc *service) logFailedTransaction(params *loanCheckerContext, trxId, msg string, status int) error {
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
