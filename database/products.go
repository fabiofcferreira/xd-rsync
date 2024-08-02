package database

import (
	"fmt"

	xd_rsync "github.com/fabiofcferreira/xd-rsync"
	"github.com/jmoiron/sqlx"
)

func (s *Service) GetProductByReferece(id string) (*xd_rsync.XdProduct, error) {
	product := &xd_rsync.XdProduct{}
	s.logger.Info("init_get_product_by_reference", "Fetching product by ID", &map[string]interface{}{
		"reference": id,
	})

	query := BuildSelectQuery(product.GetKnownFieldsQuerySelectors(), "items", []string{
		"KeyId = ?",
	})

	err := s.db.Get(product, query, id)
	if err != nil {
		s.logger.Error("failed_get_product_by_reference", "Failed fetching product by ID", &map[string]interface{}{
			"reference": id,
		})
		return nil, fmt.Errorf("could not get product: %w", err)
	}

	s.logger.Info("finished_get_product_by_reference", "Fetching product by ID", &map[string]interface{}{
		"reference": id,
	})
	return product, nil
}

func (s *Service) GetProductsByReferece(ids []string) (*xd_rsync.XdProducts, error) {
	products := &xd_rsync.XdProducts{}
	s.logger.Info("init_get_products_by_reference", "Fetching products by ID", &map[string]interface{}{
		"references": ids,
	})

	query := BuildSelectQuery(products.GetKnownFieldsQuerySelectors(), "items", []string{
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

	s.logger.Info("finished_get_products_by_reference", "Fetching product by ID", &map[string]interface{}{
		"reference": ids,
	})
	return products, nil
}
