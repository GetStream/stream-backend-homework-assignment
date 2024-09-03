package utils

import (
	"fmt"
	"strconv"
)

func GetPageNumber(pageStr string) (int, error) {
	if pageStr == "" {
		return 1, nil
	}
	pageNumber, err := strconv.Atoi(pageStr)
	if err != nil || pageNumber < 1 {
		return 0, fmt.Errorf("invalid page number")
	}
	return pageNumber, nil
}
