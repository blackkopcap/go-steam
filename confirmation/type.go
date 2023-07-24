package confirmation

type Confirmation struct {
	Type         uint64      `json:"type"`
	TypeName     string      `json:"type_name"`
	ID           string      `json:"id"`
	CreatorID    string      `json:"creator_id"`
	Nonce        string      `json:"nonce"`
	CreationTime uint64      `json:"creation_time"`
	Cancel       string      `json:"cancel"`
	Accept       string      `json:"accept"`
	Icon         string      `json:"icon"`
	Multi        bool        `json:"multi"`
	Headline     string      `json:"headline"`
	Summary      []string    `json:"summary"`
	Warn         interface{} `json:"warn"`
}

type jsonObj = map[string]interface{}
