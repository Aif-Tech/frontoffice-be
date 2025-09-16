package company

import (
	"time"

	"gorm.io/gorm"
)

type MstCompany struct {
	CompanyId       uint    `json:"company_id" gorm:"primaryKey;autoIncrement"`
	CompanyName     string  `json:"company_name"`
	CompanyAddress  string  `json:"company_address"`
	CompanyPhone    string  `json:"company_phone"`
	AgreementNumber string  `json:"agreement_number"`
	PaymentScheme   string  `json:"payment_scheme"`
	PostpaidActive  bool    `json:"active"`
	IndustryId      uint    `json:"industry_id"`
	BasePricing     float64 `json:"base_pricing"`
	// Apiconfigs      []apiconfig.MstApiconfig `json:"apiconfigs" gorm:"foreignKey:CompanyId"`
	// Products        []MstSubscribedProduct   `json:"products" gorm:"foreignKey:CompanyId"`
	CreatedAt time.Time      `json:"-"`
	UpdatedAt time.Time      `json:"-"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}
