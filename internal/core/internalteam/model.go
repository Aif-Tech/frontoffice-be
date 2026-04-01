package internalteam

import (
	"time"

	"gorm.io/gorm"
)

type registerRequest struct {
	Name  string `json:"name" validate:"required~Field Name is required"`
	Email string `json:"email" validate:"required~Field Email is required, email~Only email pattern are allowed"`
}

type MstInternalTeam struct {
	MemberID  uint           `json:"member_id"`
	Name      string         `json:"name"`
	Email     string         `json:"email"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"-"`
	DeletedAt gorm.DeletedAt `json:"-"`
}
