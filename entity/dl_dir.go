package entity

type DlDir struct {
	Username string `json:"username" validate:"nonzero,nonnil"`
	Dir      string `json:"dir" validate:"nonzero,nonnil"`
}
