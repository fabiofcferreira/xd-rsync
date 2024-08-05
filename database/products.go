package database

import (
	"fmt"
	"math"
	"sync"
	"time"

	xd_rsync "github.com/fabiofcferreira/xd-rsync"
	"github.com/jmoiron/sqlx"
)

var PRICED_PRODUCT_CONDITION = []string{
	"RetailPrice1 > 0",
	"RetailPrice2 > 0",
	"RetailPrice3 > 0",
}

func (s DatabaseClient) GetProductByReferece(id string) (*xd_rsync.XdProduct, error) {
	product := &xd_rsync.XdProduct{}
	s.logger.Info("init_get_product_by_reference", "Fetching product by ID", &map[string]interface{}{
		"reference": id,
	})

	query := buildSelectQuery(product.GetKnownColumnsQuerySelectors(), product.GetTableName(), []string{
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

func (s DatabaseClient) GetProductsByReferece(ids []string) (*xd_rsync.XdProducts, error) {
	products := &xd_rsync.XdProducts{}
	s.logger.Info("init_get_products_by_reference", "Fetching products by ID", &map[string]interface{}{
		"references": ids,
	})

	query := buildSelectQuery(products.GetKnownColumnsQuerySelectors(), products.GetTableName(), []string{
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

func (s DatabaseClient) GetPricedProductsCount(updatedAfter *time.Time) (int, error) {
	products := &xd_rsync.XdProducts{}
	s.logger.Info("init_count_priced_products", "Fetching count of all priced products", nil)

	conditions := PRICED_PRODUCT_CONDITION
	if updatedAfter != nil {
		conditions = append(conditions, fmt.Sprintf("SyncStamp > '%s'", formatTimestampToRFC3339(updatedAfter)))
	}

	countQuery := buildSelectQuery(buildCountExpression(products.GetPrimaryKeyColumnName()), products.GetTableName(), conditions)
	pricedProductsCount := 0
	err := s.db.Get(&pricedProductsCount, countQuery)
	if err != nil {
		s.logger.Error("failed_get_count_all_priced_products", "Failed fetching count of all priced product", &map[string]interface{}{
			"error": err,
		})
		return -1, err
	}

	return pricedProductsCount, nil
}

func (s DatabaseClient) GetPaginatedPricedProducts(updatedAfter *time.Time, limit int, offset int) (*xd_rsync.XdProducts, error) {
	products := &xd_rsync.XdProducts{}

	s.logger.Info("init_get_paginated_priced_products", "Fetching paginated priced products", &map[string]interface{}{
		"limit":  limit,
		"offset": offset,
	})

	conditions := PRICED_PRODUCT_CONDITION
	if updatedAfter != nil {
		conditions = append(conditions, fmt.Sprintf("SyncStamp > '%s'", formatTimestampToRFC3339(updatedAfter)))
	}

	query := buildSelectQueryWithEndClauses(products.GetKnownColumnsQuerySelectors(), products.GetTableName(), conditions,
		[]string{
			buildLimitOffsetExpression(limit, offset),
		})

	err := s.db.Select(products, query)
	if err != nil {
		s.logger.Error("failed_get_paginated_priced_products", "Failed fetching paginated priced products", &map[string]interface{}{
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

func (s DatabaseClient) GetPricedProducts() (*xd_rsync.XdProducts, error) {
	products := &xd_rsync.XdProducts{}
	pricedProductsCount, _ := s.GetPricedProductsCount(nil)

	s.logger.Info("init_get_all_priced_products", "Fetching all priced products", &map[string]interface{}{
		"productsCount": pricedProductsCount,
	})

	chunksNeeded := int(math.Ceil(float64(pricedProductsCount) / 200))
	chunkResults := make(map[int]*xd_rsync.XdProducts)
	wg := sync.WaitGroup{}

	for chunkNumber := 0; chunkNumber < chunksNeeded; chunkNumber++ {
		wg.Add(1)

		go func() {
			var err error
			chunkResults[chunkNumber], err = s.GetPaginatedPricedProducts(nil, 200, chunkNumber*200)
			if err != nil {
				s.logger.Error("failed_get_priced_products_chunk", "Failed fetching priced products chunk", &map[string]interface{}{
					"error": err,
				})
			}

			wg.Done()
		}()
	}

	wg.Wait()

	for chunkNumber := 0; chunkNumber < chunksNeeded; chunkNumber++ {
		*products = append(*products, *chunkResults[chunkNumber]...)
	}

	s.logger.Info("finished_get_all_priced_products", "Fetched all priced products in chunks", &map[string]interface{}{
		"productsCount": pricedProductsCount,
	})
	return products, nil
}

func (s DatabaseClient) GetPricedProductsSinceTimestamp(ts *time.Time) (*xd_rsync.XdProducts, error) {
	products := &xd_rsync.XdProducts{}
	pricedProductsCount, _ := s.GetPricedProductsCount(ts)

	s.logger.Info("init_get_all_priced_products_since_time", "Fetching all priced products since timestamp", &map[string]interface{}{
		"productsCount":    pricedProductsCount,
		"minimumTimestamp": ts,
	})

	chunksNeeded := int(math.Ceil(float64(pricedProductsCount) / 200))
	fmt.Println("chunk", chunksNeeded)
	chunkResults := make(map[int]*xd_rsync.XdProducts)
	wg := sync.WaitGroup{}

	for chunkNumber := 0; chunkNumber < chunksNeeded; chunkNumber++ {
		wg.Add(1)

		go func() {
			var err error
			chunkResults[chunkNumber], err = s.GetPaginatedPricedProducts(ts, 200, chunkNumber*200)
			if err != nil {
				s.logger.Error("failed_get_all_priced_products_since_time", "Failed fetching priced products since timestamp", &map[string]interface{}{
					"productsCount":    pricedProductsCount,
					"minimumTimestamp": ts,
				})
			}

			wg.Done()
		}()
	}

	wg.Wait()

	for chunkNumber := 0; chunkNumber < chunksNeeded; chunkNumber++ {
		*products = append(*products, *chunkResults[chunkNumber]...)
	}

	s.logger.Info("finished_get_all_priced_products_since_time", "Fetched all priced products since timestamp", &map[string]interface{}{
		"productsCount":    pricedProductsCount,
		"minimumTimestamp": ts,
	})
	return products, nil
}
