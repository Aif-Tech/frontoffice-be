package job

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"front-office/internal/core/log/operation"
	"front-office/internal/core/log/transaction"
	"front-office/pkg/apperror"
	"front-office/pkg/common/constant"
	"front-office/pkg/common/model"
	"front-office/pkg/helper"
	"strconv"
	"time"

	"github.com/rs/zerolog/log"
)

func NewService(repo Repository, transactionRepo transaction.Repository, operationRepo operation.Repository) Service {
	return &service{
		repo,
		transactionRepo,
		operationRepo,
	}
}

type service struct {
	repo            Repository
	transactionRepo transaction.Repository
	operationRepo   operation.Repository
}

type Service interface {
	CreateJob(req *CreateJobRequest) (*createJobRespData, error)
	UpdateJobAPI(jobId string, req *UpdateJobRequest) error
	GetJobs(filter *logFilter) (*model.AifcoreAPIResponse[*jobListResponse], error)
	GetGenRetailJobs(filter *logFilter) (*model.APIResponse[*jobGenRetailData], error)
	GetJobDetails(filter *logFilter) (*model.AifcoreAPIResponse[*jobDetailResponse], error)
	ExportJobDetails(filter *logFilter, buf *bytes.Buffer) (string, error)
	GetJobDetailsByDateRange(filter *logFilter) (*model.AifcoreAPIResponse[*jobDetailResponse], error)
	ExportJobDetailsByDateRange(filter *logFilter, buf *bytes.Buffer) (string, error)
	FinalizeJob(jobIdStr string) error
	FinalizeFailedJob(jobIdStr string) error
}

const (
	typePersonal = "personal"
	// typeCompany  = "company"
	hitTypeSingle = "single"
	hitTypeBulk   = "bulk"
)

func (svc *service) CreateJob(req *CreateJobRequest) (*createJobRespData, error) {
	result, err := svc.repo.CreateJobAPI(req)
	if err != nil {
		return nil, apperror.MapRepoError(err, constant.FailedCreateJob)
	}

	return result, nil
}

func (svc *service) UpdateJobAPI(jobId string, req *UpdateJobRequest) error {
	data := map[string]interface{}{}

	if req.SuccessCount != nil {
		data["success_count"] = *req.SuccessCount
	}

	if req.Status != nil {
		data["status"] = *req.Status
	}

	if req.EndAt != nil {
		data["end_at"] = *req.EndAt
	}

	err := svc.repo.UpdateJobAPI(jobId, data)
	if err != nil {
		return apperror.MapRepoError(err, "failed to update job")
	}

	return nil
}

func (svc *service) GetJobs(filter *logFilter) (*model.AifcoreAPIResponse[*jobListResponse], error) {
	// todo: remove
	dummyJobs := []job{
		{
			Id:           1,
			ProductId:    1,
			MemberId:     1,
			CompanyId:    1,
			Total:        2,
			SuccessCount: 2,
			Status:       "done",
			StartTime:    "2025-12-12 17:32:39",
			EndTime:      "2025-12-12 17:32:39",
		},
		{
			Id:           2,
			ProductId:    1,
			MemberId:     1,
			CompanyId:    1,
			Total:        5,
			SuccessCount: 5,
			Status:       "done",
			StartTime:    "2025-12-12 07:32:12",
			EndTime:      "2025-12-12 07:32:13",
		},
		{
			Id:           3,
			ProductId:    1,
			MemberId:     1,
			CompanyId:    1,
			Total:        4,
			SuccessCount: 4,
			Status:       "done",
			StartTime:    "2025-12-12 14:23:15",
			EndTime:      "2025-12-12 14:23:15",
		},
	}

	if filter.ProductSlug == constant.SlugRecycleNumber {
		return &model.AifcoreAPIResponse[*jobListResponse]{
			Success: true,
			Message: "success",
			Data: &jobListResponse{
				Jobs:      dummyJobs,
				TotalData: int64(len(dummyJobs)),
			},
		}, nil
	}

	result, err := svc.repo.GetJobsAPI(filter)
	if err != nil {
		return nil, apperror.MapRepoError(err, "failed to fetch jobs")
	}

	return result, nil
}

func (svc *service) GetGenRetailJobs(filter *logFilter) (*model.APIResponse[*jobGenRetailData], error) {
	result, err := svc.repo.GetJobsAPI(filter)
	if err != nil {
		return nil, apperror.MapRepoError(err, "failed to fetch jobs")
	}

	mappedJobs := []jobsScoreezy{}
	for _, item := range result.Data.Jobs {
		hitType := hitTypeSingle
		if item.Total > 1 {
			hitType = hitTypeBulk
		}

		mappedJobs = append(mappedJobs, jobsScoreezy{
			Id:          item.Id,
			MemberId:    item.MemberId,
			CompanyId:   item.CompanyId,
			ProductId:   item.ProductId,
			Total:       item.Total,
			HitType:     hitType,
			ProductType: typePersonal,
			CreatedAt:   item.StartTime,
		})
	}

	return helper.SuccessResponse(
		constant.Success,
		&jobGenRetailData{
			Logs:      mappedJobs,
			TotalData: result.Data.TotalData,
		}), nil
}

func (svc *service) GetJobDetails(filter *logFilter) (*model.AifcoreAPIResponse[*jobDetailResponse], error) {
	// todo: remove
	dummyJobDetails := []*logTransProductCatalog{
		{
			MemberID:  1,
			CompanyID: 1,
			JobID:     1,
			ProductID: 1,
			LoanNo:    "dummy",
			Status:    "success",
			Input: &logTransInput{
				PhoneNumber: helper.StringPtr("085755700000"),
				LoanNo:      "dummy",
			},
			Data: &logTransData{
				Status: helper.StringPtr("phone number has never been recycled"),
			},
			PricingStrategy: "FREE",
			TransactionId:   constant.DummyTransactionId,
			RefTransProductCatalog: map[string]any{
				"data": map[string]any{
					"status": "phone number has never been recycled",
				},
				"datetime": "2025-12-12 09:13:09",
				"input": map[string]any{
					"loan_no":      "dummy",
					"phone_number": "08575***000",
				},
				"transaction_id": constant.DummyTransactionId,
			},
		},
	}

	if filter.ProductSlug == constant.SlugRecycleNumber {
		return &model.AifcoreAPIResponse[*jobDetailResponse]{
			Success: true,
			Message: "success",
			Data: &jobDetailResponse{
				JobDetails:                 dummyJobDetails,
				TotalData:                  int64(len(dummyJobDetails)),
				TotalDataPercentageSuccess: 1,
				TotalDataPercentageFail:    0,
				TotalDataPercentageError:   0,
			},
			Meta: &model.Meta{
				Message:   "Success",
				Total:     1,
				Page:      1,
				Visible:   1,
				StartData: 1,
				EndData:   1,
				Size:      10,
			},
		}, nil
	}

	result, err := svc.repo.GetJobDetailAPI(filter)
	if err != nil {
		return nil, apperror.MapRepoError(err, "failed to fetch job detail")
	}

	return result, nil
}

func (svc *service) GetJobDetailsByDateRange(filter *logFilter) (*model.AifcoreAPIResponse[*jobDetailResponse], error) {
	result, err := svc.repo.GetJobsSummaryAPI(filter)
	if err != nil {
		return nil, apperror.MapRepoError(err, "failed to fetch job detail")
	}

	return result, nil
}

func (svc *service) ExportJobDetails(filter *logFilter, buf *bytes.Buffer) (string, error) {
	return svc.exportJobDetailsToCSV(filter, buf, svc.repo.GetJobDetailAPI, false)
}

func (svc *service) ExportJobDetailsByDateRange(filter *logFilter, buf *bytes.Buffer) (string, error) {
	return svc.exportJobDetailsToCSV(filter, buf, svc.repo.GetJobsSummaryAPI, true)
}

func (svc *service) exportJobDetailsToCSV(
	filter *logFilter,
	buf *bytes.Buffer,
	fetchFunc func(*logFilter) (*model.AifcoreAPIResponse[*jobDetailResponse], error),
	includeDate bool,
) (string, error) {
	resp, err := fetchFunc(filter)
	if err != nil {
		return "", apperror.MapRepoError(err, "failed to fetch job details")
	}

	cfg, ok := exportProductMap[filter.ProductSlug]
	if !ok {
		return "", apperror.BadRequest(constant.ErrUnsupportedProduct)
	}

	headers := cfg.headers
	eventName := cfg.event
	mapper := func(d *logTransProductCatalog) []string {
		return cfg.mapper(filter.IsMasked, d)
	}

	if includeDate {
		headers = append([]string{"Date"}, headers...)
		mapper = withDateColumn(mapper)
	}

	err = writeToCSV(buf, headers, resp.Data.JobDetails, mapper)
	if err != nil {
		return "", apperror.Internal("failed to write CSV", err)
	}

	filename := formatCSVFileName("job_detail", filter.StartDate, filter.EndDate, filter.JobId)

	if err := svc.operationRepo.AddLogOperation(&operation.AddLogRequest{
		MemberId:  filter.AuthCtx.UserId,
		CompanyId: filter.AuthCtx.CompanyId,
		Action:    eventName,
	}); err != nil {
		log.Warn().
			Err(err).
			Str("action", eventName).
			Msg("failed to add operation log")
	}

	return filename, nil
}

func (svc *service) FinalizeJob(jobIdStr string) error {
	count, err := svc.transactionRepo.ProcessedLogCountAPI(jobIdStr)
	if err != nil {
		return apperror.MapRepoError(err, "failed to get success count")
	}

	if err := svc.repo.UpdateJobAPI(jobIdStr, map[string]interface{}{
		"success_count": helper.IntPtr(int(count.ProcessedCount)),
		"status":        helper.StringPtr(constant.JobStatusDone),
		"end_at":        helper.TimePtr(time.Now()),
	}); err != nil {
		return apperror.MapRepoError(err, "failed to update job status")
	}

	return nil
}

func (svc *service) FinalizeFailedJob(jobIdStr string) error {
	count, err := svc.transactionRepo.ProcessedLogCountAPI(jobIdStr)
	if err != nil {
		return apperror.MapRepoError(err, "failed to get processed count request")
	}

	if err := svc.repo.UpdateJobAPI(jobIdStr, map[string]interface{}{
		"success_count": helper.IntPtr(int(count.ProcessedCount)),
		"status":        helper.StringPtr(constant.JobStatusFailed),
		"end_at":        helper.TimePtr(time.Now()),
	}); err != nil {
		return apperror.MapRepoError(err, "failed to update job status")
	}

	return nil
}

func writeToCSV[T any](buf *bytes.Buffer, headers []string, data []T, mapRow func(T) []string) error {
	writer := csv.NewWriter(buf)

	if err := writer.Write(headers); err != nil {
		return err
	}

	for _, item := range data {
		row := mapRow(item)
		if err := writer.Write(row); err != nil {
			return err
		}
	}

	writer.Flush()
	return writer.Error()
}

func formatCSVFileName(base, startDate, endDate, jobId string) string {
	if startDate == "" {
		return fmt.Sprintf("%s_id_%s.csv", base, jobId)
	}

	if endDate != "" && endDate != startDate {
		return fmt.Sprintf("%s_%s_until_%s.csv", base, startDate, endDate)
	}

	return fmt.Sprintf("%s_%s.csv", base, startDate)
}

type rowMapper func(*logTransProductCatalog) []string

func withDateColumn(mapper rowMapper) rowMapper {
	return func(d *logTransProductCatalog) []string {
		row := mapper(d)
		date := d.DateTime

		return append([]string{date}, row...)
	}
}

type exportProductConfig struct {
	headers []string
	event   string
	mapper  func(isMasked bool, d *logTransProductCatalog) []string
}

var exportProductMap = map[string]exportProductConfig{
	// Loan Record
	constant.SlugLoanRecordChecker: {
		headers: constant.CSVExportHeaderLoanRecord,
		event:   constant.EventLoanRecordDownload,
		mapper: func(isMasked bool, d *logTransProductCatalog) []string {
			return mapLoanRecordCheckerRow(isMasked, d)
		},
	},

	// Multiple Loan 7D
	constant.Slug7DaysMultipleLoan: {
		headers: constant.CSVExportHeaderMultipleLoan,
		event:   constant.Event7DMLDownload,
		mapper: func(isMasked bool, d *logTransProductCatalog) []string {
			return mapMultipleLoanRow(isMasked, d)
		},
	},

	// Multiple Loan 30D
	constant.Slug30DaysMultipleLoan: {
		headers: constant.CSVExportHeaderMultipleLoan,
		event:   constant.Event30DMLDownload,
		mapper: func(isMasked bool, d *logTransProductCatalog) []string {
			return mapMultipleLoanRow(isMasked, d)
		},
	},

	// Multiple Loan 90D
	constant.Slug90DaysMultipleLoan: {
		headers: constant.CSVExportHeaderMultipleLoan,
		event:   constant.Event90DMLDownload,
		mapper: func(isMasked bool, d *logTransProductCatalog) []string {
			return mapMultipleLoanRow(isMasked, d)
		},
	},

	// Tax Compliance Status
	constant.SlugTaxComplianceStatus: {
		headers: constant.CSVExportHeaderTaxCompliance,
		event:   constant.EventPTaxComplianceDownload,
		mapper: func(isMasked bool, d *logTransProductCatalog) []string {
			return mapTaxComplianceRow(isMasked, d)
		},
	},

	// Tax Score
	constant.SlugTaxScore: {
		headers: constant.CSVExportHeaderTaxScore,
		event:   constant.EventTaxScoreDownload,
		mapper: func(isMasked bool, d *logTransProductCatalog) []string {
			return mapTaxScoreRow(isMasked, d)
		},
	},

	// Tax Verification Detail
	constant.SlugTaxVerificationDetail: {
		headers: constant.CSVExportHeaderTaxVerification,
		event:   constant.EventTaxVerificationDownload,
		mapper: func(isMasked bool, d *logTransProductCatalog) []string {
			return mapTaxVerificationRow(isMasked, d)
		},
	},

	// NPWP Verification
	constant.SlugNPWPVerification: {
		headers: constant.CSVExportHeaderNPWPVerification,
		event:   constant.EventNPWPVerificationDownload,
		mapper: func(isMasked bool, d *logTransProductCatalog) []string {
			return mapNPWPVerificationRow(isMasked, d)
		},
	},
}

func mapLoanRecordCheckerRow(isMasked bool, d *logTransProductCatalog) []string {
	var (
		description string
		remarks     string
		status      string
		phoneNumber string
		nik         string
	)

	if d.Message != nil {
		description = *d.Message
	}

	if d.Data != nil {
		remarks = *d.Data.Remarks
		status = *d.Data.Status
	}

	var ref refTransProductCatalog
	if raw, err := json.Marshal(d.RefTransProductCatalog); err == nil {
		_ = json.Unmarshal(raw, &ref)
	}

	if isMasked {
		nik = ref.Input.NIK
		phoneNumber = ref.Input.PhoneNumber
	} else {
		phoneNumber = *d.Input.PhoneNumber
		nik = *d.Input.NIK
	}

	return []string{
		d.Input.LoanNo,
		*d.Input.Name,
		nik,
		phoneNumber,
		remarks,
		status,
		d.Status,
		description,
	}
}

func mapMultipleLoanRow(isMasked bool, d *logTransProductCatalog) []string {
	var (
		description string
		queryCount  int
		phoneNumber string
		nik         string
	)

	if d.Message != nil {
		description = *d.Message
	}

	if d.Data != nil {
		queryCount = *d.Data.QueryCount
	}

	var ref refTransProductCatalog
	if raw, err := json.Marshal(d.RefTransProductCatalog); err == nil {
		_ = json.Unmarshal(raw, &ref)
	}

	if isMasked {
		nik = ref.Input.NIK
		phoneNumber = ref.Input.PhoneNumber
	} else {
		phoneNumber = *d.Input.PhoneNumber
		nik = *d.Input.NIK
	}

	return []string{
		d.Input.LoanNo,
		nik,
		phoneNumber,
		strconv.Itoa(queryCount),
		d.Status,
		description,
	}
}

func mapTaxComplianceRow(isMasked bool, d *logTransProductCatalog) []string {
	var (
		description, nama, address, status, npwp string
	)

	if d.Message != nil {
		description = *d.Message
	}

	if d.Data != nil {
		nama = *d.Data.Nama
		address = *d.Data.Alamat
		status = *d.Data.Status
	}

	var ref refTransProductCatalog
	if raw, err := json.Marshal(d.RefTransProductCatalog); err == nil {
		_ = json.Unmarshal(raw, &ref)
	}

	if isMasked {
		npwp = ref.Input.NPWP
	} else {
		npwp = *d.Input.NPWP
	}

	return []string{
		// d.Input.LoanNo,
		npwp,
		nama,
		address,
		status,
		d.Status,
		description,
	}
}

func mapTaxScoreRow(isMasked bool, d *logTransProductCatalog) []string {
	var (
		description, name, address, status, score, npwp string
	)

	if d.Message != nil {
		description = *d.Message
	}

	if d.Data != nil {
		name = *d.Data.Nama
		address = *d.Data.Alamat
		status = *d.Data.Status
		score = *d.Data.Score
	}

	var ref refTransProductCatalog
	if raw, err := json.Marshal(d.RefTransProductCatalog); err == nil {
		_ = json.Unmarshal(raw, &ref)
	}

	if isMasked {
		npwp = ref.Input.NPWP
	} else {
		npwp = *d.Input.NPWP
	}

	return []string{
		d.Input.LoanNo,
		npwp,
		name,
		address,
		status,
		score,
		d.Status,
		description,
	}
}

func mapTaxVerificationRow(isMasked bool, d *logTransProductCatalog) []string {
	var (
		description, name, address, status, npwpVerification, taxCompliance, npwpOrNIK string
	)

	if d.Message != nil {
		description = *d.Message
	}

	if d.Data != nil {
		name = *d.Data.Nama
		address = *d.Data.Alamat
		npwpVerification = *d.Data.NPWPVerification
		// npwp = *d.Data.NPWP
		status = *d.Data.Status
		taxCompliance = *d.Data.TaxCompliance
	}

	var ref refTransProductCatalog
	if raw, err := json.Marshal(d.RefTransProductCatalog); err == nil {
		_ = json.Unmarshal(raw, &ref)
	}

	if isMasked {
		// if ref.Data.NPWP != "" {
		// 	npwp = ref.Data.NPWP
		// }

		npwpOrNIK = ref.Input.NPWPOrNIK
	} else {
		npwpOrNIK = *d.Input.NPWPOrNIK
	}

	return []string{
		d.Input.LoanNo,
		name,
		address,
		npwpOrNIK,
		// npwp,
		npwpVerification,
		status,
		taxCompliance,
		d.Status,
		description,
	}
}

func mapNPWPVerificationRow(isMasked bool, d *logTransProductCatalog) []string {
	var (
		description, name, address, status, npwp string
	)

	if d.Message != nil {
		description = *d.Message
	}

	if d.Data != nil {
		name = *d.Data.Nama
		address = *d.Data.Alamat
		status = *d.Data.Status
	}

	var ref refTransProductCatalog
	if raw, err := json.Marshal(d.RefTransProductCatalog); err == nil {
		_ = json.Unmarshal(raw, &ref)
	}

	if isMasked {
		npwp = ref.Input.NPWP
	} else {
		npwp = *d.Input.NPWP
	}

	return []string{
		d.Input.LoanNo,
		npwp,
		name,
		address,
		status,
		d.Status,
		description,
	}
}
