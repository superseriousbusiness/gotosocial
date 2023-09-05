package paging

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
)

func ParseIDPage(c *gin.Context, defaultlimit int) (*Page[string], gtserror.WithCode) {
	// Extract request query params.
	sinceID := c.Query("since_id")
	minID := c.Query("min_id")
	maxID := c.Query("max_id")

	// Extract request limit parameter.
	limit, errWithCode := ParseLimit(c, defaultlimit)
	if errWithCode != nil {
		return nil, errWithCode
	}

	if sinceID == "" &&
		minID == "" &&
		maxID == "" &&
		limit == 0 {
		// No ID paging params provided, and no default
		// limit value which indicates paging not enforced.
		return nil, nil
	}

	return &Page[string]{
		Min:   MinID(minID, sinceID),
		Max:   MaxID(maxID),
		Limit: limit,
	}, nil
}

func ParseShortnameDomainPage(c *gin.Context, defaultlimit int) (*Page[string], gtserror.WithCode) {
	// Extract request query params.
	min := c.Query("min_shortcode_domain")
	max := c.Query("max_shortcode_domain")

	// Extract request limit parameter.
	limit, errWithCode := ParseLimit(c, defaultlimit)
	if errWithCode != nil {
		return nil, errWithCode
	}

	if min == "" &&
		max == "" &&
		limit == 0 {
		// No ID paging params provided, and no default
		// limit value which indicates paging not enforced.
		return nil, nil
	}

	return &Page[string]{
		Min:   MinShortcodeDomain(min),
		Max:   MaxShortcodeDomain(max),
		Limit: limit,
	}, nil
}

func ParseLimit(c *gin.Context, _default int) (int, gtserror.WithCode) {
	const min, max = 2, 100

	// Get limit query param.
	str := c.Query("limit")

	// Attempt to parse limit int.
	i, err := strconv.Atoi(str)
	if err != nil {
		const help = "bad integer limit value"
		return 0, gtserror.NewErrorBadRequest(err, help)
	}

	switch {
	case i == 0:
		return _default, nil
	case i < min:
		return min, nil
	case i > max:
		return max, nil
	default:
		return i, nil
	}
}
