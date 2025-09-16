package transaction

import (
	"encoding/json"
	"front-office/configs/application"
	"front-office/pkg/httpclient"
	"front-office/pkg/jsonutil"
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
	// scoreezy
	CreateLogScoreezyAPI(req *LogTransScoreezy) error
	GetLogsScoreezyAPI() ([]*LogTransScoreezy, error)
	GetLogsScoreezyByDateAPI(companyId, date string) ([]*LogTransScoreezy, error)
	GetLogsScoreezyByDateRangeAPI(companyId, startDate, endDate string) ([]*LogTransScoreezy, error)
	GetLogsScoreezyByMonthAPI(companyId, month string) ([]*LogTransScoreezy, error)

	// product catalog
	CreateLogTransAPI(req *LogTransProCatRequest) error
	GetLogTransByJobIdAPI(jobId, companyId string) ([]*LogTransProductCatalog, error)
	ProcessedLogCountAPI(jobId string) (*getProcessedCountResp, error)
	UpdateLogTransAPI(transId string, req map[string]interface{}) error
}
