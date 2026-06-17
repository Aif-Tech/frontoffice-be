package negativerecord

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
	"strings"
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
	jobService job.Service) Service {
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
	NegativeRecord(authCtx *model.AuthContext, reqBody *negativeRecordRequest) (*model.ProCatAPIResponse[dataNegativeRecord], error)
	BulkNegativeRecord(authCtx *model.AuthContext, fileHeader *multipart.FileHeader) error
}

func (svc *service) NegativeRecord(authCtx *model.AuthContext, reqBody *negativeRecordRequest) (*model.ProCatAPIResponse[dataNegativeRecord], error) {
	subscribedResp, err := svc.memberRepo.GetSubscribedProducts(authCtx.CompanyIdStr(), constant.SlugNegativeRecord)
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

	// todo: remove
	dummyTrxId := helper.GenerateTrx(constant.TrxIdNegativeRecord)
	if err := svc.dummyLogTrans(&negativeRecordContext{
		APIKey:         authCtx.APIKey,
		JobIdStr:       jobIdStr,
		MemberIdStr:    authCtx.UserIdStr(),
		CompanyIdStr:   authCtx.CompanyIdStr(),
		MemberId:       authCtx.UserId,
		CompanyId:      authCtx.CompanyId,
		ProductId:      subscribedResp.Data.ProductId,
		ProductGroupId: subscribedResp.Data.Product.ProductGroupId,
		JobId:          jobRes.JobId,
		Request:        reqBody,
	}, dummyTrxId); err != nil {
		return nil, apperror.MapRepoError(err, constant.FailedCreateJob)
	}

	result, err := svc.repo.NegativeRecordAPI(authCtx.APIKey, dummyTrxId, reqBody)
	if err != nil {
		if err := svc.jobService.FinalizeFailedJob(jobIdStr); err != nil {
			return nil, err
		}

		return nil, apperror.MapRepoError(err, "failed to process recycle number")
	}

	if err := svc.jobService.FinalizeJob(jobIdStr); err != nil {
		return nil, err
	}

	if err := svc.operationRepo.AddLogOperation(&operation.AddLogRequest{
		MemberId:  authCtx.UserId,
		CompanyId: authCtx.CompanyId,
		Action:    constant.EventNegativeRecordSingleReq,
	}); err != nil {
		log.Warn().
			Err(err).
			Str("action", constant.EventNegativeRecordSingleReq).
			Msg(constant.MsgFailedAddOperationLog)
	}

	return result, nil
}

func (svc *service) BulkNegativeRecord(authCtx *model.AuthContext, file *multipart.FileHeader) error {
	records, err := helper.ParseCSVFile(file, constant.CSVTemplateHeaderNegativeRecord)
	if err != nil {
		return apperror.BadRequest(err.Error())
	}

	subscribedResp, err := svc.memberRepo.GetSubscribedProducts(authCtx.CompanyIdStr(), constant.SlugNegativeRecord)
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

	var requests []*negativeRecordRequest
	for i, rec := range records {
		if i == 0 {
			continue
		}

		requests = append(requests, &negativeRecordRequest{
			CompanyName: rec[0],
			LoanNo:      rec[1],
		})
	}

	var (
		wg         sync.WaitGroup
		errChan    = make(chan error, len(requests))
		batchCount = 0
	)

	for _, req := range requests {
		wg.Add(1)

		go func(negativeRecordReq *negativeRecordRequest) {
			defer wg.Done()

			if err := svc.processSingleNegativeRecord(&negativeRecordContext{
				APIKey:         authCtx.APIKey,
				JobIdStr:       jobIdStr,
				MemberIdStr:    authCtx.UserIdStr(),
				CompanyIdStr:   authCtx.CompanyIdStr(),
				MemberId:       authCtx.UserId,
				CompanyId:      authCtx.CompanyId,
				ProductId:      subscribedResp.Data.ProductId,
				ProductGroupId: subscribedResp.Data.Product.ProductGroupId,
				JobId:          jobRes.JobId,
				Request:        negativeRecordReq,
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
		log.Error().Err(err).Str("job_id", jobIdStr).Msg("error during bulk negative record processing")
	}

	if err := svc.jobService.FinalizeJob(jobIdStr); err != nil {
		return err
	}

	if err := svc.operationRepo.AddLogOperation(&operation.AddLogRequest{
		MemberId:  authCtx.UserId,
		CompanyId: authCtx.CompanyId,
		Action:    constant.EventNegativeRecordBulkReq,
	}); err != nil {
		log.Warn().
			Err(err).
			Str("action", constant.EventNegativeRecordBulkReq).
			Msg(constant.MsgFailedAddOperationLog)
	}

	return nil
}

func (svc *service) processSingleNegativeRecord(params *negativeRecordContext) error {
	trxId := helper.GenerateTrx(constant.TrxIdNegativeRecord)
	if err := validator.ValidateStruct(params.Request); err != nil {
		_ = svc.logFailedTransaction(params, trxId, err.Error(), http.StatusBadRequest)

		return apperror.BadRequest(err.Error())
	}

	if err := svc.dummyLogTrans(params, trxId); err != nil {
		return apperror.MapRepoError(err, constant.FailedCreateJob)
	}

	_, err := svc.repo.NegativeRecordAPI(params.APIKey, trxId, params.Request)
	if err != nil {
		_ = svc.logFailedTransaction(params, trxId, err.Error(), http.StatusBadGateway)
		_ = svc.jobService.FinalizeFailedJob(params.JobIdStr)

		return apperror.Internal("failed to process recycle number", err)
	}

	return nil
}

func (svc *service) logFailedTransaction(params *negativeRecordContext, trxId, msg string, status int) error {
	return svc.transactionRepo.CreateLogTransAPI(&transaction.LogTransProCatRequest{
		TransactionID:   trxId,
		MemberID:        params.MemberId,
		CompanyID:       params.CompanyId,
		ProductID:       params.ProductId,
		ProductGroupID:  params.ProductGroupId,
		JobID:           params.JobId,
		Message:         msg,
		Status:          status,
		PricingStrategy: "FREE",
		Success:         false,
		LoanNo:          params.Request.LoanNo,
		ResponseBody: &transaction.ResponseBody{
			Input:    params.Request,
			DateTime: time.Now().Format(constant.FormatDateAndTime),
		},
		RequestBody:  params.Request,
		RequestTime:  time.Now(),
		ResponseTime: time.Now(),
	})
}

// todo: remove
func (svc *service) dummyLogTrans(params *negativeRecordContext, dummyTrxId string) error {
	result := []dataNegativeRecordAPI{}
	companyName := strings.TrimSpace(params.Request.CompanyName)
	if strings.Contains(strings.ToLower(companyName), "artha") {
		result = []dataNegativeRecordAPI{
			{
				CompanyName:      "KOPERASI SIMPAN PINJAM ARTHA MULIA",
				Status:           "Penyerahan Memori Kasasi",
				Court:            "PENGADILAN NEGERI JAKARTA SELATAN",
				Province:         "DKI Jakarta",
				CaseNumber:       "411/Pdt.G/2025/PN JKT.SEL",
				CaseType:         "Wanprestasi",
				RegistrationDate: "2025-04-28",
				ProcessDuration:  "165 Hari",
				LastUpdated:      time.Now().Format("02 January 2006 15:04") + " WIB",
				SimilarityScore:  "100.0",
			},
			{
				CompanyName:      "ARTHA PRIMA FINANCE",
				Status:           "Minutasi",
				Court:            "PENGADILAN NEGERI BANDUNG",
				Province:         "Jawa Barat",
				CaseNumber:       "203/Pdt.G/2024/PN BDG",
				CaseType:         "Perbuatan Melawan Hukum",
				RegistrationDate: "2024-11-15",
				ProcessDuration:  "210 Hari",
				LastUpdated:      time.Now().Format("02 January 2006 15:04") + " WIB",
				SimilarityScore:  "90.5",
			},
		}
	}

	return svc.transactionRepo.CreateLogTransAPI(&transaction.LogTransProCatRequest{
		TransactionID:  dummyTrxId,
		MemberID:       params.MemberId,
		CompanyID:      params.CompanyId,
		ProductID:      params.ProductId,
		ProductGroupID: params.ProductGroupId,
		JobID:          params.JobId,
		Message:        constant.Success,
		Status:         http.StatusOK,
		Success:        true,
		LoanNo:         params.Request.LoanNo,
		ResponseBody: &transaction.ResponseBody{
			Data:            dataNegativeRecord{Result: result},
			Input:           params.Request,
			TransactionId:   dummyTrxId,
			PricingStrategy: "FREE",
			DateTime:        time.Now().Format(constant.FormatDateAndTime),
		},
		Data:         dataNegativeRecord{Result: result},
		RequestBody:  params.Request,
		RequestTime:  time.Now(),
		ResponseTime: time.Now(),
	})
}
