package npwpverification

import (
	"front-office/internal/core/log/transaction"
	"front-office/internal/core/member"
	"front-office/internal/datahub/job"
	"front-office/pkg/apperror"
	"front-office/pkg/common/constant"
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
	NPWPVerification(apiKey, memberId, companyId string, payload *npwpVerificationRequest) error
	BulkNPWPVerification(apiKey, quotaType string, memberId, companyId uint, file *multipart.FileHeader) error
}

func (svc *service) NPWPVerification(apiKey, memberId, companyId string, payload *npwpVerificationRequest) error {
	subscribedResp, err := svc.memberRepo.GetSubscribedProducts(companyId, constant.SlugNPWPVerification)
	if err != nil {
		return apperror.MapRepoError(err, constant.ErrFetchSubscribedProduct)
	}

	jobRes, err := svc.jobRepo.CreateJobAPI(&job.CreateJobRequest{
		ProductId: subscribedResp.Data.ProductId,
		MemberId:  memberId,
		CompanyId: companyId,
		Total:     1,
	})
	if err != nil {
		return apperror.MapRepoError(err, constant.FailedCreateJob)
	}

	jobIdStr := helper.ConvertUintToString(jobRes.JobId)
	_, err = svc.repo.NPWPVerificationAPI(apiKey, jobIdStr, payload)
	if err != nil {
		if err := svc.jobService.FinalizeFailedJob(jobIdStr); err != nil {
			return err
		}

		return apperror.Internal("failed to process npwp verification", err)
	}

	return svc.jobService.FinalizeJob(jobIdStr)
}

func (svc *service) BulkNPWPVerification(apiKey, quotaType string, memberId, companyId uint, file *multipart.FileHeader) error {
	records, err := helper.ParseCSVFile(file, constant.CSVTemplateHeaderNPWPVerification)
	if err != nil {
		return apperror.BadRequest(err.Error())
	}

	memberIdStr := strconv.Itoa(int(memberId))
	companyIdStr := strconv.Itoa(int(companyId))
	subscribedResp, err := svc.memberRepo.GetSubscribedProducts(companyIdStr, constant.SlugNPWPVerification)
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

	jobIdStr := strconv.Itoa(int(jobRes.JobId))
	var npwpVerificationReqs []*npwpVerificationRequest
	for i, rec := range records {
		if i == 0 {
			continue
		}
		npwpVerificationReqs = append(npwpVerificationReqs, &npwpVerificationRequest{
			Npwp:   rec[0],
			LoanNo: rec[1],
		})
	}

	var (
		wg         sync.WaitGroup
		errChan    = make(chan error, len(npwpVerificationReqs))
		batchCount = 0
	)

	for _, req := range npwpVerificationReqs {
		wg.Add(1)

		go func(npwpVerificationReq *npwpVerificationRequest) {
			defer wg.Done()

			if err := svc.processNPWPVerification(&npwpVerificationContext{
				APIKey:         apiKey,
				JobIdStr:       jobIdStr,
				MemberIdStr:    memberIdStr,
				CompanyIdStr:   companyIdStr,
				MemberId:       memberId,
				CompanyId:      companyId,
				ProductId:      subscribedResp.Data.ProductId,
				ProductGroupId: subscribedResp.Data.Product.ProductGroupId,
				JobId:          jobRes.JobId,
				Request:        npwpVerificationReq,
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
		logger.Error().Err(err).Msg("error during bulk npwp verification processing")
	}

	return svc.jobService.FinalizeJob(jobIdStr)
}

func (svc *service) processNPWPVerification(params *npwpVerificationContext) error {
	trxId := helper.GenerateTrx(constant.TrxIdNPWPVerification)
	if err := validator.ValidateStruct(params.Request); err != nil {
		_ = svc.logFailedTransaction(params, trxId, err.Error(), http.StatusBadRequest)

		return apperror.BadRequest(err.Error())
	}

	_, err := svc.repo.NPWPVerificationAPI(params.APIKey, params.JobIdStr, params.Request)
	if err != nil {
		_ = svc.logFailedTransaction(params, trxId, err.Error(), http.StatusBadGateway)
		_ = svc.jobService.FinalizeFailedJob(params.JobIdStr)

		return apperror.Internal("failed to process npwp verification", err)
	}

	return nil
}

func (svc *service) logFailedTransaction(params *npwpVerificationContext, trxId, msg string, status int) error {
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
