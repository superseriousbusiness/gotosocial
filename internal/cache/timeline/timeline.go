package timeline

import (
	"codeberg.org/gruf/go-structr"
	"github.com/superseriousbusiness/gotosocial/internal/paging"
)

func nextPageParams(
	curLo, curHi string, // current page params
	nextLo, nextHi string, // next lo / hi values
	order paging.Order,
) (lo string, hi string) {
	if order.Ascending() {

	} else /* i.e. descending */ {

	}
}

// toDirection converts page order to timeline direction.
func toDirection(o paging.Order) structr.Direction {
	switch o {
	case paging.OrderAscending:
		return structr.Asc
	case paging.OrderDescending:
		return structr.Desc
	default:
		return false
	}
}
