package operation

import (
	"front-office/pkg/common/model"
	"time"
)

type LogOperation struct {
	LogOpsID  uint      `json:"log_ops_id" gorm:"primaryKey;autoIncrement"`
	MemberId  uint      `json:"member_id"`
	Member    mstMember `json:"member" gorm:"foreignKey:MemberId"`
	CompanyId uint      `json:"company_id"`
	Module    string    `json:"module"`
	Action    string    `json:"action"`
	ClientIP  string    `json:"ip_address"`
	CreatedAt time.Time `json:"created_at"`
}

type mstMember struct {
	MemberID uint    `json:"member_id" gorm:"primaryKey;autoIncrement"`
	Name     string  `json:"name"`
	RoleId   uint    `json:"role_id"`
	Role     mstRole `json:"role"`
}

type mstRole struct {
	RoleId uint   `json:"role_id" gorm:"primaryKey;autoIncrement"`
	Name   string `json:"name"`
}

type logOperationFilter struct {
	CompanyId string
	Page      string
	Size      string
	Role      string
	Event     string
	Name      string
	StartDate string
	EndDate   string
}

type logRangeFilter struct {
	CompanyId string
	Page      string
	Size      string
	StartDate string
	EndDate   string
}

type AddLogRequest struct {
	MemberId  uint   `json:"member_id" validate:"required~Field Member ID is required"`
	CompanyId uint   `json:"company_id" validate:"required~Field Company ID is required"`
	Action    string `json:"action" validate:"required~Field Action is required"`
}

type logOperationAPIResponse struct {
	Message string                `json:"message"`
	Success bool                  `json:"success"`
	Data    *logOperationRespData `json:"data"`
	Meta    model.Meta            `json:"meta"`
}

type logOperationRespData struct {
	Logs interface{} `json:"logs"`
}
