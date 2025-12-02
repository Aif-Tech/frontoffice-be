package model

import (
	"strconv"
)

type AuthContext struct {
	UserId    uint
	CompanyId uint
	RoleId    uint
	QuotaType uint
	APIKey    string
}

func (a *AuthContext) UserIdStr() string {
	return strconv.FormatUint(uint64(a.UserId), 10)
}

func (a *AuthContext) CompanyIdStr() string {
	return strconv.FormatUint(uint64(a.CompanyId), 10)
}

func (a *AuthContext) IDs() (memberStr, companyStr string) {
	return strconv.FormatUint(uint64(a.UserId), 10), strconv.FormatUint(uint64(a.CompanyId), 10)
}

func (a *AuthContext) RoleIdStr() string {
	return strconv.FormatUint(uint64(a.RoleId), 10)
}

func (a *AuthContext) QuotaTypeStr() string {
	return strconv.FormatUint(uint64(a.QuotaType), 10)
}
