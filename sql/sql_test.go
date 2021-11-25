package sql

import "testing"

func TestInsert(t *testing.T) {
	if err := Insert(Product{
		Word:       "iphone",
		ProductID:  "123456",
		Name:       "iphone13",
		Price:      3000,
		ImageURL:   "http://123",
		ProductURL: "http://456",
	}); err != nil {
		t.Error("Fail to insert data.")
	} else {
		t.Log("Success to insert data")
	}
}

func TestSelect(t *testing.T) {
	products, err := Select("iphone")
	if err != nil {
		t.Error("Fail to select data.")
	}
	if products[0].Name != "iphone13" {
		t.Error("Fail to select data.")
	} else if products[0].Price != 3000 {
		t.Error("Fail to select data.")
	} else if products[0].ImageURL != "http://123" {
		t.Error("Fail to select data.")
	} else if products[0].ProductURL != "http://456" {
		t.Error("Fail to select data.")
	} else {
		t.Log("Success to select data")
	}
}
