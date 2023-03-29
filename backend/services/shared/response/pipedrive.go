package response

import "encoding/json"

type AddFileResponse struct {
	Success bool `json:"success"`
	Data    struct {
		ID         int    `json:"id"`
		Filename   string `json:"file_name"`
		DealID     int    `json:"deal_id"`
		UpdateTime string `json:"update_time"`
	} `json:"data"`
}

func (r AddFileResponse) ToJSON() []byte {
	buf, _ := json.Marshal(r)
	return buf
}
