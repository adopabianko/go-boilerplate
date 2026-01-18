package entity

type Product struct {
	ID          int      `json:"id"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Price       float64  `json:"price"`
	Category    string   `json:"category"`
	Brand       string   `json:"brand"`
	Thumbnail   string   `json:"thumbnail"`
	Images      []string `json:"images"`
}
