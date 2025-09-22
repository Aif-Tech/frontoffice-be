package member

import (
	"front-office/internal/core/company"
	"front-office/internal/core/role"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type MstMember struct {
	MemberId          uint               `json:"member_id" gorm:"primaryKey;autoIncrement"`
	Name              string             `json:"name" validate:"required,alphaspace, min(3)"`
	Username          string             `json:"username" gorm:"unique"`
	Email             string             `json:"email" gorm:"unique" validate:"required,email"`
	Password          string             `json:"password"`
	Phone             string             `json:"phone"`
	Key               string             `json:"api_key" gorm:"uniqueIndex"`
	Active            bool               `json:"active"`
	ParentId          string             `json:"parent_id"`
	CompanyId         uint               `json:"company_id"`
	MstCompany        company.MstCompany `json:"company" gorm:"foreignKey:CompanyId"`
	RoleId            uint               `json:"role_id"`
	Role              role.MstRole       `json:"role" gorm:"foreignKey:RoleId"`
	Status            bool               `json:"status"`
	MailStatus        string             `json:"mail_status" gorm:"default:pending"`
	AccountType       string             `json:"account_type"`
	ProductPermission string             `json:"product_permission"`
	IsVerified        bool               `json:"is_verified"`
	Image             string             `json:"image" gorm:"default:default-profile-image.jpg"`
	QuotaType         int8               `json:"quota_type"` //0: none, 1: Quota Total 2: Quota per product
	Quota             int                `json:"quota"`
	CreatedAt         time.Time          `json:"-"`
	UpdatedAt         time.Time          `json:"-"`
	DeletedAt         gorm.DeletedAt     `json:"-" gorm:"index"`
}

type RegisterMemberRequest struct {
	Name      string `json:"name" validate:"required~Field Name is required"`
	Email     string `json:"email" validate:"required~Field Email is required, email~Only email pattern are allowed"`
	Username  string `json:"username"`
	Phone     string `json:"phone" validate:"phone"`
	CompanyId uint   `json:"company_id"`
	RoleId    uint   `json:"role_id"`
	Key       string `json:"key"`
}

type registerResponseData struct {
	MemberId uint `json:"member_id"`
}

type updateUserRequest struct {
	Name       *string `json:"name"`
	Email      *string `json:"email" validate:"email~Only email pattern are allowed"`
	MailStatus *string `json:"mail_status"`
	RoleId     *string `json:"role_id"`
	Active     *bool   `json:"active"`
}

type userUpdateResponse struct {
	Id        uint   `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	Active    bool   `json:"active"`
	CompanyId uint   `json:"company_id"`
	RoleId    uint   `json:"role_id"`
}

type updateProfileRequest struct {
	Name  *string `json:"name"`
	Email *string `json:"email"`
}

func SetPassword(password string) string {
	hashedPass, _ := bcrypt.GenerateFromPassword([]byte(password), 12)
	password = string(hashedPass)

	return password
}

type MemberParams struct {
	Id        string
	CompanyId string
	Email     string
	Username  string
	Key       string
	Page      string
	Limit     string
	Keyword   string
	RoleName  string
	RoleId    string
	Status    string
	StartDate string
	EndDate   string
}

type QuotaParams struct {
	MemberId     string `json:"member_id"`
	CompanyId    string `json:"company_id"`
	SubscribedId string `json:"subscribed_id"`
	QuotaType    string `json:"quota_type"`
}

type quotaRespData struct {
	Quota int `json:"quota"`
}

type subscribedProductRespData struct {
	SubsribedProductID uint        `json:"subscribed_product_id"`
	CompanyId          uint        `json:"company_id"`
	ProductId          uint        `json:"product_id"`
	Product            productData `json:"product"`
	QuotaProduct       int         `json:"quota_product"`
}

type productData struct {
	ProductId      uint   `json:"product_id"`
	ProductGroupId uint   `json:"product_group_id"`
	ProductName    string `json:"product_name"`
	ProductSlug    string `json:"product_slug_name"`
}

const (
	memberRoleID = 2
)
