package request

import (
	"encoding/json"

	"github.com/golang-jwt/jwt"
)

type BaseCommandRequest struct {
	jwt.StandardClaims
	C string `json:"c"`
}

func (c BaseCommandRequest) ToJSON() []byte {
	buf, _ := json.Marshal(c)
	return buf
}

type TokenCommandRequest struct {
	Token string `json:"token"`
}

func (c TokenCommandRequest) ToJSON() []byte {
	buf, _ := json.Marshal(c)
	return buf
}
