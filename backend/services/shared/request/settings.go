package request

import (
	"encoding/json"
	"errors"
	"strings"
)

var _ErrInvalidCompanyID = errors.New("invalid company id")
var _ErrInvalidDocAddress = errors.New("invalid doc server address")
var _ErrInvalidDocSecret = errors.New("invalid doc server secret")
var _ErrInvalidDocHeader = errors.New("invalid doc server header")

type DocSettings struct {
	CompanyID  int    `json:"company_id" mapstructure:"company_id"`
	DocAddress string `json:"doc_address" mapstructure:"doc_address"`
	DocSecret  string `json:"doc_secret" mapstructure:"doc_secret"`
}

func (c DocSettings) ToJSON() []byte {
	buf, _ := json.Marshal(c)
	return buf
}

func (c DocSettings) Validate() error {
	c.DocAddress = strings.TrimSpace(c.DocAddress)
	c.DocSecret = strings.TrimSpace(c.DocSecret)

	if c.CompanyID <= 0 {
		return _ErrInvalidCompanyID
	}

	if c.DocAddress == "" {
		return _ErrInvalidDocAddress
	}

	if c.DocSecret == "" {
		return _ErrInvalidDocSecret
	}

	return nil
}
