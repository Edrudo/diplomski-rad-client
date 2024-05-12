package application

type ImagePart struct {
	DataHash   string `json:"dataHash"`
	PartNumber int    `json:"partNumber"`
	TotalParts int    `json:"totalParts"`
	PartData   []byte `json:"partData"`
}
