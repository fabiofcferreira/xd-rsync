package database

import (
	"fmt"
	"math"
	"sync"

	xd_rsync "github.com/fabiofcferreira/xd-rsync"
	"github.com/jmoiron/sqlx"
)

func (s *Service) GetProductByReferece(id string) (*xd_rsync.XdProduct, error) {
	product := &xd_rsync.XdProduct{}
	s.logger.Info("init_get_product_by_reference", "Fetching product by ID", &map[string]interface{}{
		"reference": id,
	})

	query := BuildSelectQuery(product.GetKnownColumnsQuerySelectors(), product.GetTableName(), []string{
		"KeyId = ?",
	})

	err := s.db.Get(product, query, id)
	if err != nil {
		s.logger.Error("failed_get_product_by_reference", "Failed fetching product by ID", &map[string]interface{}{
			"reference": id,
		})
		return nil, fmt.Errorf("could not get product: %w", err)
	}

	s.logger.Info("finished_get_product_by_reference", "Fetched product by ID", &map[string]interface{}{
		"reference": id,
	})
	return product, nil
}

func (s *Service) GetProductsByReferece(ids []string) (*xd_rsync.XdProducts, error) {
	products := &xd_rsync.XdProducts{}
	s.logger.Info("init_get_products_by_reference", "Fetching products by ID", &map[string]interface{}{
		"references": ids,
	})

	query := BuildSelectQuery(products.GetKnownColumnsQuerySelectors(), products.GetTableName(), []string{
		"KeyId IN (?)",
	})
	processedQuery, args, err := sqlx.In(query, ids)
	if err != nil {
		s.logger.Error("failed_get_products_by_reference_query_build", "Failed to build query to fetch products by ID", &map[string]interface{}{
			"error": err.Error(),
		})
	}

	bindedQuery := s.db.Rebind(processedQuery)

	err = s.db.Select(products, bindedQuery, args...)
	if err != nil {
		s.logger.Error("failed_get_products_by_reference", "Failed fetching products by ID", &map[string]interface{}{
			"error": err.Error(),
		})
		return nil, fmt.Errorf("could not get products: %w", err)
	}

	s.logger.Info("finished_get_products_by_reference", "Fetched products by ID", &map[string]interface{}{
		"reference": ids,
	})
	return products, nil
}

func (s *Service) GetPricedProductsCount() (int, error) {
	products := &xd_rsync.XdProducts{}
	s.logger.Info("init_count_priced_products", "Fetching count of allpriced products", nil)

	countQuery := BuildSelectQuery(BuildCountExpression(products.GetPrimaryKeyColumnName()), products.GetTableName(), []string{
		"RetailPrice1 > 0",
		"RetailPrice2 > 0",
		"RetailPrice3 > 0",
	})
	pricedProductsCount := 0
	err := s.db.Get(&pricedProductsCount, countQuery)
	if err != nil {
		s.logger.Error("failed_get_count_all_priced_products", "Failed fetching count of all priced product", nil)
		return -1, err
	}

	return pricedProductsCount, nil
}

func (s *Service) GetPaginatedPricedProducts(limit int, offset int) (*xd_rsync.XdProducts, error) {
	products := &xd_rsync.XdProducts{}

	s.logger.Info("init_get_paginated_priced_products", "Fetching paginated priced products", &map[string]interface{}{
		"limit":  limit,
		"offset": offset,
	})

	query := BuildSelectQueryWithEndClauses(products.GetKnownColumnsQuerySelectors(), products.GetTableName(), []string{
		"RetailPrice1 > 0",
		"RetailPrice2 > 0",
		"RetailPrice3 > 0",
	},
		[]string{
			BuildLimitOffsetExpression(limit, offset),
		})

	err := s.db.Select(products, query)
	if err != nil {
		s.logger.Info("init_get_paginated_priced_products", "Failed fetching paginated priced products", &map[string]interface{}{
			"limit":  limit,
			"offset": offset,
			"error":  err,
		})
	}

	s.logger.Info("finished_get_paginated_priced_products", "Fetched paginated priced products", &map[string]interface{}{
		"limit":  limit,
		"offset": offset,
	})
	return products, nil
}

func (s *Service) GetPricedProducts() (*xd_rsync.XdProducts, error) {
	products := &xd_rsync.XdProducts{}
	pricedProductsCount, _ := s.GetPricedProductsCount()

	s.logger.Info("init_get_all_priced_products", "Fetching all priced products", &map[string]interface{}{
		"productsCount": pricedProductsCount,
	})

	chunkSize := int(math.Floor(float64(pricedProductsCount) / 200))
	chunkResults := make(map[int]*xd_rsync.XdProducts)
	wg := sync.WaitGroup{}

	for chunkNumber := 0; chunkNumber <= chunkSize; chunkNumber++ {
		wg.Add(1)

		go func() {
			var err error
			chunkResults[chunkNumber], err = s.GetPaginatedPricedProducts(200, chunkNumber*200)
			if err != nil {
				s.logger.Error("failed_get_priced_products_chunk", "Failed fetching priced products chunk", &map[string]interface{}{
					"productsCount": pricedProductsCount,
				})
			}

			wg.Done()
		}()
	}

	wg.Wait()

	for chunkNumber := 0; chunkNumber <= chunkSize; chunkNumber++ {
		*products = append(*products, *chunkResults[chunkNumber]...)
	}

	s.logger.Info("finished_get_all_priced_products", "Fetched all priced products in chunks", &map[string]interface{}{
		"productsCount": pricedProductsCount,
	})
	return products, nil
}
