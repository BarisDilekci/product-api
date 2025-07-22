package response

import "product-app/domain"

type ErrorResponse struct {
	ErrorDescription string `json:"errorDescription"`
}

type ProductResponse struct {
	Name        string   `json:"name"`
	Price       float32  `json:"price"`
	Description string   `json:"description"`
	Discount    float32  `json:"discount"`
	Store       string   `json:"store"`
	ImageUrls   []string `json:"image_urls"`
	CategoryID  int64    `json:"category_id"`
}

func ToResponse(product domain.Product) ProductResponse {
	return ProductResponse{
		Name:        product.Name,
		Price:       product.Price,
		Description: product.Description,
		Discount:    product.Discount,
		Store:       product.Store,
		ImageUrls:   product.ImageUrls,
		CategoryID:  product.CategoryID,
	}
}
func ToResponseList(products []domain.Product) []ProductResponse {
	var productResponseList = []ProductResponse{}
	for _, product := range products {
		productResponseList = append(productResponseList, ToResponse(product))
	}
	return productResponseList
}
