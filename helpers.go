package main

import (
	"strconv"
	"strings"
)

func destructStringToNumbers(str string, sep string) (int, int) {
	clearString := strings.Split(str, sep)

	int1, _ := strconv.Atoi(clearString[0])
	int2, _ := strconv.Atoi(clearString[1])

	return int1, int2
}
