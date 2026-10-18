package model

type CreateProductRequest struct {
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	CategoryID  *int64                 `json:"category_id"`
	Brand       string                 `json:"brand"`
	ImageURL    string                 `json:"image_url"`
	Variants    []CreateVariantRequest `json:"variants"`
}

type CreateVariantRequest struct {
	SKU        string     `json:"sku"`
	Price      int64      `json:"price"`
	Currency   string     `json:"currency"`
	Stock      int        `json:"stock"`
	Attributes Attributes `json:"attributes"`
}

type UpdateProductRequest struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
	CategoryID  *int64  `json:"category_id"`
	Brand       *string `json:"brand"`
	ImageURL    *string `json:"image_url"`
}

type CreateCategoryRequest struct {
	ParentID *int64 `json:"parent_id"`
	Name     string `json:"name"`
}

type ProductSort string

const (
	SortNewest    ProductSort = "newest"
	SortPriceAsc  ProductSort = "price_asc"
	SortPriceDesc ProductSort = "price_desc"
	SortRating    ProductSort = "rating"
	SortRelevance ProductSort = "relevance"
)

type ProductFilter struct {
	CategoryID   int64
	CategorySlug string
	MinPrice     int64
	MaxPrice     int64
	Search       string
	SellerID     int64
	Sort         ProductSort
	Page         int
	PageSize     int
	RankedIDs    []int64
}

type ProductListResponse struct {
	Products []*Product `json:"products"`
	Total    int        `json:"total"`
	Page     int        `json:"page"`
	PageSize int        `json:"page_size"`
}

type SearchDocument struct {
	ID    int64  `json:"id"`
	Title string `json:"title"`
	Body  string `json:"body"`
}
