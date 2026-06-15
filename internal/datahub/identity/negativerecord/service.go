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
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
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

// todo: remove
func (svc *service) dummyLogTrans(params *negativeRecordContext, dummyTrxId string) error {
	result := []dataNegativeRecordAPI{}
	if params.Request.CompanyName == "artha" {
		result = []dataNegativeRecordAPI{
			{
				CompanyName:         "KOPERASI SIMPAN PINJAM ARTHA MULIA",
				CaseStatus:          "Penyerahan Memori Kasasi",
				Court:               "PENGADILAN NEGERI JAKARTA SELATAN",
				CaseNumber:          "411/Pdt.G/2025/PN JKT.SEL",
				CaseCodeDescription: "Perkara Perdata Gugatan",
				PartyStatus:         "Turut Tergugat",
				CaseClassification:  "Wanprestasi",
				RegistrationDate:    "2025-04-28",
				CaseDuration:        "165 Hari",
				SimilarityScore:     "100.0",
			},
			{
				CompanyName:         "ARTHA PRIMA FINANCE",
				CaseStatus:          "Minutasi",
				Court:               "PENGADILAN NEGERI BANDUNG",
				CaseNumber:          "203/Pdt.G/2024/PN BDG",
				CaseCodeDescription: "Perkara Perdata Gugatan",
				PartyStatus:         "Tergugat",
				CaseClassification:  "Perbuatan Melawan Hukum",
				RegistrationDate:    "2024-11-15",
				CaseDuration:        "210 Hari",
				SimilarityScore:     "87.5",
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
		RequestBody:  params.Request,
		RequestTime:  time.Now(),
		ResponseTime: time.Now(),
	})
}
