package model

type ProductCreate struct {
	Name        string
	Price       float32
	Description string
	Discount    float32
	Store       string
	ImageUrls   []string
}
