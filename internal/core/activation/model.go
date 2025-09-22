package activation

import (
	"front-office/internal/core/member"
	"time"
)

type MstActivationToken struct {
	Id        string           `gorm:"primarykey" json:"id"`
	Token     string           `gorm:"not null" json:"token"`
	MemberId  uint             `json:"member_id"`
	Member    member.MstMember `gorm:"foreignKey:MemberId" json:"-"`
	CreatedAt time.Time        `json:"created_at"`
}

type CreateActivationTokenRequest struct {
	Token string `json:"token"`
}

type CreateActivationTokenResponse struct {
	Message    string              `json:"message"`
	Success    bool                `json:"success"`
	Data       *MstActivationToken `json:"data"`
	StatusCode int                 `json:"-"`
}
