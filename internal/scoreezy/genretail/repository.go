package genretail

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"front-office/configs/application"
	"front-office/pkg/common/constant"
	"front-office/pkg/common/model"
	"front-office/pkg/helper"
	"front-office/pkg/httpclient"
	"front-office/pkg/jsonutil"
	"net/http"
)

func NewRepository(cfg *application.Config, client httpclient.HTTPClient, marshalFn jsonutil.Marshaller) Repository {
	if marshalFn == nil {
		marshalFn = json.Marshal // default behavior
	}

	return &repository{
		cfg:       cfg,
		client:    client,
		marshalFn: marshalFn,
	}
}

type repository struct {
	cfg       *application.Config
	client    httpclient.HTTPClient
	marshalFn jsonutil.Marshaller
}

type Repository interface {
	GenRetailV3API(memberId, jobId string, payload *genRetailRequest) (*model.ScoreezyAPIResponse[dataGenRetailV3], error)
	GetLogsScoreezyAPI(filter *filterLogs) (*model.AifcoreAPIResponse[[]*logTransScoreezy], error)
	GetLogsByRangeDateAPI(filter *filterLogs) (*model.AifcoreAPIResponse[[]*logTransScoreezy], error)
	GetLogByTrxIdAPI(filter *filterLogs) (*logTransScoreezy, error)
	// StoreImportData(newData []*BulkSearch, userId string) error
	// GetAllBulkSearch(tierLevel uint, userId, companyId string) ([]*BulkSearch, error)
	// CountData(tierLevel uint, userId, companyId string) (int64, error)
}

func (repo *repository) GenRetailV3API(memberId, jobId string, payload *genRetailRequest) (*model.ScoreezyAPIResponse[dataGenRetailV3], error) {
	url := fmt.Sprintf("%s/api/score/genretail/v3", repo.cfg.Env.ScoreezyHost)

	bodyBytes, err := repo.marshalFn(payload)
	if err != nil {
		return nil, errors.New(constant.ErrInvalidRequestPayload)
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, errors.New(constant.ErrMsgHTTPReqFailed)
	}

	req.Header.Set(constant.HeaderContentType, constant.HeaderApplicationJSON)
	req.Header.Set(constant.XUIDKey, memberId)

	q := req.URL.Query()
	q.Add("job_id", jobId)
	req.URL.RawQuery = q.Encode()

	res, err := repo.client.Do(req)
	if err != nil {
		return nil, errors.New(constant.ErrMsgHTTPReqFailed)
	}
	defer res.Body.Close()

	return helper.ParseScoreezyAPIResponse[dataGenRetailV3](res)
}

func (repo *repository) fetchLogsAPI(path string, query map[string]string) (*model.AifcoreAPIResponse[[]*logTransScoreezy], error) {
	url := fmt.Sprintf("%s%s", repo.cfg.Env.AifcoreHost, path)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, errors.New(constant.ErrMsgHTTPReqFailed)
	}
	req.Header.Set(constant.HeaderContentType, constant.HeaderApplicationJSON)

	q := req.URL.Query()
	for k, v := range query {
		q.Add(k, v)
	}
	req.URL.RawQuery = q.Encode()

	resp, err := repo.client.Do(req)
	if err != nil {
		return nil, errors.New(constant.ErrMsgHTTPReqFailed)
	}
	defer resp.Body.Close()

	apiResp, err := helper.ParseAifcoreAPIResponse[[]*logTransScoreezy](resp)
	if err != nil {
		return nil, err
	}

	return apiResp, nil
}

func (repo *repository) GetLogsScoreezyAPI(filter *filterLogs) (*model.AifcoreAPIResponse[[]*logTransScoreezy], error) {
	return repo.fetchLogsAPI("/api/core/logging/transaction/scoreezy/list",
		map[string]string{
			"company_id":   filter.CompanyId,
			"job_id":       filter.JobId,
			"name":         filter.Name,
			"grade":        filter.Grade,
			"product_type": filter.ProductType,
			"size":         filter.Size,
			"page":         filter.Page,
		})
}

func (repo *repository) GetLogsByRangeDateAPI(filter *filterLogs) (*model.AifcoreAPIResponse[[]*logTransScoreezy], error) {
	return repo.fetchLogsAPI("/api/core/logging/transaction/scoreezy/range",
		map[string]string{
			"company_id": filter.CompanyId,
			"job_id":     filter.JobId,
			"date_start": filter.StartDate,
			"date_end":   filter.EndDate,
			"grade":      filter.Grade,
			"size":       filter.Size,
			"page":       filter.Page,
		})
}

func (repo *repository) GetLogByTrxIdAPI(filter *filterLogs) (*logTransScoreezy, error) {
	url := fmt.Sprintf("%s/api/core/logging/transaction/scoreezy/%s", repo.cfg.Env.AifcoreHost, filter.TrxId)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, errors.New(constant.ErrMsgHTTPReqFailed)
	}

	req.Header.Set(constant.HeaderContentType, constant.HeaderApplicationJSON)

	resp, err := repo.client.Do(req)
	if err != nil {
		return nil, errors.New(constant.ErrMsgHTTPReqFailed)
	}
	defer resp.Body.Close()

	apiResp, err := helper.ParseAifcoreAPIResponse[*logTransScoreezy](resp)
	if err != nil {
		return nil, err
	}

	return apiResp.Data, nil
}

// func (repo *repository) StoreImportData(newData []*BulkSearch, userId string) error {
// 	errTx := repo.DB.Transaction(func(tx *gorm.DB) error {
// 		// remove data existing in table
// 		if err := repo.DB.Delete(&BulkSearch{}, "user_id = ?", userId).Error; err != nil {
// 			return err
// 		}

// 		// replace existing with new data
// 		if err := tx.Create(&newData).Error; err != nil {
// 			return err
// 		}

// 		return nil
// 	})

// 	if errTx != nil {
// 		return errTx
// 	}

// 	return nil
// }

// func (repo *repository) GetAllBulkSearch(tierLevel uint, userId, companyId string) ([]*BulkSearch, error) {
// 	var bulkSearches []*BulkSearch

// 	query := repo.DB.Preload("User")

// 	if tierLevel == 1 {
// 		// admin
// 		query = query.Where("company_id = ?", companyId)
// 	} else {
// 		// user
// 		query = query.Where("user_id = ?", userId)
// 	}

// 	err := query.Find(&bulkSearches)

// 	if err.Error != nil {
// 		return nil, err.Error
// 	}

// 	return bulkSearches, nil
// }

// func (repo *repository) CountData(tierLevel uint, userId, companyId string) (int64, error) {
// 	var bulkSearches []*BulkSearch
// 	var count int64

// 	query := repo.DB.Debug()

// 	if tierLevel == 1 {
// 		// admin
// 		query = query.Where("company_id = ?", companyId)
// 	} else {
// 		// user
// 		query = query.Where("user_id = ?", userId)
// 	}

// 	err := query.Find(&bulkSearches).Count(&count).Error

// 	return count, err
// }
