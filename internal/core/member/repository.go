package member

import (
	"bytes"
	"encoding/json"
	"fmt"
	"front-office/configs/application"
	"front-office/pkg/common/constant"
	"front-office/pkg/common/model"
	"front-office/pkg/helper"
	"front-office/pkg/httpclient"
	"front-office/pkg/jsonutil"
	"mime/multipart"
	"net/http"
)

func NewRepository(cfg *application.Config, client httpclient.HTTPClient, marshalFn jsonutil.Marshaller) Repository {
	if marshalFn == nil {
		marshalFn = json.Marshal
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
	AddMemberAPI(req *RegisterMemberRequest) (*registerResponseData, error)
	GetMemberAPI(query *MemberParams) (*MstMember, error)
	GetMemberListAPI(filter *MemberParams) ([]*MstMember, *model.Meta, error)
	GetSubscribedProducts(companyId, productSlug string) (*model.AifcoreAPIResponse[*subscribedProductRespData], error)
	GetQuotaAPI(query *QuotaParams) (*model.AifcoreAPIResponse[*quotaRespData], error)
	UpdateMemberAPI(id string, req map[string]interface{}) error
	DeleteMemberAPI(id string) error
}

func (repo *repository) AddMemberAPI(payload *RegisterMemberRequest) (*registerResponseData, error) {
	url := fmt.Sprintf("%s/api/core/member/addmember", repo.cfg.Env.AifcoreHost)

	var bodyBytes bytes.Buffer
	writer := multipart.NewWriter(&bodyBytes)

	writer.WriteField("name", payload.Name)
	writer.WriteField("email", payload.Email)
	writer.WriteField("key", payload.Key)
	writer.WriteField("companyid", fmt.Sprintf("%d", payload.CompanyId))
	writer.Close()

	req, err := http.NewRequest(http.MethodPost, url, &bodyBytes)
	if err != nil {
		return nil, fmt.Errorf(constant.ErrMsgHTTPReqFailed, err)
	}
	req.Header.Set(constant.HeaderContentType, writer.FormDataContentType())
	req.Header.Set(constant.XAPIKey, repo.cfg.Env.CoreModuleKey)

	resp, err := repo.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf(constant.ErrMsgHTTPReqFailed, err)
	}
	defer resp.Body.Close()

	apiResp, err := helper.ParseAifcoreAPIResponse[*registerResponseData](resp)
	if err != nil {
		return nil, err
	}

	return apiResp.Data, nil
}

func (repo *repository) GetMemberAPI(query *MemberParams) (*MstMember, error) {
	url := fmt.Sprintf(`%v/api/core/member/by`, repo.cfg.Env.AifcoreHost)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf(constant.ErrMsgHTTPReqFailed, err)
	}

	req.Header.Set(constant.HeaderContentType, constant.HeaderApplicationJSON)

	q := req.URL.Query()
	q.Add("id", query.Id)
	q.Add("company_id", query.CompanyId)
	q.Add("email", query.Email)
	q.Add("username", query.Username)
	q.Add("key", query.Key)
	req.URL.RawQuery = q.Encode()

	resp, err := repo.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf(constant.ErrMsgHTTPReqFailed, err)
	}
	defer resp.Body.Close()

	apiResp, err := helper.ParseAifcoreAPIResponse[*MstMember](resp)
	if err != nil {
		return nil, err
	}

	return apiResp.Data, nil
}

func (repo *repository) GetMemberListAPI(filter *MemberParams) ([]*MstMember, *model.Meta, error) {
	url := fmt.Sprintf(`%v/api/core/member/listbycompany/%v`, repo.cfg.Env.AifcoreHost, filter.CompanyId)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, nil, fmt.Errorf(constant.ErrMsgHTTPReqFailed, err)
	}

	req.Header.Set(constant.HeaderContentType, constant.HeaderApplicationJSON)

	q := req.URL.Query()
	q.Add(constant.Page, filter.Page)
	q.Add(constant.Size, filter.Limit)
	q.Add("keyword", filter.Keyword)
	q.Add(constant.StartDate, filter.StartDate)
	q.Add(constant.EndDate, filter.EndDate)
	req.URL.RawQuery = q.Encode()

	resp, err := repo.client.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf(constant.ErrMsgHTTPReqFailed, err)
	}
	defer resp.Body.Close()

	apiResp, err := helper.ParseAifcoreAPIResponse[[]*MstMember](resp)
	if err != nil {
		return nil, nil, err
	}

	return apiResp.Data, apiResp.Meta, nil
}

func (repo *repository) GetSubscribedProducts(companyId, productSlug string) (*model.AifcoreAPIResponse[*subscribedProductRespData], error) {
	url := fmt.Sprintf(`%v/api/core/member/subscribed-product/%v`, repo.cfg.Env.AifcoreHost, productSlug)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf(constant.ErrMsgHTTPReqFailed, err)
	}

	req.Header.Set(constant.HeaderContentType, constant.HeaderApplicationJSON)
	req.Header.Set(constant.XCompanyId, companyId)

	resp, err := repo.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf(constant.ErrMsgHTTPReqFailed, err)
	}
	defer resp.Body.Close()

	apiResp, err := helper.ParseAifcoreAPIResponse[*subscribedProductRespData](resp)
	if err != nil {
		return nil, err
	}

	return apiResp, nil
}

func (repo *repository) GetQuotaAPI(query *QuotaParams) (*model.AifcoreAPIResponse[*quotaRespData], error) {
	url := fmt.Sprintf(`%v/api/core/member/quota`, repo.cfg.Env.AifcoreHost)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf(constant.ErrMsgHTTPReqFailed, err)
	}

	req.Header.Set(constant.HeaderContentType, constant.HeaderApplicationJSON)

	q := req.URL.Query()
	q.Add("quota_type", query.QuotaType)
	q.Add("company_id", query.CompanyId)
	q.Add("member_id", query.MemberId)
	q.Add("subscribed_id", query.SubscribedId)
	req.URL.RawQuery = q.Encode()

	resp, err := repo.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf(constant.ErrMsgHTTPReqFailed, err)
	}
	defer resp.Body.Close()

	apiResp, err := helper.ParseAifcoreAPIResponse[*quotaRespData](resp)
	if err != nil {
		return nil, err
	}

	return apiResp, nil
}

func (repo *repository) UpdateMemberAPI(id string, payload map[string]interface{}) error {
	url := fmt.Sprintf(`%v/api/core/member/updateprofile/%v`, repo.cfg.Env.AifcoreHost, id)

	bodyBytes, err := repo.marshalFn(payload)
	if err != nil {
		return fmt.Errorf(constant.ErrMsgMarshalReqBody, err)
	}

	req, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(bodyBytes))
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

func (repo *repository) DeleteMemberAPI(id string) error {
	url := fmt.Sprintf(`%v/api/core/member/deletemember/%v`, repo.cfg.Env.AifcoreHost, id)

	req, err := http.NewRequest(http.MethodDelete, url, nil)
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
