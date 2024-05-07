package application

type ImagePart struct {
	ImageHash  string `json:"imageHash"`
	PartNumber int    `json:"partNumber"`
	TotalParts int    `json:"totalParts"`
	PartData   []byte `json:"partData"`
}
