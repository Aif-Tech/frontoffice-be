package taxcompliancestatus

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
	TaxComplianceStatus(authCtx *model.AuthContext, reqBody *taxComplianceStatusRequest) (*model.ProCatAPIResponse[taxComplianceRespData], error)
	BulkTaxComplianceStatus(apiKey, quotaType string, memberId, companyId uint, file *multipart.FileHeader) error
}

func (svc *service) TaxComplianceStatus(authCtx *model.AuthContext, reqBody *taxComplianceStatusRequest) (*model.ProCatAPIResponse[taxComplianceRespData], error) {
	subscribedResp, err := svc.memberRepo.GetSubscribedProducts(authCtx.CompanyIdStr(), constant.SlugTaxComplianceStatus)
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

	result, err := svc.repo.TaxComplianceStatusAPI(authCtx.APIKey, jobIdStr, reqBody)
	if err != nil {
		if err := svc.jobService.FinalizeFailedJob(jobIdStr); err != nil {
			return nil, err
		}

		return nil, apperror.MapRepoError(err, "failed to process tax compliance status")
	}

	if err := svc.jobService.FinalizeJob(jobIdStr); err != nil {
		return nil, err
	}

	if err := svc.operationRepo.AddLogOperation(&operation.AddLogRequest{
		MemberId:  authCtx.UserId,
		CompanyId: authCtx.CompanyId,
		Action:    constant.EventTaxComplianceSingleReq,
	}); err != nil {
		log.Warn().
			Err(err).
			Str("action", constant.EventTaxComplianceSingleReq).
			Msg("failed to add operation log")
	}

	return result, nil
}

func (svc *service) BulkTaxComplianceStatus(apiKey, quotaType string, memberId, companyId uint, file *multipart.FileHeader) error {
	records, err := helper.ParseCSVFile(file, constant.CSVTemplateHeaderTaxCompliance)
	if err != nil {
		return apperror.BadRequest(err.Error())
	}

	memberIdStr := strconv.Itoa(int(memberId))
	companyIdStr := strconv.Itoa(int(companyId))
	subscribedResp, err := svc.memberRepo.GetSubscribedProducts(companyIdStr, constant.SlugTaxComplianceStatus)
	if err != nil {
		return apperror.MapRepoError(err, constant.ErrFetchSubscribedProduct)
	}

	subsribedIdStr := strconv.Itoa(int(subscribedResp.Data.SubsribedProductID))
	quotaResp, err := svc.memberRepo.GetQuotaAPI(&member.QuotaParams{
		MemberId:     memberIdStr,
		CompanyId:    companyIdStr,
		SubscribedId: subsribedIdStr,
		QuotaType:    quotaType,
	})
	if err != nil {
		return apperror.MapRepoError(err, constant.FailedFetchQuota)
	}

	totalRequest := len(records) - 1
	if quotaType != "0" && quotaResp.Data.Quota < totalRequest {
		return apperror.Forbidden(constant.ErrQuotaExceeded)
	}

	jobRes, err := svc.jobRepo.CreateJobAPI(&job.CreateJobRequest{
		ProductId: subscribedResp.Data.ProductId,
		MemberId:  memberIdStr,
		CompanyId: companyIdStr,
		Total:     totalRequest,
	})
	if err != nil {
		return apperror.MapRepoError(err, constant.FailedCreateJob)
	}
	jobIdStr := helper.ConvertUintToString(jobRes.JobId)

	var taxComplianceReqs []*taxComplianceStatusRequest
	for i, record := range records {
		if i == 0 {
			continue
		}

		taxComplianceReqs = append(taxComplianceReqs, &taxComplianceStatusRequest{
			Npwp: record[0],
			// LoanNo: record[1],
		})
	}

	var (
		wg         sync.WaitGroup
		errChan    = make(chan error, len(taxComplianceReqs))
		batchCount = 0
	)

	for _, req := range taxComplianceReqs {
		wg.Add(1)

		go func(taxComplianceReq *taxComplianceStatusRequest) {
			defer wg.Done()

			if err := svc.processTaxComplianceStatus(&taxComplianceContext{
				APIKey:         apiKey,
				JobIdStr:       jobIdStr,
				MemberIdStr:    memberIdStr,
				CompanyIdStr:   companyIdStr,
				MemberId:       memberId,
				CompanyId:      companyId,
				ProductId:      subscribedResp.Data.ProductId,
				ProductGroupId: subscribedResp.Data.Product.ProductGroupId,
				JobId:          jobRes.JobId,
				Request:        taxComplianceReq,
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
		log.Error().Err(err).Msg("error during bulk tax compliance status prrocessing")
	}

	return svc.jobService.FinalizeJob(jobIdStr)
}

func (svc *service) processTaxComplianceStatus(params *taxComplianceContext) error {
	trxId := helper.GenerateTrx(constant.TrxIdTaxCompliance)
	if err := validator.ValidateStruct(params.Request); err != nil {
		_ = svc.logFailedTransaction(params, trxId, err.Error(), http.StatusBadRequest)

		return apperror.BadRequest(err.Error())
	}

	_, err := svc.repo.TaxComplianceStatusAPI(
		params.APIKey,
		params.JobIdStr,
		params.Request,
	)
	if err != nil {
		_ = svc.logFailedTransaction(params, trxId, err.Error(), http.StatusBadGateway)
		_ = svc.jobService.FinalizeFailedJob(params.JobIdStr)

		return apperror.Internal("failed to process tax compliance status", err)
	}

	return nil
}

func (svc *service) logFailedTransaction(params *taxComplianceContext, trxId, msg string, status int) error {
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
		// LoanNo:         params.Request.LoanNo,
		ResponseBody: &transaction.ResponseBody{
			Input:    params.Request,
			DateTime: time.Now().Format(constant.FormatDateAndTime),
		},
		RequestBody:  params.Request,
		RequestTime:  time.Now(),
		ResponseTime: time.Now(),
	})
}
