package template

import (
	"errors"
	"fmt"
	"front-office/pkg/apperror"
	"front-office/pkg/common/constant"
	"os"
	"regexp"
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

const (
	defaultTemplateVersion = "1.0.0"
	templateFilePattern    = "template-v%s.csv"
)

func (s *service) DownloadTemplate(req DownloadRequest) (string, error) {
	if req.Product == "" {
		return "", apperror.BadRequest("product parameter is required")
	}

	version := req.Version
	if version == "" {
		version = defaultTemplateVersion
	} else if !regexp.MustCompile(`^\d+\.\d+\.\d+$`).MatchString(version) {
		return "", apperror.BadRequest("invalid version format (expected x.y.z)")
	}

	filename := fmt.Sprintf(templateFilePattern, version)
	path, err := s.Repo.GetTemplatePath(req.Product, filename)
	if errors.Is(err, os.ErrNotExist) {
		return s.Repo.GetTemplatePath(req.Product, fmt.Sprintf(templateFilePattern, defaultTemplateVersion))
	}

	return path, nil
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
