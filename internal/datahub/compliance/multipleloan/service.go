package multipleloan

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
	MultipleLoan(apiKey, slug, memberId, companyId string, reqBody *multipleLoanRequest) (*model.ProCatAPIResponse[dataMultipleLoanResponse], error)
	BulkMultipleLoan(apiKey, quotaType, slug string, memberId, companyId uint, file *multipart.FileHeader) error
}

type multipleLoanFunc func(string, string, string, string, *multipleLoanRequest) (*model.ProCatAPIResponse[dataMultipleLoanResponse], error)

func (svc *service) MultipleLoan(apiKey, slug, memberId, companyId string, reqBody *multipleLoanRequest) (*model.ProCatAPIResponse[dataMultipleLoanResponse], error) {
	productSlug, err := mapProductSlug(slug)
	if err != nil {
		return nil, apperror.BadRequest("unsupported product slug")
	}

	subscribedResp, err := svc.memberRepo.GetSubscribedProducts(companyId, productSlug)
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

	handlers := map[string]multipleLoanFunc{
		constant.SlugMultipleLoan7Days:  svc.repo.CallMultipleLoan7Days,
		constant.SlugMultipleLoan30Days: svc.repo.CallMultipleLoan30Days,
		constant.SlugMultipleLoan90Days: svc.repo.CallMultipleLoan90Days,
	}

	handler, ok := handlers[productSlug]
	if !ok {
		return nil, apperror.BadRequest(constant.ErrUnsupportedProduct)
	}

	result, err := handler(apiKey, jobIdStr, memberId, companyId, reqBody)
	if err != nil {
		if err := svc.jobService.FinalizeFailedJob(jobIdStr); err != nil {
			return nil, err
		}

		return nil, apperror.Internal("failed to process multiple loan checker", err)
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

func (svc *service) BulkMultipleLoan(apiKey, quotaType, slug string, memberId, companyId uint, file *multipart.FileHeader) error {
	records, err := helper.ParseCSVFile(file, []string{"ID Card Number", "Phone Number"})
	if err != nil {
		return apperror.BadRequest(err.Error())
	}

	productSlug, err := mapProductSlug(slug)
	if err != nil {
		return apperror.BadRequest("unsupported product slug")
	}

	memberIdStr := strconv.Itoa(int(memberId))
	companyIdStr := strconv.Itoa(int(companyId))
	subscribedResp, err := svc.memberRepo.GetSubscribedProducts(companyIdStr, productSlug)
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

	var multipleLoanReqs []*multipleLoanRequest
	for i, rec := range records {
		if i == 0 {
			continue
		} // skip header
		multipleLoanReqs = append(multipleLoanReqs, &multipleLoanRequest{
			Nik: rec[0], Phone: rec[1],
		})
	}

	var (
		wg         sync.WaitGroup
		errChan    = make(chan error, len(multipleLoanReqs))
		batchCount = 0
	)

	for _, req := range multipleLoanReqs {
		wg.Add(1)

		go func(multipleLoanReq *multipleLoanRequest) {
			defer wg.Done()

			if err := svc.processMultipleLoan(&multipleLoanContext{
				APIKey:         apiKey,
				JobIdStr:       jobIdStr,
				MemberIdStr:    memberIdStr,
				CompanyIdStr:   companyIdStr,
				ProductSlug:    productSlug,
				MemberId:       memberId,
				CompanyId:      companyId,
				ProductId:      subscribedResp.Data.ProductId,
				ProductGroupId: subscribedResp.Data.Product.ProductGroupId,
				JobId:          jobRes.JobId,
				Request:        multipleLoanReq,
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
		logger.Error().Err(err).Msg("error during bulk multiple loan processing")
	}

	return svc.jobService.FinalizeJob(jobIdStr)
}

func (svc *service) processMultipleLoan(params *multipleLoanContext) error {
	type loanHandler struct {
		TrxPrefix string
		Func      multipleLoanFunc
	}

	handlers := map[string]loanHandler{
		constant.SlugMultipleLoan7Days:  {constant.TrxIdMultipleLoan7Days, svc.repo.CallMultipleLoan7Days},
		constant.SlugMultipleLoan30Days: {constant.TrxIdMultipleLoan30Days, svc.repo.CallMultipleLoan30Days},
		constant.SlugMultipleLoan90Days: {constant.TrxIdMultipleLoan30Days, svc.repo.CallMultipleLoan90Days},
	}

	h, ok := handlers[params.ProductSlug]
	if !ok {
		return apperror.BadRequest(constant.ErrUnsupportedProduct)
	}

	trxId := helper.GenerateTrx(h.TrxPrefix)
	if err := validator.ValidateStruct(params.Request); err != nil {
		_ = svc.logFailedTransaction(params, trxId, err.Error(), http.StatusBadRequest)

		return apperror.BadRequest(err.Error())
	}

	result, err := h.Func(params.APIKey, params.JobIdStr, params.MemberIdStr, params.CompanyIdStr, params.Request)
	if err != nil {
		_ = svc.logFailedTransaction(params, trxId, err.Error(), http.StatusBadGateway)
		_ = svc.jobService.FinalizeFailedJob(params.JobIdStr)

		return apperror.Internal("failed to process multiple loan", err)
	}

	if err := svc.transactionRepo.UpdateLogTransAPI(result.TransactionId, map[string]interface{}{
		"success": helper.BoolPtr(true),
	}); err != nil {
		return apperror.MapRepoError(err, "failed to update transaction job")
	}

	return nil
}

func (svc *service) logFailedTransaction(params *multipleLoanContext, trxId, msg string, status int) error {
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
