package constant

const (
	HeaderContentType        = "Content-Type"
	HeaderContentDisposition = "Content-Disposition"
	HeaderApplicationJSON    = "application/json"
	XAPIKey                  = "X-API-KEY"
	XUIDKey                  = "X-UID-KEY"
	XMemberId                = "X-Member-ID"
	XCompanyId               = "X-Company-ID"
	XTierLevel               = "X-Tier-Level"
	TextOrCSVContentType     = "text/csv"

	SizeUnlimited = "-1"

	Request       = "request"
	APIKey        = "apiKey"
	UserId        = "userId"
	CompanyId     = "companyId"
	RoleId        = "roleId"
	ValidatedFile = "validatedFile"

	QuotaType  = "quota_type"
	Page       = "page"
	Size       = "size"
	JobId      = "job_id"
	Masked     = "masked"
	StartDate  = "start_date"
	EndDate    = "end_date"
	Keyword    = "keyword"
	MailStatus = "mail_status"
	Role       = "role"
	Status     = "status"
	SortBy     = "sort_by"
	Order      = "order"

	MockHost        = "http://mock-host"
	MockInvalidHost = "http://[::1]:namedport"
)
