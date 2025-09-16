package constant

const (
	JobStatusPending    = "pending"
	JobStatusInProgress = "in_progress"
	JobStatusDone       = "done"
	JobStatusFailed     = "failed"
	JobStatusError      = "error"

	FormatDateAndTime = "2006-01-02 15:04:05"
	FormatYYYYMMDD    = "2006-01-02"

	SucceedGetLogTrans = "succeed to get list of log transaction"

	Success = "success"

	TestCaseSuccess          = "Success"
	TestCaseMarshalError     = "MarshalError"
	TestCaseNewRequestError  = "NewRequestError"
	TestCaseHTTPRequestError = "HTTPRequestError"
	TestCaseParseError       = "ParseError"
	InvalidJSON              = `{invalid-json`
)
