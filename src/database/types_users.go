package database

import (
	"ExamSphere/src/core/appValues"
	"time"
)

// UserInfo is a struct that holds the information of a user.
type UserInfo struct {
	// UserId is the user-id of the user.
	UserId string `json:"user_id"`

	// AuthHash the hash used to invalidate when the user changes their
	// password. User CANNOT login only with this hash; this is just
	// a security measure and has no other usage.
	AuthHash string `json:"auth_hash"`

	// FullName is the user's full name.
	FullName string `json:"full_name"`

	// Email is the email of the user.
	Email string `json:"email"`

	// Password is the hashed password of the user.
	Password string `json:"password"`

	// Role is the role of the user.
	Role appValues.UserRole `json:"role"`

	// IsBanned is true if and only if the user is banned.
	IsBanned bool `json:"is_banned"`

	// BanReason is the user's ban reason (from the platform). optional.
	BanReason *string `json:"ban_reason"`

	// CreatedAt is the time when the user was created.
	CreatedAt time.Time `json:"created_at"`

	// UserAddress is the user's address. optional.
	UserAddress *string `json:"user_address"`

	// PhoneNumber is the user's phone number. optional.
	PhoneNumber *string `json:"phone_number"`

	// SetupCompleted is true if the user has completed the account
	// setup process.
	SetupCompleted bool
}

// NewUserData is used to create a new user.
type NewUserData struct {
	UserId         string             `json:"user_id"`
	FullName       string             `json:"full_name"`
	Email          string             `json:"email"`
	RawPassword    string             `json:"password"`
	Role           appValues.UserRole `json:"role"`
	UserAddress    *string            `json:"user_address"`
	PhoneNumber    *string            `json:"phone_number"`
	SetupCompleted bool               `json:"setup_completed"`
}

// SearchUserData is used to search for users.
type SearchUserData struct {
	Query  string `json:"query"`
	Offset int    `json:"offset"`
	Limit  int    `json:"limit" validate:"min=1"`
}

type UpdateUserData struct {
	UserId      string  `json:"user_id"`
	FullName    string  `json:"full_name"`
	Email       string  `json:"email"`
	UserAddress *string `json:"user_address"`
	PhoneNumber *string `json:"phone_number"`
}

type BanUserData struct {
	UserId    string  `json:"user_id"`
	IsBanned  bool    `json:"is_banned"`
	BanReason *string `json:"ban_reason"`
}

type UpdateUserPasswordData struct {
	UserId      string `json:"user_id"`
	RawPassword string `json:"password"`
}
