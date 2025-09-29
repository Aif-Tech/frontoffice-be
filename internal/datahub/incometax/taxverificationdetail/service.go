package taxverificationdetail

import (
	"front-office/internal/core/log/transaction"
	"front-office/internal/core/member"
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
	memberRepo member.Repository,
	jobRepo job.Repository,
	transactionRepo transaction.Repository,
	jobService job.Service,
) Service {
	return &service{
		repo,
		productRepo,
		memberRepo,
		jobRepo,
		transactionRepo,
		jobService,
	}
}

type service struct {
	repo            Repository
	productRepo     product.Repository
	memberRepo      member.Repository
	jobRepo         job.Repository
	transactionRepo transaction.Repository
	jobService      job.Service
}

type Service interface {
	CallTaxVerification(apiKey, memberId, companyId string, request *taxVerificationRequest) (*model.ProCatAPIResponse[taxVerificationRespData], error)
	BulkTaxVerification(apiKey, quotaType string, memberId, companyId uint, file *multipart.FileHeader) error
}

func (svc *service) CallTaxVerification(apiKey, memberId, companyId string, request *taxVerificationRequest) (*model.ProCatAPIResponse[taxVerificationRespData], error) {
	subscribedResp, err := svc.memberRepo.GetSubscribedProducts(companyId, constant.SlugTaxVerificationDetail)
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

	result, err := svc.repo.TaxVerificationAPI(apiKey, jobIdStr, request)
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

func (svc *service) BulkTaxVerification(apiKey, quotaType string, memberId, companyId uint, file *multipart.FileHeader) error {
	if err := helper.ValidateUploadedFile(file, 30*1024*1024, []string{".csv"}); err != nil {
		return apperror.BadRequest(err.Error())
	}

	records, err := helper.ParseCSVFile(file, []string{"ID Card Number"})
	if err != nil {
		return apperror.Internal(constant.FailedParseCSV, err)
	}

	memberIdStr := strconv.Itoa(int(memberId))
	companyIdStr := strconv.Itoa(int(companyId))
	subscribedResp, err := svc.memberRepo.GetSubscribedProducts(companyIdStr, constant.SlugTaxVerificationDetail)
	if err != nil {
		return apperror.MapRepoError(err, constant.ErrFetchSubscribedProduct)
	}
	if subscribedResp.Data.ProductId == 0 {
		return apperror.NotFound(constant.ErrSubscribtionNotFound)
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

	var taxVerificationRequests []*taxVerificationRequest
	for i, record := range records {
		if i == 0 {
			continue
		}

		taxVerificationRequests = append(taxVerificationRequests, &taxVerificationRequest{
			NpwpOrNik: record[0],
		})
	}

	var (
		wg         sync.WaitGroup
		errChan    = make(chan error, len(taxVerificationRequests))
		batchCount = 0
	)

	for _, req := range taxVerificationRequests {
		wg.Add(1)

		go func(taxScoreReq *taxVerificationRequest) {
			defer wg.Done()

			if err := svc.processTaxVerification(&taxVerificationContext{
				APIKey:         apiKey,
				JobIdStr:       jobIdStr,
				MemberIdStr:    memberIdStr,
				CompanyIdStr:   companyIdStr,
				MemberId:       memberId,
				CompanyId:      companyId,
				ProductId:      subscribedResp.Data.ProductId,
				ProductGroupId: subscribedResp.Data.Product.ProductGroupId,
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

func (svc *service) processTaxVerification(params *taxVerificationContext) error {
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

	result, err := svc.repo.TaxVerificationAPI(
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
