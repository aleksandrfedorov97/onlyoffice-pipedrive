package request

import (
	"encoding/json"

	"github.com/golang-jwt/jwt"
)

type PipedriveTokenContext struct {
	jwt.StandardClaims
	UID int `json:"userId" mapstructure:"userId"`
	CID int `json:"companyId" mapstructure:"companyId"`
	IAT int `json:"iat" mapstructure:"iat"`
	EXP int `json:"exp" mapstructure:"exp"`
}

func (c PipedriveTokenContext) ToJSON() []byte {
	buf, _ := json.Marshal(c)
	return buf
}
