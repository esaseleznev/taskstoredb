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

type OwnerRegRequest struct {
	Owner    string   `json:"o"`
	Kinds    []string `json:"k"`
	Internal bool     `json:"i"`
}

type SetOffsetRequest struct {
	Owner    string `json:"o"`
	Kind     string `json:"k"`
	StartId  string `json:"id"`
	Internal bool   `json:"i"`
}

type GetFirstInGroupResponse struct {
	Id string `json:"id"`
}

type PoolRequest struct {
	Owner    string `json:"o"`
	Kind     string `json:"k"`
	Internal bool   `json:"i"`
}
