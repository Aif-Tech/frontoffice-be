package grade

import (
	"front-office/internal/core/company"
	"time"

	"gorm.io/gorm"
)

type Grading struct {
	Id           string             `gorm:"primarykey" json:"id"`
	GradingLabel string             `gorm:"not null" json:"grading_label"`
	MinGrade     float64            `gorm:"not null" json:"min_grade"`
	MaxGrade     float64            `gorm:"not null" json:"max_grade"`
	CompanyId    string             `json:"company_id"`
	Company      company.MstCompany `gorm:"foreignKey:CompanyId" json:"-"`
	CreatedAt    time.Time          `json:"-"`
	UpdatedAt    time.Time          `json:"-"`
	DeletedAt    gorm.DeletedAt     `gorm:"index" json:"-"`
}

type MstGrade struct {
	Id    uint    `json:"id"`
	Grade string  `json:"grade"`
	Start float64 `json:"start"`
	End   float64 `json:"end"`
}

type createGradeRequest struct {
	Grades []gradeInput `json:"grades" validate:"required"`
}

type gradeInput struct {
	Grade string  `json:"grade" validate:"required"`
	Start float64 `json:"start" validate:"range(0|1)"`
	End   float64 `json:"end" validate:"range(0|1)"`
}

type createGradePayload struct {
	CompanyId   string             `json:"company_id"`
	ProductSlug string             `json:"product_slug"`
	Request     createGradeRequest `json:"grades"`
}

type refGrade struct {
	SubscribedProductID uint    `json:"-"`
	Grade               string  `json:"grade"`
	Start               float64 `json:"start"`
	End                 float64 `json:"end"`
}

type gradesResponseData struct {
	Grades []refGrade `json:"grades"`
}
