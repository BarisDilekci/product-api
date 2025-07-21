package domain

type Product struct {
	Id          int64     `json:"id"`
	Name        string    `json:"name"`
	Price       float32   `json:"price"`
	Description string    `json:"description"`
	Discount    float32   `json:"discount"`
	Store       string    `json:"store"`
	ImageUrls   []string  `json:"image_urls"`
	CategoryID  int64     `json:"category_id"`
	UserID      int64     `json:"user_id"`
}
