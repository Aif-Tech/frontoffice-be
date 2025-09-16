package transaction

import (
	"bytes"
	"fmt"
	"front-office/pkg/common/constant"
	"front-office/pkg/helper"
	"net/http"
)

func (repo *repository) CreateLogScoreezyAPI(payload *LogTransScoreezy) error {
	url := fmt.Sprintf("%s/api/core/logging/transaction/scoreezy", repo.cfg.Env.AifcoreHost)

	bodyBytes, err := repo.marshalFn(payload)
	if err != nil {
		return fmt.Errorf(constant.ErrMsgMarshalReqBody, err)
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return fmt.Errorf(constant.ErrMsgHTTPReqFailed, err)
	}

	req.Header.Set(constant.HeaderContentType, constant.HeaderApplicationJSON)

	resp, err := repo.client.Do(req)
	if err != nil {
		return fmt.Errorf(constant.ErrMsgHTTPReqFailed, err)
	}
	defer resp.Body.Close()

	_, err = helper.ParseAifcoreAPIResponse[any](resp)
	if err != nil {
		return err
	}

	return nil
}

func (repo *repository) GetLogsScoreezyAPI() ([]*LogTransScoreezy, error) {
	url := fmt.Sprintf("%s/api/core/logging/transaction/scoreezy/list", repo.cfg.Env.AifcoreHost)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf(constant.ErrMsgHTTPReqFailed, err)
	}

	req.Header.Set(constant.HeaderContentType, constant.HeaderApplicationJSON)

	resp, err := repo.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf(constant.ErrMsgHTTPReqFailed, err)
	}
	defer resp.Body.Close()

	apiResp, err := helper.ParseAifcoreAPIResponse[[]*LogTransScoreezy](resp)
	if err != nil {
		return nil, err
	}

	return apiResp.Data, nil
}

func (repo *repository) GetLogsScoreezyByDateAPI(companyId, date string) ([]*LogTransScoreezy, error) {
	url := fmt.Sprintf("%s/api/core/logging/transaction/scoreezy/by", repo.cfg.Env.AifcoreHost)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf(constant.ErrMsgHTTPReqFailed, err)
	}

	req.Header.Set(constant.HeaderContentType, constant.HeaderApplicationJSON)

	q := req.URL.Query()
	q.Add("company_id", companyId)
	q.Add("date", date)
	req.URL.RawQuery = q.Encode()

	resp, err := repo.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf(constant.ErrMsgHTTPReqFailed, err)
	}
	defer resp.Body.Close()

	apiResp, err := helper.ParseAifcoreAPIResponse[[]*LogTransScoreezy](resp)
	if err != nil {
		return nil, err
	}

	return apiResp.Data, nil
}

func (repo *repository) GetLogsScoreezyByDateRangeAPI(companyId, startDate, endDate string) ([]*LogTransScoreezy, error) {
	url := fmt.Sprintf("%s/api/core/logging/transaction/scoreezy/range", repo.cfg.Env.AifcoreHost)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf(constant.ErrMsgHTTPReqFailed, err)
	}

	req.Header.Set(constant.HeaderContentType, constant.HeaderApplicationJSON)

	q := req.URL.Query()
	q.Add("date_start", startDate)
	q.Add("date_end", endDate)
	q.Add("company_id", companyId)
	req.URL.RawQuery = q.Encode()

	resp, err := repo.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf(constant.ErrMsgHTTPReqFailed, err)
	}
	defer resp.Body.Close()

	apiResp, err := helper.ParseAifcoreAPIResponse[[]*LogTransScoreezy](resp)
	if err != nil {
		return nil, err
	}

	return apiResp.Data, nil
}

func (repo *repository) GetLogsScoreezyByMonthAPI(companyId, month string) ([]*LogTransScoreezy, error) {
	url := fmt.Sprintf("%s/api/core/logging/transaction/scoreezy/month", repo.cfg.Env.AifcoreHost)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf(constant.ErrMsgHTTPReqFailed, err)
	}

	req.Header.Set(constant.HeaderContentType, constant.HeaderApplicationJSON)

	q := req.URL.Query()
	q.Add("company_id", companyId)
	q.Add("month", month)
	req.URL.RawQuery = q.Encode()

	resp, err := repo.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf(constant.ErrMsgHTTPReqFailed, err)
	}
	defer resp.Body.Close()

	apiResp, err := helper.ParseAifcoreAPIResponse[[]*LogTransScoreezy](resp)
	if err != nil {
		return nil, err
	}

	return apiResp.Data, nil
}
