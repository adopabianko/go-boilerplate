package dto

type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type PaginationQuery struct {
	Page  int    `form:"page"`
	Limit int    `form:"limit"`
	Order string `form:"order"`
}
