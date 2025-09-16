package role

import (
	"fmt"
	"front-office/configs/application"
	"front-office/pkg/common/constant"
	"front-office/pkg/helper"
	"front-office/pkg/httpclient"
	"net/http"
)

func NewRepository(cfg *application.Config, client httpclient.HTTPClient) Repository {
	return &repository{cfg, client}
}

type repository struct {
	cfg    *application.Config
	client httpclient.HTTPClient
}

type Repository interface {
	GetRolesAPI(filter RoleFilter) ([]*MstRole, error)
	GetRoleByIdAPI(id string) (*MstRole, error)
}

func (repo *repository) GetRoleByIdAPI(id string) (*MstRole, error) {
	url := fmt.Sprintf(`%v/api/core/role/%v`, repo.cfg.Env.AifcoreHost, id)

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

	apiResp, err := helper.ParseAifcoreAPIResponse[*MstRole](resp)
	if err != nil {
		return nil, err
	}

	return apiResp.Data, nil
}

func (repo *repository) GetRolesAPI(filter RoleFilter) ([]*MstRole, error) {
	url := fmt.Sprintf(`%v/api/core/role`, repo.cfg.Env.AifcoreHost)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf(constant.ErrMsgHTTPReqFailed, err)
	}

	req.Header.Set(constant.HeaderContentType, constant.HeaderApplicationJSON)

	q := req.URL.Query()
	q.Add("name", filter.Name)
	req.URL.RawQuery = q.Encode()

	resp, err := repo.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf(constant.ErrMsgHTTPReqFailed, err)
	}
	defer resp.Body.Close()

	apiResp, err := helper.ParseAifcoreAPIResponse[[]*MstRole](resp)
	if err != nil {
		return nil, err
	}

	return apiResp.Data, nil
}
