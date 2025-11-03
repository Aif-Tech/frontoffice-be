package template

type DownloadRequest struct {
	Product string `query:"product"`
	Version string `query:"version"`
}

type Templates struct {
	Templates []TemplateInfo `json:"templates"`
}

type TemplateInfo struct {
	Product     string   `json:"product"`
	Files       []string `json:"files"`
	Description string   `json:"description"`
}
