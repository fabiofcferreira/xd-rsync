package database

import (
	"fmt"

	xd_rsync "github.com/fabiofcferreira/xd-rsync"
)

func (s *Service) GetProductByReferece(id string) (*xd_rsync.XdProduct, error) {
	product := &xd_rsync.XdProduct{}
	s.logger.Info("init_get_product_by_reference", "Fetching product by ID", &map[string]interface{}{
		"reference": id,
	})

	query := BuildSelectQuery(product.GetKnownFieldsSelectors(), "items", []string{
		"KeyId = ?",
	})

	err := s.db.Get(product, query, id)
	if err != nil {
		return nil, fmt.Errorf("could not get product: %w", err)
	}

	s.logger.Info("finished_get_product_by_reference", "Fetching product by ID", &map[string]interface{}{
		"reference": id,
	})
	return product, nil
}
