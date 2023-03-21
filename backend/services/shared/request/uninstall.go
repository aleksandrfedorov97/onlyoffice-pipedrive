package request

import "encoding/json"

type UninstallRequest struct {
	ClientID  string `json:"client_id"`
	CompanyID int    `json:"company_id"`
	UserID    int    `json:"user_id"`
	Timestamp string `json:"timestamp"`
}

func (r UninstallRequest) ToJSON() []byte {
	buf, _ := json.Marshal(r)
	return buf
}
