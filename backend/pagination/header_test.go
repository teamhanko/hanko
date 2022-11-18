package pagination

import (
	"github.com/stretchr/testify/assert"
	"net/url"
	"testing"
)

func TestCreateHeader_FirstPage(t *testing.T) {
	u, _ := url.Parse("http://localhost:8080")
	header := CreateHeader(u, 95, 1, 10)
	assert.Equal(t, "<http://localhost:8080?page=10&per_page=10>; rel=\"last\",<http://localhost:8080?page=2&per_page=10>; rel=\"next\"", header)
}

func TestCreateHeader_LastPage(t *testing.T) {
	u, _ := url.Parse("http://localhost:8080")
	header := CreateHeader(u, 95, 10, 10)
	assert.Equal(t, "<http://localhost:8080?page=1&per_page=10>; rel=\"first\",<http://localhost:8080?page=9&per_page=10>; rel=\"prev\"", header)
}

func TestCreateHeader_MiddlePage(t *testing.T) {
	u, _ := url.Parse("http://localhost:8080")
	header := CreateHeader(u, 95, 4, 10)
	assert.Equal(t, "<http://localhost:8080?page=1&per_page=10>; rel=\"first\",<http://localhost:8080?page=10&per_page=10>; rel=\"last\",<http://localhost:8080?page=5&per_page=10>; rel=\"next\",<http://localhost:8080?page=3&per_page=10>; rel=\"prev\"", header)
}

func TestCreateHeader_TotalCountZero(t *testing.T) {
	u, _ := url.Parse("http://localhost:8080")
	header := CreateHeader(u, 0, 1, 10)
	assert.Equal(t, "<http://localhost:8080?page=1&per_page=10>; rel=\"first\"", header)
}

func TestCreateHeader_ItemsPerPageGreaterTotalCount(t *testing.T) {
	u, _ := url.Parse("http://localhost:8080")
	header := CreateHeader(u, 10, 1, 20)
	assert.Equal(t, "<http://localhost:8080?page=1&per_page=20>; rel=\"first\"", header)
}

func TestCreateHeader_UrlWithQuery(t *testing.T) {
	u, _ := url.Parse("http://localhost:8080?start_time=2022-09-12T12:48:48Z&end_time=2022-09-12T14:48:48Z")
	header := CreateHeader(u, 95, 1, 10)
	assert.Equal(t, "<http://localhost:8080?end_time=2022-09-12T14%3A48%3A48Z&page=10&per_page=10&start_time=2022-09-12T12%3A48%3A48Z>; rel=\"last\",<http://localhost:8080?end_time=2022-09-12T14%3A48%3A48Z&page=2&per_page=10&start_time=2022-09-12T12%3A48%3A48Z>; rel=\"next\"", header)
}
