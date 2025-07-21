package domain

type Product struct {
	Id          int64
	Name        string
	Price       float32
	Description string
	Discount    float32
	Store       string
	ImageUrls   []string
}
