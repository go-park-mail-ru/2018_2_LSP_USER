package handlers

type updateUserPayload struct {
	FirstName   string `json:"firstname"`
	LastName    string `json:"lastname"`
	Password    string `json:"password"`
	OldPassword string `json:"oldpassword"`
}

type getOneUserPayload struct {
	Fields string `json:"fields"`
}

type getManyUsersPayload struct {
	Page    int
	Fields  string
	OrderBy string
}
