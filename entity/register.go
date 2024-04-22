package entity

type Register struct {
	Username  string `json:"username" validate:"nonzero,nonnil"`
	Email     string `json:"email" validate:"nonzero,nonnil"`
	Password  string `json:"password" validate:"nonzero,nonnil"`
	FirstName string `json:"firstName" validate:"nonzero,nonnil"`
	LastName  string `json:"lastName" validate:"nonzero,nonnil"`
	CreatedBy string `json:"createdBy" validate:"nonzero,nonnil"`
	Company   string `json:"company" validate:"nonzero,nonnil"`
}
