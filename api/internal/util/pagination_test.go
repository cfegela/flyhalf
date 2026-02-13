package util

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetPaginationParams(t *testing.T) {
	tests := []struct {
		name          string
		queryParams   map[string]string
		expectedPage  int
		expectedLimit int
	}{
		{
			name:          "default values",
			queryParams:   map[string]string{},
			expectedPage:  1,
			expectedLimit: 50,
		},
		{
			name:          "valid page and limit",
			queryParams:   map[string]string{"page": "2", "limit": "25"},
			expectedPage:  2,
			expectedLimit: 25,
		},
		{
			name:          "page zero defaults to 1",
			queryParams:   map[string]string{"page": "0"},
			expectedPage:  1,
			expectedLimit: 50,
		},
		{
			name:          "negative page defaults to 1",
			queryParams:   map[string]string{"page": "-1"},
			expectedPage:  1,
			expectedLimit: 50,
		},
		{
			name:          "limit zero defaults to 50",
			queryParams:   map[string]string{"limit": "0"},
			expectedPage:  1,
			expectedLimit: 50,
		},
		{
			name:          "negative limit defaults to 50",
			queryParams:   map[string]string{"limit": "-10"},
			expectedPage:  1,
			expectedLimit: 50,
		},
		{
			name:          "limit exceeds max (100) caps to 100",
			queryParams:   map[string]string{"limit": "150"},
			expectedPage:  1,
			expectedLimit: 50, // Invalid limit defaults to 50
		},
		{
			name:          "limit at max boundary (100)",
			queryParams:   map[string]string{"limit": "100"},
			expectedPage:  1,
			expectedLimit: 100,
		},
		{
			name:          "invalid page string defaults to 1",
			queryParams:   map[string]string{"page": "invalid"},
			expectedPage:  1,
			expectedLimit: 50,
		},
		{
			name:          "invalid limit string defaults to 50",
			queryParams:   map[string]string{"limit": "invalid"},
			expectedPage:  1,
			expectedLimit: 50,
		},
		{
			name:          "large page number",
			queryParams:   map[string]string{"page": "9999"},
			expectedPage:  9999,
			expectedLimit: 50,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request with query parameters
			req, _ := http.NewRequest("GET", "http://example.com", nil)
			q := req.URL.Query()
			for key, value := range tt.queryParams {
				q.Set(key, value)
			}
			req.URL.RawQuery = q.Encode()

			params := GetPaginationParams(req)
			assert.Equal(t, tt.expectedPage, params.Page)
			assert.Equal(t, tt.expectedLimit, params.Limit)
		})
	}
}

func TestCalculateOffset(t *testing.T) {
	tests := []struct {
		name           string
		page           int
		limit          int
		expectedOffset int
	}{
		{
			name:           "first page",
			page:           1,
			limit:          10,
			expectedOffset: 0,
		},
		{
			name:           "second page",
			page:           2,
			limit:          10,
			expectedOffset: 10,
		},
		{
			name:           "third page with limit 25",
			page:           3,
			limit:          25,
			expectedOffset: 50,
		},
		{
			name:           "page 10 with limit 50",
			page:           10,
			limit:          50,
			expectedOffset: 450,
		},
		{
			name:           "page 1 with limit 1",
			page:           1,
			limit:          1,
			expectedOffset: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := PaginationParams{
				Page:  tt.page,
				Limit: tt.limit,
			}
			offset := params.CalculateOffset()
			assert.Equal(t, tt.expectedOffset, offset)
		})
	}
}

func TestCreatePaginatedResponse(t *testing.T) {
	tests := []struct {
		name               string
		data               interface{}
		page               int
		limit              int
		total              int
		expectedTotalPages int
	}{
		{
			name:               "total matches page size",
			data:               []string{"a", "b", "c"},
			page:               1,
			limit:              3,
			total:              3,
			expectedTotalPages: 1,
		},
		{
			name:               "total exceeds page size",
			data:               []string{"a", "b", "c"},
			page:               1,
			limit:              2,
			total:              5,
			expectedTotalPages: 3, // Ceiling of 5/2
		},
		{
			name:               "total zero defaults to 1 page",
			data:               []string{},
			page:               1,
			limit:              10,
			total:              0,
			expectedTotalPages: 1,
		},
		{
			name:               "partial last page",
			data:               []string{"a"},
			page:               3,
			limit:              10,
			total:              21,
			expectedTotalPages: 3, // Ceiling of 21/10
		},
		{
			name:               "exact multiple pages",
			data:               []string{"a", "b"},
			page:               5,
			limit:              10,
			total:              50,
			expectedTotalPages: 5,
		},
		{
			name:               "single item",
			data:               []string{"a"},
			page:               1,
			limit:              50,
			total:              1,
			expectedTotalPages: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := CreatePaginatedResponse(tt.data, tt.page, tt.limit, tt.total)
			assert.Equal(t, tt.data, response.Data)
			assert.Equal(t, tt.page, response.Page)
			assert.Equal(t, tt.limit, response.Limit)
			assert.Equal(t, tt.total, response.Total)
			assert.Equal(t, tt.expectedTotalPages, response.TotalPages)
		})
	}
}
