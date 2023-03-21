package request

import "encoding/json"

type BuildConfigRequest struct {
	UID       int    `json:"uid"`
	CID       int    `json:"cid"`
	Deal      string `json:"deal_id"`
	UserAgent string `json:"user_agent"`
	FileID    string `json:"file_id"`
	Filename  string `json:"file_name"`
	DocKey    string `json:"doc_key"`
}

func (c BuildConfigRequest) ToJSON() []byte {
	buf, _ := json.Marshal(c)
	return buf
}
