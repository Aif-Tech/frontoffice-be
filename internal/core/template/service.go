package template

import (
	"front-office/pkg/common/constant"
)

type Service interface {
	ListTemplates() (*Templates, error)
	DownloadTemplate(req DownloadRequest) (string, error)
}

type service struct {
	Repo Repository
}

func NewService(repo Repository) Service {
	return &service{Repo: repo}
}
func (s *service) DownloadTemplate(req DownloadRequest) (string, error) {
	return s.Repo.GetTemplatePath(req.Product, req.Filename)
}

func (s *service) ListTemplates() (*Templates, error) {
	availableTemplates, err := s.Repo.GetAvailableTemplates()
	if err != nil {
		return nil, err
	}

	var templates []TemplateInfo
	for product, files := range availableTemplates {
		desc := getTemplateDescription(product)

		templates = append(templates, TemplateInfo{
			Product:     product,
			Files:       files,
			Description: desc,
		})
	}

	result := &Templates{
		Templates: templates,
	}

	return result, nil
}

func getTemplateDescription(product string) string {
	switch product {
	case constant.PhoneLiveTemplates:
		return "Phone live status template"
	case constant.LoanRecordCheckerTemplates:
		return "Loan record checker template"
	case constant.MultipleLoanTemplates:
		return "Multiple loan template"
	case constant.TaxComplianceStatusTemplates:
		return "Tax compliance status template"
	case constant.TaxScoreTemplates:
		return "Tax score template"
	case constant.TaxVerificationTemplates:
		return "Tax verification detail template"
	case constant.GenRetailTemplates:
		return "Gen Retail v3 template"
	default:
		return "Common template"
	}
}
