package multipleloan

import (
	"errors"
	"front-office/internal/core/log/transaction"
	"front-office/internal/core/product"
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
	productRepo product.Repository,
	jobRepo job.Repository,
	transactionRepo transaction.Repository,
	jobService job.Service,
) Service {
	return &service{
		repo,
		productRepo,
		jobRepo,
		transactionRepo,
		jobService,
	}
}

type service struct {
	repo            Repository
	productRepo     product.Repository
	jobRepo         job.Repository
	transactionRepo transaction.Repository
	jobService      job.Service
}

type Service interface {
	MultipleLoan(apiKey, slug, memberId, companyId string, reqBody *multipleLoanRequest) (*model.ProCatAPIResponse[dataMultipleLoanResponse], error)
	BulkMultipleLoan(apiKey, slug string, memberId, companyId uint, file *multipart.FileHeader) error
}

type multipleLoanFunc func(string, string, string, string, *multipleLoanRequest) (*model.ProCatAPIResponse[dataMultipleLoanResponse], error)

func (svc *service) MultipleLoan(apiKey, slug, memberId, companyId string, reqBody *multipleLoanRequest) (*model.ProCatAPIResponse[dataMultipleLoanResponse], error) {
	productSlug, err := mapProductSlug(slug)
	if err != nil {
		return nil, apperror.BadRequest("unsupported product slug")
	}

	product, err := svc.productRepo.GetProductAPI(productSlug)
	if err != nil {
		return nil, apperror.MapRepoError(err, constant.FailedFetchProduct)
	}
	if product.ProductId == 0 {
		return nil, apperror.NotFound(constant.ProductNotFound)
	}

	jobRes, err := svc.jobRepo.CreateJobAPI(&job.CreateJobRequest{
		ProductId: product.ProductId,
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
		return nil, apperror.BadRequest("unsupported product type")
	}

	result, err := handler(apiKey, jobIdStr, memberId, companyId, reqBody)
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

func (svc *service) BulkMultipleLoan(apiKey, slug string, memberId, companyId uint, file *multipart.FileHeader) error {
	productSlug, err := mapProductSlug(slug)
	if err != nil {
		return apperror.BadRequest("unsupported product slug")
	}

	product, err := svc.productRepo.GetProductAPI(productSlug)
	if err != nil {
		return apperror.MapRepoError(err, constant.FailedFetchProduct)
	}
	if product.ProductId == 0 {
		return apperror.NotFound(constant.ProductNotFound)
	}

	if err := helper.ValidateUploadedFile(file, 30*1024*1024, []string{".csv"}); err != nil {
		return apperror.BadRequest(err.Error())
	}

	records, err := helper.ParseCSVFile(file, []string{"ID Card Number", "Phone Number"})
	if err != nil {
		return apperror.Internal(constant.FailedParseCSV, err)
	}

	memberIdStr := strconv.Itoa(int(memberId))
	companyIdStr := strconv.Itoa(int(companyId))
	jobRes, err := svc.jobRepo.CreateJobAPI(&job.CreateJobRequest{
		ProductId: product.ProductId,
		MemberId:  memberIdStr,
		CompanyId: companyIdStr,
		Total:     len(records) - 1,
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
				ProductId:      product.ProductId,
				ProductGroupId: product.ProductGroupId,
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
	if err := validator.ValidateStruct(params.Request); err != nil {
		_ = svc.transactionRepo.CreateLogTransAPI(&transaction.LogTransProCatRequest{
			MemberID:       params.MemberId,
			CompanyID:      params.CompanyId,
			ProductID:      params.ProductId,
			ProductGroupID: params.ProductGroupId,
			JobID:          params.JobId,
			Message:        err.Error(),
			Status:         http.StatusBadRequest,
			Success:        false,
			ResponseBody: &transaction.ResponseBody{
				Input:    params.Request,
				DateTime: time.Now().Format(constant.FormatDateAndTime),
			},
			Data:        nil,
			RequestBody: params.Request,
		})

		return apperror.BadRequest(err.Error())
	}

	handlers := map[string]multipleLoanFunc{
		constant.SlugMultipleLoan7Days:  svc.repo.CallMultipleLoan7Days,
		constant.SlugMultipleLoan30Days: svc.repo.CallMultipleLoan30Days,
		constant.SlugMultipleLoan90Days: svc.repo.CallMultipleLoan90Days,
	}

	handler, ok := handlers[params.ProductSlug]
	if !ok {
		return apperror.BadRequest("unsupported product type")
	}

	result, err := handler(params.APIKey, params.JobIdStr, params.MemberIdStr, params.CompanyIdStr, params.Request)
	if err != nil {
		if err := svc.transactionRepo.CreateLogTransAPI(&transaction.LogTransProCatRequest{
			MemberID:       params.MemberId,
			CompanyID:      params.CompanyId,
			ProductID:      params.ProductId,
			ProductGroupID: params.ProductGroupId,
			JobID:          params.JobId,
			Message:        result.Message,
			Status:         result.StatusCode,
			Success:        false,
			ResponseBody: &transaction.ResponseBody{
				Input:    params.Request,
				DateTime: time.Now().Format(constant.FormatDateAndTime),
			},
			Data:         nil,
			RequestBody:  params.Request,
			RequestTime:  time.Now(),
			ResponseTime: time.Now(),
		}); err != nil {
			return err
		}

		if err := svc.jobService.FinalizeFailedJob(params.JobIdStr); err != nil {
			return err
		}

		var apiErr *apperror.ExternalAPIError
		if errors.As(err, &apiErr) {
			return apperror.MapLoanError(apiErr)
		}

		return apperror.Internal("failed to process multiple loan", err)
	}

	if err := svc.transactionRepo.UpdateLogTransAPI(result.TransactionId, map[string]interface{}{
		"success": helper.BoolPtr(true),
	}); err != nil {
		return apperror.MapRepoError(err, "failed to update transaction job")
	}

	return nil
}
