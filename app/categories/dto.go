package categories

type Category struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

type ListResponse struct {
	Categories []Category `json:"categories"`
}

// CreateInput is the untrimmed input of a create request.
type CreateInput struct {
	Code string
	Name string
}
