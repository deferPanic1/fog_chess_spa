package dto

import (
	"fmt"
	"net/mail"
	"strings"
)

type RegisterRequest struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password"`
}

func (r *RegisterRequest) Validate() error {
	r.Email = strings.TrimSpace(r.Email)
	r.Username = strings.TrimSpace(r.Username)

	if r.Username == "" || len(r.Username) < 3 || len(r.Username) > 20 {
		return fmt.Errorf("username must be 3-20 characters")
	}
	if _, err := mail.ParseAddress(r.Email); err != nil {
		return fmt.Errorf("invalid email format")
	}
	if len(r.Password) < 6 {
		return fmt.Errorf("password must be at least 6 characters")
	}
	return nil
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (r *LoginRequest) Validate() error {
	r.Username = strings.TrimSpace(r.Username)
	if r.Username == "" {
		return fmt.Errorf("username is required")
	}
	if r.Password == "" {
		return fmt.Errorf("password is required")
	}
	return nil
}

type LoginResponse struct {
	AccessToken string       `json:"accessToken"`
	User        UserResponse `json:"user"`
}
