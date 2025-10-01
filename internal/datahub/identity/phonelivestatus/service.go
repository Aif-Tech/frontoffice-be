package phonelivestatus

import (
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"front-office/internal/core/log/transaction"
	"front-office/internal/core/member"
	"front-office/internal/datahub/job"
	"front-office/pkg/apperror"
	"front-office/pkg/common/constant"
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
	PhoneLiveStatus(apiKey, memberId, companyId string, reqBody *phoneLiveStatusRequest) error
	BulkPhoneLiveStatus(apiKey, memberId, companyId, quotaType string, fileHeader *multipart.FileHeader) error
	GetJobs(filter *phoneLiveStatusFilter) (*jobListRespData, error)
	GetJobDetails(filter *phoneLiveStatusFilter) (*jobDetailsDTO, error)
	ExportJobDetails(filter *phoneLiveStatusFilter, buf *bytes.Buffer) (string, error)
	GetJobsSummary(filter *phoneLiveStatusFilter) (*jobsSummaryDTO, error)
	ExportJobsSummary(filter *phoneLiveStatusFilter, buf *bytes.Buffer) (string, error)
}

func (svc *service) PhoneLiveStatus(apiKey, memberId, companyId string, reqBody *phoneLiveStatusRequest) error {
	subscribedResp, err := svc.memberRepo.GetSubscribedProducts(companyId, constant.SlugPhoneLiveStatus)
	if err != nil {
		return apperror.MapRepoError(err, constant.ErrFetchSubscribedProduct)
	}
	if subscribedResp.Data.ProductId == 0 {
		return apperror.NotFound(constant.ErrSubscribtionNotFound)
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

	result, err := svc.repo.PhoneLiveStatusAPI(apiKey, jobIdStr, reqBody)
	if err != nil {
		if err := svc.jobService.FinalizeFailedJob(jobIdStr); err != nil {
			return err
		}

		var apiErr *apperror.ExternalAPIError
		if errors.As(err, &apiErr) {
			return apperror.MapLoanError(apiErr)
		}

		return apperror.Internal("failed to process phone live status", err)
	}

	if err := svc.transactionRepo.UpdateLogTransAPI(result.TransactionId, map[string]interface{}{
		"success": helper.BoolPtr(true),
	}); err != nil {
		return apperror.MapRepoError(err, "failed to update transaction log")
	}

	return svc.jobService.FinalizeJob(jobIdStr)
}

func (svc *service) BulkPhoneLiveStatus(apiKey, memberId, companyId, quotaType string, file *multipart.FileHeader) error {
	if err := helper.ValidateUploadedFile(file, 30*1024*1024, []string{".csv"}); err != nil {
		return apperror.BadRequest(err.Error())
	}

	records, err := helper.ParseCSVFile(file, []string{"Phone Number"})
	if err != nil {
		return apperror.Internal(constant.FailedParseCSV, err)
	}

	subscribedResp, err := svc.memberRepo.GetSubscribedProducts(companyId, constant.SlugPhoneLiveStatus)
	if err != nil {
		return apperror.MapRepoError(err, constant.ErrFetchSubscribedProduct)
	}
	if subscribedResp.Data.ProductId == 0 {
		return apperror.NotFound(constant.ProductNotFound)
	}

	subscribedIdStr := strconv.Itoa(int(subscribedResp.Data.SubsribedProductID))
	quotaResp, err := svc.memberRepo.GetQuotaAPI(&member.QuotaParams{
		MemberId:     memberId,
		CompanyId:    companyId,
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
		MemberId:  memberId,
		CompanyId: companyId,
		Total:     totalRequests,
	})
	if err != nil {
		return apperror.MapRepoError(err, constant.FailedCreateJob)
	}
	jobIdStr := helper.ConvertUintToString(jobRes.JobId)

	var phoneReqs []*phoneLiveStatusRequest
	for i := 1; i < len(records); i++ { // Skip header
		phoneReqs = append(phoneReqs, &phoneLiveStatusRequest{
			PhoneNumber: records[i][0],
		})
	}

	var (
		wg         sync.WaitGroup
		errChan    = make(chan error, len(phoneReqs))
		batchCount = 0
	)

	for _, req := range phoneReqs {
		wg.Add(1)

		go func(phoneLiveReq *phoneLiveStatusRequest) {
			defer wg.Done()
			if err := svc.processSingle(&phoneLiveStatusContext{
				APIKey:         apiKey,
				JobIdStr:       jobIdStr,
				MemberId:       jobRes.MemberId,
				CompanyId:      jobRes.CompanyId,
				ProductId:      subscribedResp.Data.ProductId,
				ProductGroupId: subscribedResp.Data.Product.ProductGroupId,
				JobId:          jobRes.JobId,
				Request:        phoneLiveReq,
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
		log.Error().Err(err).Msg("error during bulk phone live status processing")
	}

	return svc.jobService.FinalizeJob(jobIdStr)
}

func (svc *service) processSingle(params *phoneLiveStatusContext) error {
	if err := validator.ValidateStruct(params.Request); err != nil {
		_ = svc.transactionRepo.CreateLogTransAPI(&transaction.LogTransProCatRequest{
			MemberID:       params.MemberId,
			CompanyID:      params.CompanyId,
			ProductID:      params.ProductId,
			ProductGroupID: params.ProductGroupId,
			JobID:          params.JobId,
			Message:        err.Error(),
			Status:         http.StatusBadRequest,
			Success:        false,
			ResponseBody: &transaction.ResponseBody{
				Input:    params.Request,
				DateTime: time.Now().Format(constant.FormatDateAndTime),
			},
			Data:         nil,
			RequestBody:  params.Request,
			RequestTime:  time.Now(),
			ResponseTime: time.Now(),
		})

		return apperror.BadRequest(err.Error())
	}

	result, err := svc.repo.PhoneLiveStatusAPI(params.APIKey, params.JobIdStr, params.Request)
	if err != nil {
		if err := svc.transactionRepo.CreateLogTransAPI(&transaction.LogTransProCatRequest{
			MemberID:       params.MemberId,
			CompanyID:      params.CompanyId,
			ProductID:      params.ProductId,
			ProductGroupID: params.ProductGroupId,
			JobID:          params.JobId,
			Message:        result.Message,
			Status:         result.StatusCode,
			Success:        false,
			ResponseBody: &transaction.ResponseBody{
				Input:    params.Request,
				DateTime: time.Now().Format(constant.FormatDateAndTime),
			},
			Data:         nil,
			RequestBody:  params.Request,
			RequestTime:  time.Now(),
			ResponseTime: time.Now(),
		}); err != nil {
			return err
		}

		if err := svc.jobService.FinalizeFailedJob(params.JobIdStr); err != nil {
			return err
		}

		var apiErr *apperror.ExternalAPIError
		if errors.As(err, &apiErr) {
			return apperror.MapLoanError(apiErr)
		}

		return apperror.Internal("failed to process loan record checker", err)
	}

	if err := svc.transactionRepo.UpdateLogTransAPI(result.TransactionId, map[string]interface{}{
		"success": helper.BoolPtr(true),
	}); err != nil {
		return apperror.MapRepoError(err, "failed to update log transaction")
	}

	return nil
}

func (svc *service) GetJobs(filter *phoneLiveStatusFilter) (*jobListRespData, error) {
	jobs, err := svc.repo.GetPhoneLiveStatusJobAPI(filter)
	if err != nil {
		return nil, apperror.MapRepoError(err, "failed to fetch phone live status jobs")
	}

	return jobs, nil
}

func (svc *service) GetJobDetails(filter *phoneLiveStatusFilter) (*jobDetailsDTO, error) {
	data, err := svc.repo.GetJobDetailsAPI(filter)
	if err != nil {
		return nil, apperror.MapRepoError(err, constant.ErrFetchPhoneLiveDetail)
	}

	var (
		mappedDetails []*mstPhoneLiveStatusJobDetail
	)

	metrics, err := svc.repo.GetJobMetricsAPI(filter)
	if err != nil {
		return nil, apperror.MapRepoError(err, constant.ErrFetchJobMetrics)
	}

	for _, raw := range data.JobDetails {
		mapped, err := mapToJobDetail(filter.Masked, raw)
		if err != nil {
			continue
		}

		mappedDetails = append(mappedDetails, mapped)
	}

	result := &jobDetailsDTO{
		TotalData:                  data.TotalData,
		TotalDataPercentageSuccess: data.TotalDataPercentageSuccess,
		TotalDataPercentageFail:    data.TotalDataPercentageFail,
		TotalDataPercentageError:   data.TotalDataPercentageError,
		SubsActive:                 metrics.SubsActive,
		SubsDisconnected:           metrics.SubsDisconnected,
		DevReachable:               metrics.DevReachable,
		DevUnreachable:             metrics.DevUnreachable,
		DevUnavailable:             metrics.DevUnavailable,
		JobDetails:                 mappedDetails,
	}

	return result, nil
}

func (svc *service) ExportJobDetails(filter *phoneLiveStatusFilter, buf *bytes.Buffer) (string, error) {
	data, err := svc.repo.GetJobDetailsAPI(filter)
	if err != nil {
		return "", apperror.MapRepoError(err, constant.ErrFetchPhoneLiveDetail)
	}

	var (
		mappedDetails []*mstPhoneLiveStatusJobDetail
	)

	for _, raw := range data.JobDetails {
		mapped, err := mapToJobDetail(filter.Masked, raw)
		if err != nil {
			continue
		}

		mappedDetails = append(mappedDetails, mapped)
	}

	if err := writeJobDetailsToCSV(buf, mappedDetails); err != nil {
		return "", apperror.Internal("failed to write CSV", err)
	}

	filename := formatCSVFileName("job_summary", filter.StartDate, filter.EndDate)

	return filename, nil
}

func (svc *service) GetJobsSummary(filter *phoneLiveStatusFilter) (*jobsSummaryDTO, error) {
	data, err := svc.repo.GetJobsSummaryAPI(filter)
	if err != nil {
		return nil, apperror.MapRepoError(err, constant.ErrFetchPhoneLiveDetail)
	}

	metrics, err := svc.repo.GetJobMetricsAPI(filter)
	if err != nil {
		return nil, apperror.MapRepoError(err, constant.ErrFetchJobMetrics)
	}

	result := &jobsSummaryDTO{
		TotalData:                  data.TotalData,
		TotalDataPercentageSuccess: data.TotalDataPercentageSuccess,
		TotalDataPercentageFail:    data.TotalDataPercentageFail,
		TotalDataPercentageError:   data.TotalDataPercentageError,
		SubsActive:                 metrics.SubsActive,
		SubsDisconnected:           metrics.SubsDisconnected,
		DevReachable:               metrics.DevReachable,
		DevUnreachable:             metrics.DevUnreachable,
		DevUnavailable:             metrics.DevUnavailable,
		Mobile:                     metrics.Mobile,
		FixedLine:                  metrics.FixedLine,
	}

	return result, nil
}

func (svc *service) ExportJobsSummary(filter *phoneLiveStatusFilter, buf *bytes.Buffer) (string, error) {
	data, err := svc.repo.GetJobsSummaryAPI(filter)
	if err != nil {
		return "", apperror.MapRepoError(err, constant.ErrFetchPhoneLiveDetail)
	}

	var (
		mappedDetails []*mstPhoneLiveStatusJobDetail
	)

	for _, raw := range data.JobDetails {
		mapped, err := mapToJobDetail(filter.Masked, raw)
		if err != nil {
			continue
		}

		mappedDetails = append(mappedDetails, mapped)
	}

	if err := writeJobDetailsToCSV(buf, mappedDetails); err != nil {
		return "", apperror.Internal("failed to write CSV", err)
	}

	filename := formatCSVFileName("job_summary", filter.StartDate, filter.EndDate)

	return filename, nil
}

func mapToJobDetail(masked bool, raw *logTransProductCatalog) (*mstPhoneLiveStatusJobDetail, error) {
	var subscriberStatus, deviceStatus, phoneType, operator, phoneNumber string
	if raw.Data != nil {
		liveStatusParts := strings.Split(raw.Data.LiveStatus, ",")
		if len(liveStatusParts) >= 2 {
			subscriberStatus = strings.TrimSpace(liveStatusParts[0])
			deviceStatus = strings.TrimSpace(liveStatusParts[1])
		} else if len(liveStatusParts) == 1 {
			subscriberStatus = strings.TrimSpace(liveStatusParts[0])
		}

		operator = raw.Data.Operator
		phoneType = raw.Data.PhoneType
	}

	createdAt, err := time.Parse("2006-01-02 15:04:05", raw.DateTime)
	if err != nil {
		return nil, fmt.Errorf("invalid datetime format: %v", err)
	}

	if masked {
		phoneNumber = raw.RefTransProductCatalog.Input.PhoneNumber
	} else {
		phoneNumber = raw.Input.PhoneNumber
	}

	return &mstPhoneLiveStatusJobDetail{
		MemberId:         raw.MemberID,
		CompanyId:        raw.CompanyID,
		JobId:            raw.JobID,
		PhoneNumber:      phoneNumber,
		Status:           raw.Status,
		Message:          raw.Message,
		SubscriberStatus: subscriberStatus,
		DeviceStatus:     deviceStatus,
		PhoneType:        phoneType,
		Operator:         operator,
		PricingStrategy:  raw.PricingStrategy,
		TransactionId:    raw.TransactionId,
		CreatedAt:        createdAt,
		RefLogTrx: RefLogTrx{
			PhoneNumber: raw.RefTransProductCatalog.Input.PhoneNumber,
		},
	}, nil
}

func writeJobDetailsToCSV(buf *bytes.Buffer, data []*mstPhoneLiveStatusJobDetail) error {
	w := csv.NewWriter(buf)
	headers := []string{"Phone Number", "Subscriber Status", "Device Status", "Operator", "Phone Type", "Status", "Description"}

	if err := w.Write(headers); err != nil {
		return err
	}

	for _, d := range data {
		desc := ""
		if d.Message != nil {
			desc = *d.Message
		}

		row := []string{d.PhoneNumber, d.SubscriberStatus, d.DeviceStatus, d.Operator, d.PhoneType, d.Status, desc}
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
