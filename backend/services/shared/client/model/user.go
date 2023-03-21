package model

import "strings"

type User struct {
	ID        int      `json:"id"`
	CompanyID int      `json:"company_id" mapstructure:"company_id"`
	Name      string   `json:"name"`
	Email     string   `json:"email"`
	Language  Language `json:"language" mapstructure:"language"`
	Access    []Access `json:"access" mapstructure:"access"`
}

func (u *User) Validate() error {
	u.Name = strings.TrimSpace(u.Name)
	u.Email = strings.TrimSpace(u.Email)

	if u.Name == "" {
		return ErrInvalidTokenFormat
	}

	if err := u.Language.Validate(); err != nil {
		return err
	}

	return nil
}
