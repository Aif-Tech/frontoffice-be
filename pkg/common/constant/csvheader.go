package constant

const (
	CSVHeaderName       = "Name"
	CSVHeaderIDCard     = "ID Card Number"
	CSVHeaderPhone      = "Phone Number"
	CSVHeaderLoanNumber = "Loan Number"
)

var CSVHeaderLoanRecord = []string{
	CSVHeaderName,
	CSVHeaderIDCard,
	CSVHeaderPhone,
	CSVHeaderLoanNumber,
}

var CSVHeaderMultipleLoan = []string{
	CSVHeaderIDCard,
	CSVHeaderPhone,
	CSVHeaderLoanNumber,
}

var CSVHeaderPhoneLive = []string{
	CSVHeaderPhone,
	CSVHeaderLoanNumber,
}
