package transaction

func NewService(repo Repository) Service {
	return &service{repo}
}

type service struct {
	repo Repository
}

type Service interface {
	// scoreezy
	GetScoreezyLogs() ([]*scoreezyLogResponse, error)
	GetScoreezyLogsByDate(companyId, date string) ([]*scoreezyLogResponse, error)
	GetScoreezyLogsByDateRange(startDate, endDate, companyId, page string) ([]*scoreezyLogResponse, error)
	GetScoreezyLogsByMonth(companyId, month string) ([]*scoreezyLogResponse, error)

	// product catalog
	GetProcessedLogCount(jobId string) (*getProcessedCountResp, error)
	UpdateLogProCat(transId string, req *UpdateTransRequest) error
}
