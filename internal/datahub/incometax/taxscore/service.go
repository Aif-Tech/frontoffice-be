package taxscore

import (
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

	"github.com/rs/zerolog/log"
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
	TaxScore(apiKey, memberId, companyId string, request *taxScoreRequest) (*model.ProCatAPIResponse[taxScoreRespData], error)
	BulkTaxScore(apiKey string, memberId, companyId uint, file *multipart.FileHeader) error
}

func (svc *service) TaxScore(apiKey, memberId, companyId string, request *taxScoreRequest) (*model.ProCatAPIResponse[taxScoreRespData], error) {
	product, err := svc.productRepo.GetProductAPI(constant.SlugTaxScore)
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

	result, err := svc.repo.TaxScoreAPI(apiKey, jobIdStr, request)
	if err != nil {
		if err := svc.jobService.FinalizeFailedJob(jobIdStr); err != nil {
			return nil, err
		}

		return nil, apperror.MapRepoError(err, "failed to process tax score")
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

func (svc *service) BulkTaxScore(apiKey string, memberId, companyId uint, file *multipart.FileHeader) error {
	product, err := svc.productRepo.GetProductAPI(constant.SlugTaxScore)
	if err != nil {
		return apperror.MapRepoError(err, constant.FailedFetchProduct)
	}
	if product.ProductId == 0 {
		return apperror.NotFound(constant.ProductNotFound)
	}

	if err := helper.ValidateUploadedFile(file, 30*1024*1024, []string{".csv"}); err != nil {
		return apperror.BadRequest(err.Error())
	}

	records, err := helper.ParseCSVFile(file, []string{"NPWP"})
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

	var taxScoreReqs []*taxScoreRequest
	for i, record := range records {
		if i == 0 {
			continue
		}

		taxScoreReqs = append(taxScoreReqs, &taxScoreRequest{
			Npwp: record[0],
		})
	}

	var (
		wg         sync.WaitGroup
		errChan    = make(chan error, len(taxScoreReqs))
		batchCount = 0
	)

	for _, req := range taxScoreReqs {
		wg.Add(1)

		go func(taxScoreReq *taxScoreRequest) {
			defer wg.Done()

			if err := svc.processTaxScore(&taxScoreContext{
				APIKey:         apiKey,
				JobIdStr:       jobIdStr,
				MemberIdStr:    memberIdStr,
				CompanyIdStr:   companyIdStr,
				MemberId:       memberId,
				CompanyId:      companyId,
				ProductId:      product.ProductId,
				ProductGroupId: product.ProductGroupId,
				JobId:          jobRes.JobId,
				Request:        taxScoreReq,
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
		log.Error().Err(err).Msg("error during bulk tax score prrocessing")
	}

	return svc.jobService.FinalizeJob(jobIdStr)
}

func (svc *service) processTaxScore(params *taxScoreContext) error {
	if err := validator.ValidateStruct(params.Request); err != nil {
		_ = svc.transactionRepo.CreateLogTransAPI(&transaction.LogTransProCatRequest{
			MemberID:       params.MemberId,
			CompanyID:      params.CompanyId,
			ProductID:      params.ProductId,
			ProductGroupID: params.ProductGroupId,
			JobID:          params.JobId,
			Success:        false,
			Message:        err.Error(),
			Status:         http.StatusBadRequest,
			ResponseBody: &transaction.ResponseBody{
				Input:    params.Request,
				DateTime: time.Now().Format(constant.FormatDateAndTime),
			},
			Data:        nil,
			RequestBody: params.Request,
		})

		return apperror.BadRequest(err.Error())
	}

	result, err := svc.repo.TaxScoreAPI(
		params.APIKey,
		params.JobIdStr,
		params.Request,
	)
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

		return apperror.Internal("failed to process tax compliance status", err)
	}

	if err := svc.transactionRepo.UpdateLogTransAPI(result.TransactionId, map[string]interface{}{
		"success": helper.BoolPtr(true),
	}); err != nil {
		return apperror.MapRepoError(err, "failed to update log transaction")
	}

	return nil
}
