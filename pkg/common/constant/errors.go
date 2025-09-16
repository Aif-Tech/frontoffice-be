package constant

const (
	// general
	DataAlreadyExist     = "data already exists"
	DataNotFound         = "data not found"
	FileSizeIsTooLarge   = "file size should not exceed 200 KB"
	FailedToUploadImage  = "failed to upload image"
	InvalidDateFormat    = "invalid date format"
	InvalidRequestFormat = "invalid request format"
	RecordNotFound       = "record not found"
	OnlyUploadCSVfile    = "only CSV files allowed"
	TemplateNotFound     = "template not found"
	UpstreamError        = "upstream error"

	// auth
	AlreadyVerified            = "the account has already verified"
	ActivationTokenExpired     = "user activation token has expired"
	ConfirmNewPasswordMismatch = "ensure that the new password and confirm password fields match exactly"
	ConfirmPasswordMismatch    = "ensure that password and confirm password fields match exactly"
	InvalidEmailOrPassword     = "email or password is incorrect"
	InvalidPassword            = "password must contain a combination of uppercase, lowercase, number, and symbol"
	InvalidPasswordResetLink   = "invalid password reset link"
	InvalidActivationLink      = "invalid activation link"
	IncorrectPassword          = "password is incorrect"
	RequestProhibited          = "request is prohibited"
	UnverifiedUser             = "please check your email, you need to verify your email address before you can reset your password"
	UserNotFoundForgotEmail    = "if your account exists, we've sent password reset instructions to your email"
	UserNotFound               = "user not found"
	TokenExpired               = "Token is expired"
	BcryptPasswordMismatch     = "crypto/bcrypt: hashedPassword is not the hash of the given password"
	WrongCurrentPassword       = "current password is wrong"

	//grading
	DuplicateGrading       = "duplicate grading"
	FieldGradingLabelEmpty = "field grading label is required"
	FieldMinGradeEmpty     = "field min grade is required"
	FieldMaxGradeEmpty     = "field max grade is required"
	FieldGradingValueEmpty = "field grading value is required"

	// gen-retail
	InvalidDocumentFile    = "invalid document file"
	ErrorGettingFile       = "error getting file"
	ErrorOpeningFile       = "error opening file"
	ErrorReadingCSV        = "error reading CSV file"
	HeaderTemplateNotValid = "header template is not valid"
	ErrorReadingCSVRecords = "error reading CSV records"
	ErrorUploadDataCSV     = "error upload data CSV file"
	FailedParseCSV         = "failed to parse csv"

	//parameter settings
	ParamSettingIsNotSet = "parameter settings is not set"

	EmailAlreadyExists = "email already exists"
	InvalidImageFile   = "invalid image file"
	InvalidStatusValue = "invalid value for 'status'"
	SendEmailFailed    = "send email failed"

	ProductNotFound = "product not found"

	ErrFailedMarshalReq   = "failed to marshal request body"
	ErrHTTPReqFailed      = "failed to make HTTP request"
	FailedFetchMember     = "failed to fetch member"
	FailedUpdateMember    = "failed to update member"
	FailedFetchLogs       = "failed to fetch logs"
	FailedFetchProduct    = "failed to fetch product"
	FailedCreateJob       = "failed to create job"
	InvalidUserSession    = "invalid user session"
	InvalidCompanySession = "invalid company session"
	MissingUserId         = "missing user id"
	MissingAccessToken    = "no access token provided"
	MissingStartDate      = "start_date is required"
	MissingEndDate        = "end_date is required"

	ErrMsgMarshalReqBody = "failed to marshal request body: %w"
	ErrMsgHTTPReqFailed  = "HTTP request failed: %w"

	ErrCreatePhoneLiveJob       = "failed to create phone live status job"
	ErrFetchPhoneLiveDetail     = "failed to fetch phone live status job detail"
	ErrFetchJobMetrics          = "failed to fetch job metrics"
	ErrMsgUpdatePhoneLiveJob    = "failed to update phone live status job"
	ErrMsgUpdatePhoneLiveDetail = "failed to update phone live status job detail"
)
