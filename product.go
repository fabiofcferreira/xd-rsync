package xd_rsync

import (
	"encoding/json"
	"fmt"
	"reflect"
	"time"
)

var ErrProductJsonNotValid = fmt.Errorf("emitted product JSON is not valid")

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

func (p *XdProduct) GetTableName() string {
	return "items"
}

func (p *XdProduct) GetPrimaryKeyColumnName() string {
	val := reflect.ValueOf(p).Elem()

	// Primary key is the first column
	tag := val.Type().Field(0).Tag
	return tag.Get("db")
}

func (p *XdProduct) GetKnownColumns() []string {
	columnNames := []string{}
	val := reflect.ValueOf(p).Elem()
	for i := 0; i < val.NumField(); i++ {
		tag := val.Type().Field(i).Tag
		columnNames = append(columnNames, tag.Get("db"))
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
	return "items"
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
