package contract

type ErrorResponse struct {
	Error string `json:"error"`
}

type AddRequest struct {
	Group string            `json:"g"`
	Kind  string            `json:"k"`
	Param map[string]string `json:"p"`
}

type AddResponse struct {
	Id string `json:"id"`
}

type UpdateRequest struct {
	Id     string            `json:"id"`
	Group  string            `json:"g"`
	Status int               `json:"s"`
	Param  map[string]string `json:"p"`
	Error  *string           `json:"e"`
}
