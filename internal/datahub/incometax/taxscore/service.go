package taxscore

import (
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
	transactionRepo transaction.Repository,
	operationRepo operation.Repository,
	jobService job.Service,
) Service {
	return &service{
		repo,
		memberRepo,
		jobRepo,
		transactionRepo,
		operationRepo,
		jobService,
	}
}

type service struct {
	repo            Repository
	memberRepo      member.Repository
	jobRepo         job.Repository
	transactionRepo transaction.Repository
	operationRepo   operation.Repository
	jobService      job.Service
}

type Service interface {
	TaxScore(authCtx *model.AuthContext, request *taxScoreRequest) (*model.ProCatAPIResponse[taxScoreRespData], error)
	BulkTaxScore(apiKey, quotaType string, memberId, companyId uint, file *multipart.FileHeader) error
}

func (svc *service) TaxScore(authCtx *model.AuthContext, request *taxScoreRequest) (*model.ProCatAPIResponse[taxScoreRespData], error) {
	subscribedResp, err := svc.memberRepo.GetSubscribedProducts(authCtx.CompanyIdStr(), constant.SlugTaxScore)
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

	result, err := svc.repo.TaxScoreAPI(authCtx.APIKey, jobIdStr, request)
	if err != nil {
		if err := svc.jobService.FinalizeFailedJob(jobIdStr); err != nil {
			return nil, err
		}

		return nil, apperror.MapRepoError(err, "failed to process tax score")
	}

	if err := svc.jobService.FinalizeJob(jobIdStr); err != nil {
		return nil, err
	}

	if err := svc.operationRepo.AddLogOperation(&operation.AddLogRequest{
		MemberId:  authCtx.UserId,
		CompanyId: authCtx.CompanyId,
		Action:    constant.EventTaxScoreSingleReq,
	}); err != nil {
		log.Warn().
			Err(err).
			Str("action", constant.EventTaxScoreSingleReq).
			Msg("failed to add operation log")
	}

	return result, nil
}

func (svc *service) BulkTaxScore(apiKey, quotaType string, memberId, companyId uint, file *multipart.FileHeader) error {
	records, err := helper.ParseCSVFile(file, constant.CSVTemplateHeaderTaxScore)
	if err != nil {
		return apperror.BadRequest(err.Error())
	}

	memberIdStr := strconv.Itoa(int(memberId))
	companyIdStr := strconv.Itoa(int(companyId))
	subscribedResp, err := svc.memberRepo.GetSubscribedProducts(companyIdStr, constant.SlugTaxScore)
	if err != nil {
		return apperror.MapRepoError(err, constant.ErrFetchSubscribedProduct)
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
			Npwp:   record[0],
			LoanNo: record[1],
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

func (svc *service) processTaxScore(params *taxScoreContext) error {
	trxId := helper.GenerateTrx(constant.TrxIdTaxScore)
	if err := validator.ValidateStruct(params.Request); err != nil {
		_ = svc.logFailedTransaction(params, trxId, err.Error(), http.StatusBadRequest)

		return apperror.BadRequest(err.Error())
	}

	_, err := svc.repo.TaxScoreAPI(
		params.APIKey,
		params.JobIdStr,
		params.Request,
	)
	if err != nil {
		_ = svc.logFailedTransaction(params, trxId, err.Error(), http.StatusBadGateway)
		_ = svc.jobService.FinalizeFailedJob(params.JobIdStr)

		return apperror.Internal("failed to process tax score", err)
	}

	return nil
}

func (svc *service) logFailedTransaction(params *taxScoreContext, trxId, msg string, status int) error {
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
