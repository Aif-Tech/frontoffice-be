package multipleloan

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
	MultipleLoan(authCtx *model.AuthContext, slug string, reqBody *multipleLoanRequest) (*model.ProCatAPIResponse[dataMultipleLoanResponse], error)
	BulkMultipleLoan(authCtx *model.AuthContext, slug string, file *multipart.FileHeader) error
}

type multipleLoanFunc func(
	apiKey string,
	jobID string,
	userID string,
	companyID string,
	req *multipleLoanRequest,
) (*model.ProCatAPIResponse[dataMultipleLoanResponse], error)

func (svc *service) MultipleLoan(authCtx *model.AuthContext, slug string, reqBody *multipleLoanRequest) (*model.ProCatAPIResponse[dataMultipleLoanResponse], error) {
	type multipleLoanHandler struct {
		handler            multipleLoanFunc
		event, productSlug string
	}

	var multipleLoanMap = map[string]multipleLoanHandler{
		"7d-multiple-loan": {
			handler:     svc.repo.CallMultipleLoan7Days,
			productSlug: constant.Slug7DaysMultipleLoan,
			event:       constant.Event7DMLSingleReq,
		},
		"30d-multiple-loan": {
			handler:     svc.repo.CallMultipleLoan30Days,
			productSlug: constant.Slug30DaysMultipleLoan,
			event:       constant.Event30DMLSingleReq,
		},
		"90d-multiple-loan": {
			handler:     svc.repo.CallMultipleLoan90Days,
			productSlug: constant.Slug90DaysMultipleLoan,
			event:       constant.Event90DMLSingleReq,
		},
	}

	mlCfg, ok := multipleLoanMap[slug]
	if !ok {
		return nil, apperror.BadRequest(constant.ErrUnsupportedProduct)
	}

	subscribedResp, err := svc.memberRepo.GetSubscribedProducts(authCtx.CompanyIdStr(), mlCfg.productSlug)
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

	result, err := mlCfg.handler(authCtx.APIKey, jobIdStr, authCtx.UserIdStr(), authCtx.CompanyIdStr(), reqBody)
	if err != nil {
		if err := svc.jobService.FinalizeFailedJob(jobIdStr); err != nil {
			return nil, err
		}

		var apiErr *apperror.ExternalAPIError
		if errors.As(err, &apiErr) {
			return nil, apperror.MapLoanError(apiErr)
		}

		return nil, apperror.Internal("failed to process multiple loan checker", err)
	}

	if err := svc.jobService.FinalizeJob(jobIdStr); err != nil {
		return nil, err
	}

	if err := svc.operationRepo.AddLogOperation(&operation.AddLogRequest{
		MemberId:  authCtx.UserId,
		CompanyId: authCtx.CompanyId,
		Action:    mlCfg.event,
	}); err != nil {
		log.Warn().
			Err(err).
			Str("action", mlCfg.event).
			Msg("failed to add operation log")
	}

	return result, nil
}

func (svc *service) BulkMultipleLoan(authCtx *model.AuthContext, slug string, file *multipart.FileHeader) error {
	records, err := helper.ParseCSVFile(file, constant.CSVTemplateHeaderMultipleLoan)
	if err != nil {
		return apperror.BadRequest(err.Error())
	}

	type multipleLoanHandler struct {
		handler            multipleLoanFunc
		TrxPrefix          string
		event, productSlug string
	}

	var multipleLoanMap = map[string]multipleLoanHandler{
		"7d-multiple-loan": {
			handler:     svc.repo.CallMultipleLoan7Days,
			TrxPrefix:   constant.TrxId7DaysMultipleLoan,
			productSlug: constant.Slug7DaysMultipleLoan,
			event:       constant.Event7DMLBulkReq,
		},
		"30d-multiple-loan": {
			handler:     svc.repo.CallMultipleLoan30Days,
			TrxPrefix:   constant.TrxId30DaysMultipleLoan,
			productSlug: constant.Slug30DaysMultipleLoan,
			event:       constant.Event30DMLBulkReq,
		},
		"90d-multiple-loan": {
			handler:     svc.repo.CallMultipleLoan90Days,
			TrxPrefix:   constant.TrxId90DaysMultipleLoan,
			productSlug: constant.Slug90DaysMultipleLoan,
			event:       constant.Event90DMLBulkReq,
		},
	}

	mlCfg, ok := multipleLoanMap[slug]
	if !ok {
		return apperror.BadRequest(constant.ErrUnsupportedProduct)
	}

	memberIdStr := authCtx.UserIdStr()
	companyIdStr := authCtx.CompanyIdStr()
	subscribedResp, err := svc.memberRepo.GetSubscribedProducts(companyIdStr, mlCfg.productSlug)
	if err != nil {
		return apperror.MapRepoError(err, constant.ErrFetchSubscribedProduct)
	}

	subscribedIdStr := strconv.Itoa(int(subscribedResp.Data.SubsribedProductID))
	quotaResp, err := svc.memberRepo.GetQuotaAPI(&member.QuotaParams{
		MemberId:     memberIdStr,
		CompanyId:    companyIdStr,
		SubscribedId: subscribedIdStr,
		QuotaType:    authCtx.QuotaTypeStr(),
	})
	if err != nil {
		return apperror.MapRepoError(err, constant.FailedFetchQuota)
	}

	totalRequests := len(records) - 1
	if authCtx.QuotaType != 0 && quotaResp.Data.Quota < totalRequests {
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
			Nik:    rec[0],
			Phone:  rec[1],
			LoanNo: rec[2],
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
				APIKey:         authCtx.APIKey,
				JobIdStr:       jobIdStr,
				MemberIdStr:    memberIdStr,
				CompanyIdStr:   companyIdStr,
				ProductSlug:    mlCfg.productSlug,
				MemberId:       authCtx.UserId,
				CompanyId:      authCtx.CompanyId,
				ProductId:      subscribedResp.Data.ProductId,
				ProductGroupId: subscribedResp.Data.Product.ProductGroupId,
				JobId:          jobRes.JobId,
				TrxPrefix:      mlCfg.TrxPrefix,
				Handler:        mlCfg.handler,
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
		log.Error().Err(err).Str("job_id", jobIdStr).Msg("error during bulk multiple loan processing")
	}

	if err := svc.jobService.FinalizeJob(jobIdStr); err != nil {
		return err
	}

	if err := svc.operationRepo.AddLogOperation(&operation.AddLogRequest{
		MemberId:  authCtx.UserId,
		CompanyId: authCtx.CompanyId,
		Action:    mlCfg.event,
	}); err != nil {
		log.Warn().
			Err(err).
			Str("action", mlCfg.event).
			Msg("failed to add operation log")
	}

	return nil
}

func (svc *service) processMultipleLoan(params *multipleLoanContext) error {
	trxId := helper.GenerateTrx(params.TrxPrefix)
	if err := validator.ValidateStruct(params.Request); err != nil {
		_ = svc.logFailedTransaction(params, trxId, err.Error(), http.StatusBadRequest)

		return apperror.BadRequest(err.Error())
	}

	result, err := params.Handler(params.APIKey, params.JobIdStr, params.MemberIdStr, params.CompanyIdStr, params.Request)
	if err != nil {
		statusCode := http.StatusBadGateway
		if result != nil {
			statusCode = result.StatusCode
		}

		_ = svc.logFailedTransaction(params, trxId, err.Error(), statusCode)
		_ = svc.jobService.FinalizeFailedJob(params.JobIdStr)

		return apperror.Internal("failed to process multiple loan", err)
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
