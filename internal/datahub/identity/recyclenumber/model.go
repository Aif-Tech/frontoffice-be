package recyclenumber

type recycleNumberRequest struct {
	Phone  string `json:"phone_number" validate:"required~Phone Number cannot be empty, indophone, min(9)"`
	LoanNo string `json:"loan_no" validate:"required~Loan No cannot be empty."`
}

type dataRecycleNumberAPI struct {
	Status string `json:"status"`
}
