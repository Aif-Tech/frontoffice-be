package grade

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
