package worker

import (
	"testing"
)

func TestFindMaxPchomePage(t *testing.T) {
	keyword := "iphone13"
	page := FindMaxPchomePage(keyword)
	t.Log("page: ", page)
	if page != 100 {
		t.Error("fail")
	} else {
		t.Log("success")
	}
}
