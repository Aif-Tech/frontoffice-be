package job

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"front-office/internal/core/log/transaction"
	"front-office/pkg/apperror"
	"front-office/pkg/common/constant"
	"front-office/pkg/common/model"
	"front-office/pkg/helper"
	"strconv"
	"time"
)

func NewService(repo Repository, transactionRepo transaction.Repository) Service {
	return &service{
		repo,
		transactionRepo,
	}
}

type service struct {
	repo            Repository
	transactionRepo transaction.Repository
}

type Service interface {
	CreateJob(req *CreateJobRequest) (*createJobRespData, error)
	UpdateJobAPI(jobId string, req *UpdateJobRequest) error
	GetJobs(filter *logFilter) (*model.AifcoreAPIResponse[*jobListResponse], error)
	GetGenRetailJobs(filter *logFilter) (*jobsGenRetailResponse, error)
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
	result, err := svc.repo.GetJobsAPI(filter)
	if err != nil {
		return nil, apperror.MapRepoError(err, "failed to fetch jobs")
	}

	return result, nil
}

func (svc *service) GetGenRetailJobs(filter *logFilter) (*jobsGenRetailResponse, error) {
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

	response := &jobsGenRetailResponse{
		Success: result.Success,
		Message: result.Message,
		Data:    &jobsGenRetailData{Logs: mappedJobs, TotalData: result.Data.TotalData},
	}

	return response, nil
}

func (svc *service) GetJobDetails(filter *logFilter) (*model.AifcoreAPIResponse[*jobDetailResponse], error) {
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

	headers := []string{}
	var mapper func(*logTransProductCatalog) []string

	switch filter.ProductSlug {
	case constant.SlugLoanRecordChecker:
		headers = []string{"Name", "NIK", "Phone Number", "Remarks", "Data Status", "Status", "Description"}
		mapper = func(d *logTransProductCatalog) []string {
			return mapLoanRecordCheckerRow(filter.IsMasked, d)
		}
	case constant.SlugMultipleLoan7Days, constant.SlugMultipleLoan30Days, constant.SlugMultipleLoan90Days:
		headers = []string{"NIK", "Phone Number", "Query Count", "Status", "Description"}
		mapper = func(d *logTransProductCatalog) []string {
			return mapMultipleLoanRow(filter.IsMasked, d)
		}
	case constant.SlugTaxComplianceStatus:
		headers = []string{"NPWP", "Nama", "Alamat", "Data Status", "Status", "Description"}
		mapper = func(d *logTransProductCatalog) []string {
			return mapTaxComplianceRow(filter.IsMasked, d)
		}
	case constant.SlugTaxScore:
		headers = []string{"NPWP", "Nama", "Alamat", "Data Status", "Score", "Status", "Description"}
		mapper = func(d *logTransProductCatalog) []string {
			return mapTaxScoreRow(filter.IsMasked, d)
		}
	case constant.SlugTaxVerificationDetail:
		headers = []string{"Nama", "Alamat", "NPWP", "NPWP Verification", "Data Status", "Tax Compliance", "Status", "Description"}
		mapper = func(d *logTransProductCatalog) []string {
			return mapTaxVerificationRow(filter.IsMasked, d)
		}
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
		nik,
		phoneNumber,
		strconv.Itoa(queryCount),
		d.Status,
		description,
	}
}

func mapTaxComplianceRow(isMasked bool, d *logTransProductCatalog) []string {
	var (
		description, nama, alamat, status, npwp string
	)

	if d.Message != nil {
		description = *d.Message
	}

	if d.Data != nil {
		nama = *d.Data.Nama
		alamat = *d.Data.Alamat
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
		npwp,
		nama,
		alamat,
		status,
		d.Status,
		description,
	}
}

func mapTaxScoreRow(isMasked bool, d *logTransProductCatalog) []string {
	var (
		description, nama, alamat, status, score, npwp string
	)

	if d.Message != nil {
		description = *d.Message
	}

	if d.Data != nil {
		nama = *d.Data.Nama
		alamat = *d.Data.Alamat
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
		npwp,
		nama,
		alamat,
		status,
		score,
		d.Status,
		description,
	}
}

func mapTaxVerificationRow(isMasked bool, d *logTransProductCatalog) []string {
	var (
		description, nama, alamat, status, npwpVerification, taxCompliance, npwpOrNIK string
	)

	if d.Message != nil {
		description = *d.Message
	}

	if d.Data != nil {
		nama = *d.Data.Nama
		alamat = *d.Data.Alamat
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
		nama,
		alamat,
		npwpOrNIK,
		// npwp,
		npwpVerification,
		status,
		taxCompliance,
		d.Status,
		description,
	}
}
