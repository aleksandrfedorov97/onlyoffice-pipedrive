package response

import "encoding/json"

type DocSettingsResponse struct {
	DocAddress string `json:"doc_address"`
	DocSecret  string `json:"doc_secret"`
}

func (r DocSettingsResponse) ToJSON() []byte {
	buf, _ := json.Marshal(r)
	return buf
}
