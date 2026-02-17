package dto

type ListProductsRequest struct {
	Page  int    `form:"page"`
	Limit int    `form:"limit"`
	Order string `form:"order"`
}
