package database

import (
	"fmt"
	"sync"
	"time"

	xd_rsync "github.com/fabiofcferreira/xd-rsync"
	"github.com/jmoiron/sqlx"
)

var PRICED_PRODUCT_CONDITION = []string{
	"i.RetailPrice2 > 0",
}

var ITEM_TO_ITEMSTOCK_JOIN_EXPRESSION = "LEFT JOIN xd.itemstock istock ON istock.ItemKeyId = i.KeyId"

func getUpdatedAfterCondition(updatedAfter *time.Time) string {
	return fmt.Sprintf("(i.SyncStamp > '%s' OR istock.SyncStamp > '%s' OR istock.LastEntrance > '%s' OR istock.LastExit > '%s')", formatTimestampToRFC3339(updatedAfter), formatTimestampToRFC3339(updatedAfter))
}

func (s DatabaseClient) GetProductByReferece(id string) (*xd_rsync.XdProduct, error) {
	product := &xd_rsync.XdProduct{}
	s.logger.Info("init_get_product_by_reference", "Fetching product by ID", &map[string]interface{}{
		"reference": id,
	})

	query := joinAllExpressions([]string{
		buildSelectTableExpression(product.GetKnownColumnsQuerySelectors(), product.GetTableName()),
		buildWhereExpression([]string{
			"KeyId = ?",
		}),
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

	query := joinAllExpressions([]string{
		buildSelectTableExpression(products.GetKnownColumnsQuerySelectors(), products.GetTableName()),
		buildWhereExpression([]string{
			"KeyId IN (?)",
		}),
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
		conditions = append(conditions, getUpdatedAfterCondition(updatedAfter))
	}

	query := joinAllExpressions([]string{
		buildSelectTableExpression(buildCountExpression(products.GetPrimaryKeyColumnName()), products.GetTableName()),
		ITEM_TO_ITEMSTOCK_JOIN_EXPRESSION,
		buildWhereExpression(conditions),
	})

	pricedProductsCount := 0
	err := s.db.Get(&pricedProductsCount, query)
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
		conditions = append(conditions, getUpdatedAfterCondition(updatedAfter))
	}

	query := joinAllExpressions([]string{
		buildSelectTableExpression(products.GetKnownColumnsQuerySelectors(), products.GetTableName()),
		ITEM_TO_ITEMSTOCK_JOIN_EXPRESSION,
		buildWhereExpression(conditions),
		buildLimitOffsetExpression(limit, offset),
	})

	fmt.Println("paginated_priced_products_query", query)

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

	numberOfPagesNeeded := GetPagesNeeded(pricedProductsCount, 200)
	productsChunks := xd_rsync.XdProductsChunksWithMutex{
		Chunks: &map[int]xd_rsync.XdProducts{},
	}

	wg := sync.WaitGroup{}
	for pageNumber := 0; pageNumber < numberOfPagesNeeded; pageNumber++ {
		wg.Add(1)

		go func() {
			var err error
			page, err := s.GetPaginatedPricedProducts(nil, 200, pageNumber*200)
			if err != nil {
				s.logger.Error("failed_get_priced_products_chunk", "Failed fetching priced products chunk", &map[string]interface{}{
					"error": err,
				})
			}

			productsChunks.UpdateChunk(pageNumber, page)
			wg.Done()
		}()
	}

	wg.Wait()

	productsChunks.GetList(products)

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

	numberOfPagesNeeded := GetPagesNeeded(pricedProductsCount, 200)
	productsChunks := xd_rsync.XdProductsChunksWithMutex{
		Chunks: &map[int]xd_rsync.XdProducts{},
	}

	wg := sync.WaitGroup{}
	for pageNumber := 0; pageNumber < numberOfPagesNeeded; pageNumber++ {
		wg.Add(1)

		go func() {
			page, err := s.GetPaginatedPricedProducts(ts, 200, pageNumber*200)
			if err != nil {
				s.logger.Error("failed_get_all_priced_products_since_time", "Failed fetching priced products since timestamp", &map[string]interface{}{
					"productsCount":    pricedProductsCount,
					"minimumTimestamp": ts,
				})
			}

			productsChunks.UpdateChunk(pageNumber, page)
			wg.Done()
		}()
	}

	wg.Wait()

	productsChunks.GetList(products)

	s.logger.Info("finished_get_all_priced_products_since_time", "Fetched all priced products since timestamp", &map[string]interface{}{
		"productsCount":    pricedProductsCount,
		"minimumTimestamp": ts,
	})
	return products, nil
}
