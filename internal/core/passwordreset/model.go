package passwordreset

import (
	"front-office/internal/core/member"
	"time"

	"gorm.io/gorm"
)

type MstPasswordResetToken struct {
	Id         uint             `gorm:"primarykey" json:"id"`
	Token      string           `gorm:"not null" json:"token"`
	Activation bool             `gorm:"not null;default:false" json:"activation"`
	MemberId   uint             `json:"member_id"`
	Member     member.MstMember `gorm:"foreignKey:MemberId" json:"member"`
	CreatedAt  time.Time        `json:"created_at"`
	UpdatedAt  time.Time        `json:"-"`
	DeletedAt  gorm.DeletedAt   `gorm:"index" json:"-"`
}

type CreatePasswordResetTokenRequest struct {
	Token string `json:"token"`
}
