package xd_rsync

import (
	"reflect"
	"time"
)

type XdProduct struct {
	KeyId        string     `db:"KeyId"`
	Description  string     `db:"Description"`
	ShortName1   string     `db:"ShortName1"`
	RetailPrice1 float64    `db:"RetailPrice1"`
	RetailPrice2 float64    `db:"RetailPrice2"`
	RetailPrice3 float64    `db:"RetailPrice3"`
	CurrentStock float64    `db:"CurrentStock"`
	SyncStamp    *time.Time `db:"SyncStamp"`
}

func (p *XdProduct) GetKnownFields() []string {
	columnNames := []string{}
	val := reflect.ValueOf(p).Elem()
	for i := 0; i < val.NumField(); i++ {
		tag := val.Type().Field(i).Tag
		columnNames = append(columnNames, tag.Get("db"))
	}

	return columnNames
}

func (p *XdProduct) GetKnownFieldsQuerySelectors() string {
	columnNames := p.GetKnownFields()

	var expression string = ""
	for index, name := range columnNames {
		expression += name

		if index < len(columnNames)-1 {
			expression += ", "
		}
	}

	return expression
}

type XdProducts []XdProduct

func (ps *XdProducts) GetKnownFieldsQuerySelectors() string {
	product := &XdProduct{}
	return product.GetKnownFieldsQuerySelectors()
}
