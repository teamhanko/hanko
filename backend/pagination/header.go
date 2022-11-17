package pagination

import (
	"fmt"
	"net/url"
	"strings"
)

func CreateHeader(u *url.URL, total int, page int, itemsPerPage int) string {
	var lastPage int

	if total%itemsPerPage == 0 {
		lastPage = total / itemsPerPage
	} else {
		lastPage = total/itemsPerPage + 1
	}

	if lastPage == 0 {
		lastPage = 1
	}

	var links []string
	if page != 1 || page == lastPage {
		links = append(links, formatter(u, "first", itemsPerPage, 1))
	}

	if page != lastPage {
		links = append(links, formatter(u, "last", itemsPerPage, lastPage))
	}

	if page < lastPage {
		links = append(links, formatter(u, "next", itemsPerPage, page+1))
	}

	if page > 1 {
		links = append(links, formatter(u, "prev", itemsPerPage, page-1))
	}

	return strings.Join(links, ",")
}

func formatter(u *url.URL, rel string, perPage int, page int) string {
	q := u.Query()
	q.Set("page", fmt.Sprintf("%d", page))
	q.Set("per_page", fmt.Sprintf("%d", perPage))
	u.RawQuery = q.Encode()
	return fmt.Sprintf("<%s>; rel=\"%s\"", u.String(), rel)
}
