package contract

type ErrorResponse struct {
	Error string `json:"error"`
}

type AddRequest struct {
	Group string            `json:"g"`
	Kind  string            `json:"k"`
	Owner *string           `json:"o"`
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

type OwnerUnRegRequest struct {
	Owner    string `json:"o"`
	Internal bool   `json:"i"`
}

type GetFirstInGroupResponse struct {
	Id string `json:"id"`
}

type SearchTaskRequest struct {
	Condition *Condition `json:"c"`
	Kind      *string    `json:"k"`
	Size      *uint      `json:"s"`
	Internal  bool       `json:"i"`
}

type SearchUpdateTaskRequest struct {
	Up        TaskUpdate
	Condition *Condition `json:"c"`
	Kind      *string    `json:"k"`
	Size      *uint      `json:"s"`
	Internal  bool       `json:"i"`
}
