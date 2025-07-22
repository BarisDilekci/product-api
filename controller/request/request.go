package request

import "product-app/service/model"

type AddProductRequest struct {
	Name        string   `json:"name"`
	Price       float32  `json:"price"`
	Description string   `json:"description"`
	Discount    float32  `json:"discount"`
	Store       string   `json:"store"`
	ImageUrls   []string `json:"image_urls"`
	CategoryID  int64    `json:"category_id"`
}

func (addProductRequest AddProductRequest) ToModel() model.ProductCreate {
	return model.ProductCreate{
		Name:        addProductRequest.Name,
		Price:       addProductRequest.Price,
		Description: addProductRequest.Description,
		Discount:    addProductRequest.Discount,
		Store:       addProductRequest.Store,
		ImageUrls:   addProductRequest.ImageUrls,
		CategoryID:  addProductRequest.CategoryID,
	}
}
