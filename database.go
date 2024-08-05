package xd_rsync

import "time"

type DatabaseService interface {
	GetProductByReferece(id string) (*XdProduct, error)
	GetProductsByReferece(ids []string) (*XdProducts, error)
	GetPricedProductsCount(ts *time.Time) (int, error)
	GetPaginatedPricedProducts(ts *time.Time, limit int, offset int) (*XdProducts, error)
	GetPricedProducts() (*XdProducts, error)
	GetPricedProductsSinceTimestamp(ts *time.Time) (*XdProducts, error)
}
