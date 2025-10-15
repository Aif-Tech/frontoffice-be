package constant

const (
	CSVHeaderName       = "Name"
	CSVHeaderIDCard     = "ID Card Number"
	CSVHeaderPhone      = "Phone Number"
	CSVHeaderLoanNumber = "Loan Number"
	CSVHeaderNPWP       = "NPWP"
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

var CSVHeaderTaxCompliance = []string{
	CSVHeaderNPWP,
	CSVHeaderLoanNumber,
}
