package lcov

import (
	"fmt"
	"strconv"
	"strings"
)

func parseCount(line, prefix string) (int, error) {
	count, err := strconv.Atoi(strings.TrimPrefix(line, prefix))
	if err != nil {
		return 0, fmt.Errorf("%w: %s: %w", ErrInvalidLine, prefix, err)
	}
	return count, nil
}
