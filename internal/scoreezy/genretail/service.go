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
	"log"
	"mime/multipart"
	"strconv"
	"strings"
	"sync"
	"time"

	logger "github.com/rs/zerolog/log"
	"github.com/usepzaka/validator"
)

func NewService(
	repo Repository,
	gradeRepo grade.Repository,
	transRepo transaction.Repository,
	productRepo product.Repository,
	logRepo operation.Repository,
	jobRepo job.Repository,
	memberRepo member.Repository,
) Service {
	return &service{
		repo,
		gradeRepo,
		transRepo,
		productRepo,
		logRepo,
		jobRepo,
		memberRepo,
	}
}

type service struct {
	repo        Repository
	gradeRepo   grade.Repository
	transRepo   transaction.Repository
	productRepo product.Repository
	logRepo     operation.Repository
	jobRepo     job.Repository
	memberRepo  member.Repository
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
	if subscribedResp.Data.ProductId == 0 {
		return nil, apperror.NotFound(constant.ErrSubscribtionNotFound)
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

	addLogRequest := &operation.AddLogRequest{
		MemberId:  memberId,
		CompanyId: companyId,
		Action:    constant.EventCalculateScore,
	}

	err = svc.logRepo.AddLogOperation(addLogRequest)
	if err != nil {
		log.Println("Failed to log operation for calculate score")
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
	if subscribedResp.Data.ProductId == 0 {
		return 0, apperror.NotFound(constant.ErrSubscribtionNotFound)
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

		time.Sleep(20 * time.Millisecond) // add delay between processSingleGenRetail calls to avoid identical trx_id generation

		batchCount++
		if batchCount == 100 {
			time.Sleep(time.Second)
			batchCount = 0
		}
	}

	wg.Wait()
	close(errChan)

	for err := range errChan {
		logger.Error().Err(err).Msg("error during bulk gen retail processing")
	}

	return jobRes.JobId, nil
}

func (svc *service) GetLogsScoreezy(filter *filterLogs) (*model.AifcoreAPIResponse[[]*logTransScoreezy], error) {
	var result *model.AifcoreAPIResponse[[]*logTransScoreezy]
	var err error

	if filter.JobId == "" {
		return nil, apperror.BadRequest("job id is required")
	}

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
		if log.Data != nil {
			log.Data.Type = deriveTypeFromTrxId(log.Data.TrxId)
		}
	}

	return result, nil
}

func (svc *service) GetLogScoreezy(filter *filterLogs) (*logTransScoreezy, error) {
	result, err := svc.repo.GetLogByTrxIdAPI(filter)
	if err != nil {
		return nil, apperror.MapRepoError(err, "failed to fetch log scoreezy")
	}

	if result.LogTrxId == 0 {
		return nil, apperror.NotFound(constant.DataNotFound)
	}

	return result, nil
}

func (svc *service) ExportJobDetails(filter *filterLogs, buf *bytes.Buffer) (string, error) {
	if _, err := time.Parse(constant.FormatYYYYMMDD, filter.StartDate); err != nil {
		return "", apperror.BadRequest("invalid start_date format, use YYYY-MM-DD")
	}
	if _, err := time.Parse(constant.FormatYYYYMMDD, filter.EndDate); err != nil {
		return "", apperror.BadRequest("invalid end_date format, use YYYY-MM-DD")
	}

	result, err := svc.repo.GetLogsByRangeDateAPI(filter)
	if err != nil {
		return "", apperror.MapRepoError(err, "failed to fetch logs scoreezy")
	}

	var mappedDetails []*logTransScoreezy
	mappedDetails = append(mappedDetails, result.Data...)

	if err := writeToCSV(buf, mappedDetails); err != nil {
		return "", apperror.Internal("failed to write CSV", err)
	}

	filename := formatCSVFileName("job_summary", filter.StartDate, filter.EndDate)

	return filename, nil
}

func writeToCSV(buf *bytes.Buffer, logs []*logTransScoreezy) error {
	w := csv.NewWriter(buf)
	headers := []string{"Date Created", "Name", "Loan ID", "ID Card Number", "Phone Number", "Probability To Default", "Grade", "Description"}

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
			message              string
		)

		if log.CreatedAt.IsZero() {
			createdAt = ""
		} else {
			createdAt = log.CreatedAt.Format(constant.FormatDateAndTime)
		}

		if log.Data != nil && log.Data.Data != nil {
			name = log.Data.Data.Name
			loanID = log.Data.Data.LoanNo
			idCardNo = log.Data.Data.IdCardNo
			phoneNumber = log.Data.Data.PhoneNumber
			probabilityToDefault = log.Data.ProbabilityToDefault
			grade = log.Data.Grade
			message = log.Data.Message
		}

		row := []string{
			createdAt,
			name,
			loanID,
			idCardNo,
			phoneNumber,
			probabilityToDefault,
			grade,
			message,
		}

		if err := w.Write(row); err != nil {
			return err
		}
	}

	w.Flush()
	return w.Error()
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

// func (svc *service) finalizeJob(jobIdStr string) error {
// 	count, err := svc.transRepo.ProcessedLogCountAPI(jobIdStr)
// 	if err != nil {
// 		return apperror.MapRepoError(err, "failed to get success count")
// 	}

// 	if err := svc.jobRepo.UpdateJobAPI(jobIdStr, map[string]interface{}{
// 		"success_count": helper.IntPtr(int(count.ProcessedCount)),
// 		"status":        helper.StringPtr(constant.JobStatusDone),
// 		"end_at":        helper.TimePtr(time.Now()),
// 	}); err != nil {
// 		return apperror.MapRepoError(err, "failed to update job status")
// 	}

// 	return nil
// }

// func (svc *service) finalizeFailedJob(jobIdStr string) error {
// 	count, err := svc.transRepo.ProcessedLogCountAPI(jobIdStr)
// 	if err != nil {
// 		return apperror.MapRepoError(err, "failed to get processed count request")
// 	}

// 	if err := svc.jobRepo.UpdateJobAPI(jobIdStr, map[string]interface{}{
// 		"success_count": helper.IntPtr(int(count.ProcessedCount)),
// 		"status":        helper.StringPtr(constant.JobStatusFailed),
// 		"end_at":        helper.TimePtr(time.Now()),
// 	}); err != nil {
// 		return apperror.MapRepoError(err, "failed to update job status")
// 	}

// 	return nil
// }

// func (svc *service) BulkSearchUploadSvc(req []BulkSearchRequest, tempType, apiKey, userId, companyId string) error {
// 	var bulkSearches []*BulkSearch
// 	uploadId := uuid.NewString()

// 	for _, v := range req {
// 		// check api aif-core to get grade data

// 		genRetailRequest := &genRetailRequest{
// 			LoanNo:   v.LoanNo,
// 			Name:     v.Name,
// 			IdCardNo: v.NIK,
// 			PhoneNo:  v.PhoneNumber,
// 		}

// 		genRetailResponse, errRequest := svc.GenRetailV3(genRetailRequest, apiKey)
// 		if errRequest != nil {
// 			return errRequest
// 		}

// 		bulkSearch := &BulkSearch{
// 			UploadId:             uploadId,
// 			TransactionId:        genRetailResponse.Data.TransactionId, // from API
// 			Name:                 v.Name,
// 			IdCardNo:             v.NIK,
// 			PhoneNo:              v.PhoneNumber,
// 			LoanNo:               v.LoanNo,
// 			ProbabilityToDefault: genRetailResponse.Data.ProbabilityToDefault, // from API
// 			Grade:                genRetailResponse.Data.Grade,                // from API
// 			Date:                 genRetailResponse.Data.Date,                 // from API
// 			Type:                 tempType,
// 			UserId:               userId,
// 			CompanyId:            companyId,
// 		}

// 		bulkSearches = append(bulkSearches, bulkSearch)
// 	}

// 	err := svc.Repo.StoreImportData(bulkSearches, userId)
// 	if err != nil {
// 		return err
// 	}

// 	return nil
// }

// func (svc *service) GetBulkSearchSvc(tierLevel uint, userId, companyId string) ([]BulkSearchResponse, error) {

// 	bulkSearches, err := svc.Repo.GetAllBulkSearch(tierLevel, userId, companyId)
// 	if err != nil {
// 		return nil, err
// 	}

// 	var responseBulkSearches []BulkSearchResponse
// 	for _, v := range bulkSearches {
// 		bulkSearch := BulkSearchResponse{
// 			TransactionId:        v.TransactionId,
// 			Name:                 v.Name,
// 			IdCardNo:             v.IdCardNo,
// 			PhoneNo:              v.PhoneNo,
// 			LoanNo:               v.LoanNo,
// 			ProbabilityToDefault: v.ProbabilityToDefault,
// 			Grade:                v.Grade,
// 			Type:                 v.Type,
// 			Date:                 v.Date,
// 		}

// 		if tierLevel != 2 {
// 			// make sure only pick from the member uploads
// 			if userId != v.UserId {
// 				bulkSearch.PIC = v.User.Name
// 			}
// 		}

// 		responseBulkSearches = append(responseBulkSearches, bulkSearch)
// 	}

// 	return responseBulkSearches, nil
// }

// func (svc *service) GetTotalDataBulk(tierLevel uint, userId, companyId string) (int64, error) {
// 	count, err := svc.Repo.CountData(tierLevel, userId, companyId)
// 	return count, err
// }
