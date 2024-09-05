package xd_rsync

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sync"
	"time"
)

var ErrProductJsonNotValid = fmt.Errorf("emitted product JSON is not valid")

type XdProduct struct {
	SKU               string     `db:"KeyId" dbSelector:"i.KeyId" json:"sku"`
	Description       string     `db:"Description" dbSelector:"i.Description" json:"name"`
	RetailPrice1      float64    `db:"RetailPrice1" dbSelector:"i.RetailPrice1" json:"clientCompareAtPrice"`
	RetailPrice2      float64    `db:"RetailPrice2" dbSelector:"i.RetailPrice2" json:"clientPrice"`
	AvailableQuantity float64    `db:"AvailableQuantity" dbSelector:"IFNULL(istock.AvailableQuantity, 0) as AvailableQuantity" json:"availableQuantity"`
	SyncStamp         *time.Time `db:"SyncStamp" dbSelector:"i.SyncStamp as SyncStamp" json:"syncStamp"`
	StockSyncStamp    *time.Time `db:"StockSyncStamp" dbSelector:"istock.SyncStamp as StockSyncStamp" json:"stockSyncStamp"`
	StockLastEntrance *time.Time `db:"StockLastEntrance" dbSelector:"istock.LastEntrance as StockLastEntrance" json:"stockLastEntrance"`
	StockLastExit     *time.Time `db:"StockLastExit" dbSelector:"istock.LastExit as StockLastExit" json:"stockLastExit"`
}

func (p *XdProduct) GetTableName() string {
	return "items"
}

func (p *XdProduct) GetPrimaryKeyColumnName() string {
	val := reflect.ValueOf(p).Elem()

	// Primary key is the first column
	tag := val.Type().Field(0).Tag
	return tag.Get("dbSelector")
}

func (p *XdProduct) GetKnownColumns() []string {
	columnNames := []string{}
	val := reflect.ValueOf(p).Elem()
	for i := 0; i < val.NumField(); i++ {
		tag := val.Type().Field(i).Tag
		columnNames = append(columnNames, tag.Get("dbSelector"))
	}

	return columnNames
}

func (p *XdProduct) GetKnownColumnsQuerySelectors() string {
	columnNames := p.GetKnownColumns()

	var expression string = ""
	for index, name := range columnNames {
		expression += name

		if index < len(columnNames)-1 {
			expression += ", "
		}
	}

	return expression
}

func (p *XdProduct) ToJSON() (string, error) {
	bytes, err := json.Marshal(p)
	if err != nil {
		return "", ErrProductJsonNotValid
	}

	return string(bytes), nil
}

type XdProducts []XdProduct

func (ps *XdProducts) GetTableName() string {
	return "items i"
}

func (ps *XdProducts) GetPrimaryKeyColumnName() string {
	product := &XdProduct{}
	return product.GetPrimaryKeyColumnName()
}

func (ps *XdProducts) GetKnownColumnsQuerySelectors() string {
	product := &XdProduct{}
	return product.GetKnownColumnsQuerySelectors()
}

func (ps *XdProducts) ToJSON() (string, error) {
	bytes, err := json.Marshal(ps)
	if err != nil {
		return "", ErrProductJsonNotValid
	}

	return string(bytes), nil
}

type XdProductsChunksWithMutex struct {
	mutex  sync.Mutex
	Chunks *map[int]XdProducts
}

func (c *XdProductsChunksWithMutex) UpdateChunk(index int, productsList *XdProducts) {
	c.mutex.Lock()

	defer c.mutex.Unlock()
	(*c.Chunks)[index] = *productsList
}

func (c *XdProductsChunksWithMutex) GetList(list *XdProducts) {
	for chunkNumber := 0; chunkNumber < len(*c.Chunks); chunkNumber++ {
		*list = append(*list, (*c.Chunks)[chunkNumber]...)
	}
}
