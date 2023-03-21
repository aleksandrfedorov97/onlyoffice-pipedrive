package model

type Language struct {
	Code string `json:"country_code" mapstructure:"country_code"`
	Lang string `json:"language_code" mapstructure:"language_code"`
}

func (l *Language) Validate() error {
	if l.Code == "" {
		l.Code = "US"
	}

	if l.Lang == "" {
		l.Lang = "en"
	}

	return nil
}
