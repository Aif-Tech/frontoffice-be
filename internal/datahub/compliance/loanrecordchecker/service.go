package loanrecordchecker

import (
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

	logger "github.com/rs/zerolog/log"
	"github.com/usepzaka/validator"
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
	LoanRecordChecker(apiKey, memberId, companyId string, reqBody *loanRecordCheckerRequest) (*model.ProCatAPIResponse[dataLoanRecord], error)
	BulkLoanRecordChecker(apiKey, quotaType string, memberId, companyId uint, file *multipart.FileHeader) error
}

func (svc *service) LoanRecordChecker(apiKey, memberId, companyId string, reqBody *loanRecordCheckerRequest) (*model.ProCatAPIResponse[dataLoanRecord], error) {
	subscribedResp, err := svc.memberRepo.GetSubscribedProducts(companyId, constant.SlugLoanRecordChecker)
	if err != nil {
		return nil, apperror.MapRepoError(err, constant.ErrFetchSubscribedProduct)
	}
	if subscribedResp.Data.ProductId == 0 {
		return nil, apperror.NotFound(constant.ErrSubscribtionNotFound)
	}

	jobRes, err := svc.jobRepo.CreateJobAPI(&job.CreateJobRequest{
		ProductId: subscribedResp.Data.ProductId,
		MemberId:  memberId,
		CompanyId: companyId,
		Total:     1,
	})
	if err != nil {
		return nil, apperror.MapRepoError(err, constant.FailedCreateJob)
	}
	jobIdStr := helper.ConvertUintToString(jobRes.JobId)

	result, err := svc.repo.LoanRecordCheckerAPI(apiKey, jobIdStr, memberId, companyId, reqBody)
	if err != nil {
		if err := svc.jobService.FinalizeFailedJob(jobIdStr); err != nil {
			return nil, err
		}

		return nil, apperror.Internal("failed to process loan record checker", err)
	}

	if err := svc.transactionRepo.UpdateLogTransAPI(result.TransactionId, map[string]interface{}{
		"success": helper.BoolPtr(true),
	}); err != nil {
		return nil, apperror.MapRepoError(err, "failed to update transaction log")
	}

	if err := svc.jobService.FinalizeJob(jobIdStr); err != nil {
		return nil, err
	}

	return result, nil
}

func (svc *service) BulkLoanRecordChecker(apiKey, quotaType string, memberId, companyId uint, file *multipart.FileHeader) error {
	records, err := helper.ParseCSVFile(file, []string{"Name", "ID Card Number", "Phone Number"})
	if err != nil {
		return apperror.BadRequest(err.Error())
	}

	memberIdStr := strconv.Itoa(int(memberId))
	companyIdStr := strconv.Itoa(int(companyId))
	subscribedResp, err := svc.memberRepo.GetSubscribedProducts(companyIdStr, constant.SlugLoanRecordChecker)
	if err != nil {
		return apperror.MapRepoError(err, constant.ErrFetchSubscribedProduct)
	}
	if subscribedResp.Data.ProductId == 0 {
		return apperror.NotFound(constant.ProductNotFound)
	}

	subscribedIdStr := strconv.Itoa(int(subscribedResp.Data.SubsribedProductID))
	quotaResp, err := svc.memberRepo.GetQuotaAPI(&member.QuotaParams{
		MemberId:     memberIdStr,
		CompanyId:    companyIdStr,
		SubscribedId: subscribedIdStr,
		QuotaType:    quotaType,
	})
	if err != nil {
		return apperror.MapRepoError(err, constant.FailedFetchQuota)
	}

	totalRequests := len(records) - 1
	if quotaType != "0" && quotaResp.Data.Quota < totalRequests {
		return apperror.Forbidden(constant.ErrQuotaExceeded)
	}

	jobRes, err := svc.jobRepo.CreateJobAPI(&job.CreateJobRequest{
		ProductId: subscribedResp.Data.ProductId,
		MemberId:  memberIdStr,
		CompanyId: companyIdStr,
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
			Name: rec[0], Nik: rec[1], Phone: rec[2],
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
				APIKey:         apiKey,
				JobIdStr:       jobIdStr,
				MemberIdStr:    memberIdStr,
				CompanyIdStr:   companyIdStr,
				MemberId:       memberId,
				CompanyId:      companyId,
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
		logger.Error().Err(err).Msg("error during bulk loan record checker processing")
	}

	return svc.jobService.FinalizeJob(jobIdStr)
}

func (svc *service) processSingleLoanRecord(params *loanCheckerContext) error {
	trxId := helper.GenerateTrx(constant.TrxIdLoanRecord)
	if err := validator.ValidateStruct(params.Request); err != nil {
		_ = svc.logFailedTransaction(params, trxId, err.Error(), http.StatusBadRequest)

		return apperror.BadRequest(err.Error())
	}

	result, err := svc.repo.LoanRecordCheckerAPI(params.APIKey, params.JobIdStr, params.MemberIdStr, params.CompanyIdStr, params.Request)
	if err != nil {
		_ = svc.logFailedTransaction(params, trxId, err.Error(), http.StatusBadGateway)
		_ = svc.jobService.FinalizeFailedJob(params.JobIdStr)

		return apperror.Internal("failed to process loan record checker", err)
	}

	if err := svc.transactionRepo.UpdateLogTransAPI(result.TransactionId, map[string]interface{}{
		"success": helper.BoolPtr(true),
	}); err != nil {
		return apperror.MapRepoError(err, "failed to update log transaction")
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
		ResponseBody: &transaction.ResponseBody{
			Input:    params.Request,
			DateTime: time.Now().Format(constant.FormatDateAndTime),
		},
		RequestBody:  params.Request,
		RequestTime:  time.Now(),
		ResponseTime: time.Now(),
	})
}
