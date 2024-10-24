package services

import (
	"strconv"
)

func parseID(idString string) int64 {
	id, _ := strconv.ParseInt(idString, 10, 64)
	return id
}
