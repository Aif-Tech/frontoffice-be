package operation

import (
	"front-office/pkg/apperror"
	"front-office/pkg/common/constant"
)

func NewService(repo Repository) Service {
	return &service{repo}
}

type service struct {
	repo Repository
}

type Service interface {
	GetLogsOperation(filter *logOperationFilter) (*logOperationAPIResponse, error)
	GetLogsByRange(filter *logRangeFilter) (*logOperationAPIResponse, error)
	AddLogOperation(req *AddLogRequest) error
}

func (svc *service) GetLogsOperation(filter *logOperationFilter) (*logOperationAPIResponse, error) {
	result, err := svc.repo.GetLogsOperationAPI(filter)
	if err != nil {
		return nil, apperror.MapRepoError(err, "failed to fetch log operations")
	}

	response := &logOperationAPIResponse{
		Message: constant.SucceedGetLogTrans,
		Success: result.Success,
		Data: &logOperationRespData{
			Logs: result.Data,
		},
		Meta: *result.Meta,
	}

	return response, nil
}

func (svc *service) GetLogsByRange(filter *logRangeFilter) (*logOperationAPIResponse, error) {
	result, err := svc.repo.GetLogsByRangeAPI(filter)
	if err != nil {
		return nil, apperror.MapRepoError(err, "failed to fetch log operations")
	}

	response := &logOperationAPIResponse{
		Message: constant.SucceedGetLogTrans,
		Success: result.Success,
		Data: &logOperationRespData{
			Logs: result.Data,
		},
		Meta: *result.Meta,
	}

	return response, nil
}

func (svc *service) AddLogOperation(req *AddLogRequest) error {
	if err := svc.repo.AddLogOperation(req); err != nil {
		return apperror.MapRepoError(err, "failed to create log")
	}

	return nil
}
