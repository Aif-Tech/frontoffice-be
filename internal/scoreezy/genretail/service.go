package genretail

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"front-office/internal/core/grade"
	"front-office/internal/core/log/operation"
	"front-office/internal/core/log/transaction"
	"front-office/internal/core/member"
	"front-office/internal/core/product"
	"front-office/internal/datahub/job"
	"front-office/pkg/apperror"
	"front-office/pkg/common/constant"
	"front-office/pkg/common/model"
	"front-office/pkg/helper"
	"mime/multipart"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/usepzaka/validator"
)

func NewService(
	repo Repository,
	gradeRepo grade.Repository,
	transRepo transaction.Repository,
	productRepo product.Repository,
	operationRepo operation.Repository,
	jobRepo job.Repository,
	memberRepo member.Repository,
) Service {
	return &service{
		repo,
		gradeRepo,
		transRepo,
		productRepo,
		operationRepo,
		jobRepo,
		memberRepo,
	}
}

type service struct {
	repo          Repository
	gradeRepo     grade.Repository
	transRepo     transaction.Repository
	productRepo   product.Repository
	operationRepo operation.Repository
	jobRepo       job.Repository
	memberRepo    member.Repository
}

const (
	typePersonal = "personal"
	// typeCompany  = "company"
)

type Service interface {
	GenRetailV3(memberId, companyId uint, payload *genRetailRequest) (*model.ScoreezyAPIResponse[dataGenRetailV3], error)
	BulkGenRetailV3(memberId, companyId uint, quotaType string, file *multipart.FileHeader) (uint, error)
	GetLogsScoreezy(filter *filterLogs) (*model.AifcoreAPIResponse[[]*logTransScoreezy], error)
	GetLogScoreezy(filter *filterLogs) (*logTransScoreezy, error)
	ExportJobDetails(filter *filterLogs, buf *bytes.Buffer) (string, error)
	// BulkSearchUploadSvc(req []BulkSearchRequest, tempType, apiKey, userId, companyId string) error
	// GetBulkSearchSvc(tierLevel uint, userId, companyId string) ([]BulkSearchResponse, error)
	// GetTotalDataBulk(tierLevel uint, userId, companyId string) (int64, error)
}

func (svc *service) GenRetailV3(memberId, companyId uint, payload *genRetailRequest) (*model.ScoreezyAPIResponse[dataGenRetailV3], error) {
	memberIdStr := helper.ConvertUintToString(memberId)
	companyIdStr := helper.ConvertUintToString(companyId)
	subscribedResp, err := svc.memberRepo.GetSubscribedProducts(companyIdStr, constant.SlugGenRetailV3)
	if err != nil {
		return nil, apperror.MapRepoError(err, constant.ErrFetchSubscribedProduct)
	}

	// make sure parameter settings are set
	gradeResp, err := svc.gradeRepo.GetGradesAPI(constant.SlugGenRetailV3, strconv.FormatUint(uint64(companyId), 10))
	if err != nil {
		return nil, apperror.MapRepoError(err, "failed to get grades")
	}

	if len(gradeResp.Grades) < 1 {
		return nil, apperror.BadRequest(constant.ParamSettingIsNotSet)
	}

	jobRes, err := svc.jobRepo.CreateJobAPI(&job.CreateJobRequest{
		ProductId: subscribedResp.Data.ProductId,
		MemberId:  memberIdStr,
		CompanyId: companyIdStr,
		Total:     1,
	})
	if err != nil {
		return nil, apperror.MapRepoError(err, constant.FailedCreateJob)
	}
	jobIdStr := helper.ConvertUintToString(jobRes.JobId)

	result, err := svc.repo.GenRetailV3API(strconv.FormatUint(uint64(memberId), 10), jobIdStr, payload)
	if err != nil {
		return nil, apperror.MapRepoError(err, "failed to process gen retail v3")
	}

	if err := svc.operationRepo.AddLogOperation(&operation.AddLogRequest{
		MemberId:  memberId,
		CompanyId: companyId,
		Action:    constant.EventScoreezySingleReq,
	}); err != nil {
		log.Warn().
			Err(err).
			Str("action", constant.EventScoreezySingleReq).
			Msg("failed to add operation log")
	}

	return result, err
}

func (svc *service) BulkGenRetailV3(memberId, companyId uint, quotaType string, file *multipart.FileHeader) (uint, error) {
	records, err := helper.ParseCSVFile(file, []string{"Name", "Loan Number", "ID Card Number", "Phone Number"})
	if err != nil {
		return 0, apperror.BadRequest(err.Error())
	}

	memberIdStr := helper.ConvertUintToString(memberId)
	companyIdStr := helper.ConvertUintToString(companyId)
	subscribedResp, err := svc.memberRepo.GetSubscribedProducts(companyIdStr, constant.SlugGenRetailV3)
	if err != nil {
		return 0, apperror.MapRepoError(err, constant.ErrFetchSubscribedProduct)
	}

	subscribedIdStr := strconv.Itoa(int(subscribedResp.Data.SubsribedProductID))
	quotaResp, err := svc.memberRepo.GetQuotaAPI(&member.QuotaParams{
		MemberId:     memberIdStr,
		CompanyId:    companyIdStr,
		SubscribedId: subscribedIdStr,
		QuotaType:    quotaType,
	})
	if err != nil {
		return 0, apperror.MapRepoError(err, constant.FailedFetchQuota)
	}

	totalRequests := len(records) - 1
	if quotaType != "0" && quotaResp.Data.Quota < totalRequests {
		return 0, apperror.Forbidden(constant.ErrQuotaExceeded)
	}

	// make sure parameter settings are set
	gradeResp, err := svc.gradeRepo.GetGradesAPI(constant.SlugGenRetailV3, strconv.FormatUint(uint64(companyId), 10))
	if err != nil {
		return 0, apperror.MapRepoError(err, "failed to get grades")
	}

	if len(gradeResp.Grades) < 1 {
		return 0, apperror.BadRequest(constant.ParamSettingIsNotSet)
	}

	var reqs []*genRetailRequest
	for i, rec := range records {
		if i == 0 { // skip header
			continue
		}

		reqs = append(reqs, &genRetailRequest{
			Name:     rec[0],
			LoanNo:   rec[1],
			IdCardNo: rec[2],
			PhoneNo:  rec[3],
		})
	}

	jobRes, err := svc.jobRepo.CreateJobAPI(&job.CreateJobRequest{
		ProductId: subscribedResp.Data.ProductId,
		MemberId:  memberIdStr,
		CompanyId: companyIdStr,
		Total:     totalRequests,
	})
	if err != nil {
		return 0, apperror.MapRepoError(err, constant.FailedCreateJob)
	}

	var (
		wg         sync.WaitGroup
		errChan    = make(chan error, len(reqs))
		batchCount = 0
	)

	for _, req := range reqs {
		wg.Add(1)

		go func(req *genRetailRequest) {
			defer wg.Done()

			if err := svc.processSingleGenRetail(&genRetailContext{
				MemberId:  memberId,
				CompanyId: companyId,
				ProductId: subscribedResp.Data.ProductId,
				JobId:     jobRes.JobId,
				Request:   req,
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
		log.Error().Err(err).Msg("error during bulk gen retail processing")
	}

	if err := svc.operationRepo.AddLogOperation(&operation.AddLogRequest{
		MemberId:  memberId,
		CompanyId: companyId,
		Action:    constant.EventScoreezyBulkReq,
	}); err != nil {
		log.Warn().
			Err(err).
			Str("action", constant.EventScoreezyBulkReq).
			Msg("failed to add operation log")
	}

	return jobRes.JobId, nil
}

func (svc *service) GetLogsScoreezy(filter *filterLogs) (*model.AifcoreAPIResponse[[]*logTransScoreezy], error) {
	var result *model.AifcoreAPIResponse[[]*logTransScoreezy]
	var err error

	validProductTypes := map[string]bool{
		"personal": true,
		"company":  true,
	}

	if filter.ProductType != "" {
		if !validProductTypes[filter.ProductType] {
			return nil, apperror.BadRequest(fmt.Sprintf("invalid product type: %s", filter.ProductType))
		}
	}

	if filter.StartDate != "" {
		if _, err := time.Parse(constant.FormatYYYYMMDD, filter.StartDate); err != nil {
			return nil, apperror.BadRequest("invalid start_date format, use YYYY-MM-DD")
		}
	}

	if filter.EndDate != "" {
		if _, err := time.Parse(constant.FormatYYYYMMDD, filter.EndDate); err != nil {
			return nil, apperror.BadRequest("invalid end_date format, use YYYY-MM-DD")
		}
	}

	result, err = svc.repo.GetLogsScoreezyAPI(filter)
	if err != nil {
		return nil, apperror.MapRepoError(err, "failed to fetch logs scoreezy")
	}

	for _, log := range result.Data {
		if log.Data == nil {
			continue
		}

		log.Data.Type = deriveTypeFromTrxId(log.Data.TrxId)

		if log.Data.Data != nil {
			if log.Data.RefTrans == nil {
				log.Data.RefTrans = &refTrans{}
			}

			log.Data.RefTrans.IdCardNo = helper.MaskingHead(log.Data.Data.IdCardNo, 10)
			log.Data.RefTrans.PhoneNumber = helper.MaskingMiddle(log.Data.Data.PhoneNumber)
		} else {
			log.Data.RefTrans = nil
		}
	}

	return result, nil
}

func (svc *service) GetLogScoreezy(filter *filterLogs) (*logTransScoreezy, error) {
	if filter.StartDate != "" {
		if _, err := time.Parse(constant.FormatYYYYMMDD, filter.StartDate); err != nil {
			return nil, apperror.BadRequest("invalid start_date format, use YYYY-MM-DD")
		}
	}

	if filter.EndDate != "" {
		if _, err := time.Parse(constant.FormatYYYYMMDD, filter.EndDate); err != nil {
			return nil, apperror.BadRequest("invalid end_date format, use YYYY-MM-DD")
		}
	}

	result, err := svc.repo.GetLogByTrxIdAPI(filter)
	if err != nil {
		return nil, apperror.MapRepoError(err, "failed to fetch log scoreezy")
	}

	if result.LogTrxId == 0 {
		return nil, apperror.NotFound(constant.DataNotFound)
	}

	if result.Data.Data != nil {
		if result.Data.Data != nil {
			if result.Data.RefTrans == nil {
				result.Data.RefTrans = &refTrans{}
			}

			result.Data.RefTrans.IdCardNo = helper.MaskingHead(result.Data.Data.IdCardNo, 10)
			result.Data.RefTrans.PhoneNumber = helper.MaskingMiddle(result.Data.Data.PhoneNumber)
		} else {
			result.Data.RefTrans = nil
		}
	}

	return result, nil
}

func (svc *service) ExportJobDetails(filter *filterLogs, buf *bytes.Buffer) (string, error) {
	includeDate := false
	if filter.StartDate != "" {
		includeDate = true
		if _, err := time.Parse(constant.FormatYYYYMMDD, filter.StartDate); err != nil {
			return "", apperror.BadRequest("invalid start_date format, use YYYY-MM-DD")
		}
	}

	if filter.EndDate != "" {
		if _, err := time.Parse(constant.FormatYYYYMMDD, filter.EndDate); err != nil {
			return "", apperror.BadRequest("invalid end_date format, use YYYY-MM-DD")
		}
	}

	result, err := svc.repo.GetLogsScoreezyAPI(filter)
	if err != nil {
		return "", apperror.MapRepoError(err, "failed to fetch logs scoreezy")
	}

	var mappedDetails []*logTransScoreezy
	mappedDetails = append(mappedDetails, result.Data...)

	if err := writeToCSV(buf, includeDate, filter.Masked, mappedDetails); err != nil {
		return "", apperror.Internal("failed to write CSV", err)
	}

	filename := formatCSVFileName("job_summary", filter.StartDate, filter.EndDate)

	companyIdUint, err := strconv.ParseUint(filter.CompanyId, 10, 32)
	if err != nil {
		return "", apperror.Internal("failed to parse company id", err)
	}

	var action string
	if len(result.Data) > 1 {
		action = constant.EventScoreezyBulkDownload
	} else {
		action = constant.EventScoreezySingleDownload
	}

	if err := svc.operationRepo.AddLogOperation(&operation.AddLogRequest{
		MemberId:  filter.MemberId,
		CompanyId: uint(companyIdUint),
		Action:    action,
	}); err != nil {
		log.Warn().
			Err(err).
			Str("action", action).
			Msg("failed to add operation log")
	}

	return filename, nil
}

func writeToCSV(buf *bytes.Buffer, includeDate, masked bool, logs []*logTransScoreezy) error {
	w := csv.NewWriter(buf)
	defer w.Flush()

	headers := buildHeaders(constant.CSVExportHeaderGenRetail, includeDate)
	if err := w.Write(headers); err != nil {
		return err
	}

	for _, log := range logs {
		var (
			createdAt            string
			name                 string
			loanID               string
			idCardNo             string
			phoneNumber          string
			probabilityToDefault string
			grade                string
			behavior             string
			identity             string
			message              string
		)

		if log.Data != nil && log.Data.Data != nil {
			name = log.Data.Data.Name
			loanID = log.Data.Data.LoanNo

			probabilityToDefault = log.Data.ProbabilityToDefault
			grade = log.Data.Grade
			behavior = log.Data.Behavior
			identity = log.Data.Identity
			message = log.Data.Message

			if !masked {
				idCardNo = log.Data.Data.IdCardNo
				phoneNumber = log.Data.Data.PhoneNumber
			} else {
				idCardNo = helper.MaskingHead(log.Data.Data.IdCardNo, 10)
				phoneNumber = helper.MaskingMiddle(log.Data.Data.PhoneNumber)
			}
		}

		row := []string{
			loanID,
			name,
			idCardNo,
			phoneNumber,
			probabilityToDefault,
			grade,
			behavior,
			identity,
			message,
		}

		if includeDate {
			if log.CreatedAt.IsZero() {
				createdAt = ""
			} else {
				createdAt = log.CreatedAt.Format(constant.FormatDateAndTime)
			}

			row = append([]string{createdAt}, row...)
		}

		if err := w.Write(row); err != nil {
			return err
		}
	}

	return w.Error()
}

func buildHeaders(base []string, includeDate bool) []string {
	if includeDate {
		return append([]string{"Date"}, base...)
	}
	return base
}

func formatCSVFileName(base, startDate, endDate string) string {
	if endDate != "" && endDate != startDate {
		return fmt.Sprintf("%s_%s_until_%s.csv", base, startDate, endDate)
	}

	return fmt.Sprintf("%s_%s.csv", base, startDate)
}

func (svc *service) processSingleGenRetail(params *genRetailContext) error {
	if err := validator.ValidateStruct(params.Request); err != nil {
		_ = svc.transRepo.CreateLogScoreezyAPI(&transaction.LogTransScoreezy{
			TrxId:     helper.GenerateTrx(constant.TrxIdGenRetailV3),
			MemberId:  params.MemberId,
			CompanyId: params.CompanyId,
			ProductId: params.ProductId,
			JobId:     params.JobId,
			Message:   err.Error(),
			Status:    "FREE",
			Success:   false,
		})

		return apperror.BadRequest(err.Error())
	}

	_, err := svc.repo.GenRetailV3API(
		strconv.FormatUint(uint64(params.MemberId), 10),
		strconv.FormatUint(uint64(params.JobId), 10),
		params.Request,
	)
	if err != nil {
		return apperror.MapRepoError(err, "failed to process gen retail v3")
	}

	return nil
}

func deriveTypeFromTrxId(trxId string) string {
	switch {
	case strings.Contains(trxId, constant.TrxIdGenRetailV3):
		return typePersonal
	default:
		return ""
	}
}
