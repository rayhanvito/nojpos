package httpx

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/gin-gonic/gin"
)

func listTestContext(query string) *gin.Context {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodGet, "/items"+query, nil)
	return c
}

func TestPaginateListWithoutQueryReturnsAllAndCanonicalKeys(t *testing.T) {
	page := PaginateList(listTestContext(""), []int{1, 2, 3})
	if !reflect.DeepEqual(page.Items, []int{1, 2, 3}) || page.Total != 3 || page.Page != 1 || page.Limit != 3 {
		t.Fatalf("tanpa query harus mengembalikan semua baris: %+v", page)
	}
	body, err := json.Marshal(page)
	if err != nil {
		t.Fatal(err)
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(body, &fields); err != nil {
		t.Fatal(err)
	}
	if len(fields) != 4 {
		t.Fatalf("respons harus punya tepat empat kunci, dapat %v", fields)
	}
	for _, key := range []string{"items", "total", "page", "limit"} {
		if _, ok := fields[key]; !ok {
			t.Fatalf("kunci %q tidak ada dalam respons %v", key, fields)
		}
	}
	if empty := PaginateList(listTestContext(""), []int(nil)).Items; empty == nil {
		t.Fatal("items harus berupa slice kosong saat sumber nil")
	}
}

func TestPaginateListSlicesRequestedPage(t *testing.T) {
	page := PaginateList(listTestContext("?page=2&limit=2"), []int{1, 2, 3, 4, 5})
	if !reflect.DeepEqual(page.Items, []int{3, 4}) || page.Total != 5 || page.Page != 2 || page.Limit != 2 {
		t.Fatalf("page kedua harus berisi item 3 dan 4 dengan total 5: %+v", page)
	}
}

func TestPaginateListPastEndIsEmptyArray(t *testing.T) {
	page := PaginateList(listTestContext("?page=4&limit=2"), []int{1, 2, 3, 4, 5})
	if page.Items == nil || len(page.Items) != 0 {
		t.Fatalf("page melewati ujung harus berisi slice kosong, dapat %#v", page.Items)
	}
	body, err := json.Marshal(page)
	if err != nil {
		t.Fatal(err)
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(body, &fields); err != nil {
		t.Fatal(err)
	}
	if string(fields["items"]) != "[]" {
		t.Fatalf("items kosong harus JSON [] bukan null, dapat %s", fields["items"])
	}
}

func TestPaginateListInvalidLimitUsesDefault(t *testing.T) {
	page := PaginateList(listTestContext("?limit=abc"), make([]int, 25))
	if page.Page != 1 || page.Limit != 20 || len(page.Items) != 20 || page.Total != 25 {
		t.Fatalf("limit tidak valid harus memakai 20, dapat %+v", page)
	}
}

func TestPaginateListInvalidPageAndLargeLimitAreClamped(t *testing.T) {
	page := PaginateList(listTestContext("?page=0&limit=500"), make([]int, 205))
	if page.Page != 1 || page.Limit != 200 || len(page.Items) != 200 || page.Total != 205 {
		t.Fatalf("page tidak valid harus 1 dan limit dibatasi 200, dapat page=%+v", page)
	}
}
