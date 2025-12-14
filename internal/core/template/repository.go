package template

import (
	"front-office/pkg/common/constant"
	"os"
	"path/filepath"
)

func NewRepository() Repository {
	return &repository{}
}

type repository struct{}

type Repository interface {
	GetAvailableTemplates() (map[string][]string, error)
	GetTemplatePath(category, filename string) (string, error)
}

func (r *repository) GetTemplatePath(category, filename string) (string, error) {
	path := filepath.Join(constant.TemplateBaseDir, category, filename)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return "", err
	}

	return path, nil
}

func (r *repository) GetAvailableTemplates() (map[string][]string, error) {
	products := []string{
		constant.PhoneLiveTemplates,
		constant.RecycleNumberTemplates,
		constant.LoanRecordCheckerTemplates,
		constant.MultipleLoanTemplates,
		constant.NPWPVerificationTemplates,
		constant.TaxComplianceStatusTemplates,
		constant.TaxScoreTemplates,
		constant.TaxVerificationTemplates,
		constant.GenRetailTemplates,
	}

	result := make(map[string][]string)

	for _, category := range products {
		files, err := os.ReadDir(filepath.Join(constant.TemplateBaseDir, category))
		if err != nil {
			return nil, err
		}

		var filenames []string
		for _, file := range files {
			filenames = append(filenames, file.Name())
		}

		result[category] = filenames
	}

	return result, nil
}
