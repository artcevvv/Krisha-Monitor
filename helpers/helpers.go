package helpers

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

func DestructStringToNumbers(str string, sep string) (int, int, error) {
	if str == "" {
		return 0, 0, errors.New("input string is empty")
	}

	if sep == "" {
		return 0, 0, errors.New("separator cannot be empty")
	}

	parts := strings.Split(str, sep)

	if len(parts) < 2 {
		return 0, 0, fmt.Errorf("string '%s' does not contain at least two values separated by '%s'", str, sep)
	}

	int1, err1 := strconv.Atoi(strings.TrimSpace(parts[0]))
	if err1 != nil {
		return 0, 0, fmt.Errorf("failed to convert first value '%s' to integer: %w", parts[0], err1)
	}

	int2, err2 := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err2 != nil {
		return 0, 0, fmt.Errorf("failed to convert second value '%s' to integer: %w", parts[1], err2)
	}

	return int1, int2, nil
}

func parsePrice(priceStr string) (float64, error) {
	priceStr = strings.ReplaceAll(priceStr, " ", "")
	priceStr = strings.ReplaceAll(priceStr, "тыс.", "000") // Пример для "100 тыс." -> "100000"
	priceStr = strings.ReplaceAll(priceStr, "₸", "")       // Убираем символ тенге, если он есть

	price, err := strconv.ParseFloat(priceStr, 64)
	if err != nil {
		return 0, fmt.Errorf("не удалось распарсить цену: %v", err)
	}
	return price, nil
}
