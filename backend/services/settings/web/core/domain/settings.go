package domain

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
)

type DocSettings struct {
	CompanyID  string `json:"company_id" mapstructure:"company_id"`
	DocAddress string `json:"doc_address" mapstructure:"doc_address"`
	DocSecret  string `json:"doc_secret" mapstructure:"doc_secret"`
}

func (u DocSettings) ToJSON() []byte {
	buf, _ := json.Marshal(u)
	return buf
}

func (u *DocSettings) Validate() error {
	u.CompanyID = strings.TrimSpace(u.CompanyID)
	u.DocAddress = strings.TrimSpace(u.DocAddress)
	u.DocSecret = strings.TrimSpace(u.DocSecret)

	if u.CompanyID == "" {
		return &InvalidModelFieldError{
			Model:  "Docserver",
			Field:  "CompanyID",
			Reason: "Should not be empty",
		}
	}

	url, err := url.Parse(u.DocAddress)
	if err != nil {
		return &InvalidModelFieldError{
			Model:  "Docserver",
			Field:  "Document Address",
			Reason: err.Error(),
		}
	}

	u.DocAddress = fmt.Sprintf("%s://%s/%s", url.Scheme, url.Host, url.Path)
	for {
		if strings.LastIndex(u.DocAddress, "/") == len(u.DocAddress)-1 {
			u.DocAddress = u.DocAddress[:len(u.DocAddress)-1]
		} else {
			break
		}
	}

	u.DocAddress += "/"

	if u.DocSecret == "" {
		return &InvalidModelFieldError{
			Model:  "Docserver",
			Field:  "Document Secret",
			Reason: "Should not be empty",
		}
	}

	return nil
}
