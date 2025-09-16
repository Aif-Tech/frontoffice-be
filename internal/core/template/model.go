package template

type DownloadRequest struct {
	Product  string `query:"product"`
	Filename string `query:"filename"`
}

type Templates struct {
	Templates []TemplateInfo `json:"templates"`
}

type TemplateInfo struct {
	Product     string   `json:"product"`
	Files       []string `json:"files"`
	Description string   `json:"description"`
}
