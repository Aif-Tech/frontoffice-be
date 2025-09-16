package product

import (
	"front-office/pkg/apperror"
	"front-office/pkg/common/constant"
)

func NewService(repo Repository) Service {
	return &service{
		repo,
	}
}

type service struct {
	repo Repository
}

type Service interface {
	GetProductBySlug(slug string) (*productResponseData, error)
}

func (svc *service) GetProductBySlug(slug string) (*productResponseData, error) {
	product, err := svc.repo.GetProductAPI(slug)
	if err != nil {
		return nil, apperror.MapRepoError(err, constant.FailedFetchProduct)
	}
	if product.ProductId == 0 {
		return nil, apperror.NotFound(constant.ProductNotFound)
	}

	return product, err
}
