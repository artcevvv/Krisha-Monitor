package main

import (
	"errors"
	"strconv"
	"strings"
	"testing"
)

func TestDestructStringToNumbers(t *testing.T) {
	tests := []struct {
		name           string
		inputStr       string
		separator      string
		expectedInt1   int
		expectedInt2   int
		expectedErrMsg string
	}{
		{
			name:         "Valid input with comma separator",
			inputStr:     "42,24",
			separator:    ",",
			expectedInt1: 42,
			expectedInt2: 24,
		},
		{
			name:         "Valid input with dash separator",
			inputStr:     "100-200",
			separator:    "-",
			expectedInt1: 100,
			expectedInt2: 200,
		},
		{
			name:         "Valid input with spaces",
			inputStr:     " 5 , 10 ",
			separator:    ",",
			expectedInt1: 5,
			expectedInt2: 10,
		},
		{
			name:         "Valid input with additional parts (ignore extras)",
			inputStr:     "1,2,3,4",
			separator:    ",",
			expectedInt1: 1,
			expectedInt2: 2,
		},
		{
			name:           "Empty input string",
			inputStr:       "",
			separator:      ",",
			expectedErrMsg: "input string is empty",
		},
		{
			name:           "Empty separator",
			inputStr:       "42,24",
			separator:      "",
			expectedErrMsg: "separator cannot be empty",
		},
		{
			name:           "Not enough parts",
			inputStr:       "42",
			separator:      ",",
			expectedErrMsg: "does not contain at least two values",
		},
		{
			name:           "First part not an integer",
			inputStr:       "abc,24",
			separator:      ",",
			expectedErrMsg: "failed to convert first value",
		},
		{
			name:           "Second part not an integer",
			inputStr:       "42,xyz",
			separator:      ",",
			expectedErrMsg: "failed to convert second value",
		},
		{
			name:           "Both parts not integers",
			inputStr:       "abc,xyz",
			separator:      ",",
			expectedErrMsg: "failed to convert first value",
		},
		{
			name:         "Negative numbers",
			inputStr:     "-10,-20",
			separator:    ",",
			expectedInt1: -10,
			expectedInt2: -20,
		},
		{
			name:         "Zero values",
			inputStr:     "0,0",
			separator:    ",",
			expectedInt1: 0,
			expectedInt2: 0,
		},
	}

	// Run test cases
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			int1, int2, err := DestructStringToNumbers(tc.inputStr, tc.separator)

			// Check error cases
			if tc.expectedErrMsg != "" {
				if err == nil {
					t.Fatalf("Expected error containing '%s', but got nil", tc.expectedErrMsg)
				}
				if !strings.Contains(err.Error(), tc.expectedErrMsg) {
					t.Fatalf("Expected error containing '%s', but got '%s'", tc.expectedErrMsg, err.Error())
				}
			} else {
				// Check success cases
				if err != nil {
					t.Fatalf("Expected no error, but got: %v", err)
				}
				if int1 != tc.expectedInt1 {
					t.Errorf("Expected first integer to be %d, but got %d", tc.expectedInt1, int1)
				}
				if int2 != tc.expectedInt2 {
					t.Errorf("Expected second integer to be %d, but got %d", tc.expectedInt2, int2)
				}
			}
		})
	}
}

// Test that errors are properly wrapped
func TestErrorWrapping(t *testing.T) {
	// Test case where first value is not an integer
	_, _, err := DestructStringToNumbers("abc,123", ",")
	var numError *strconv.NumError
	if !errors.As(err, &numError) {
		t.Errorf("Expected error to wrap a strconv.NumError, but it didn't")
	}

	// Test case where second value is not an integer
	_, _, err = DestructStringToNumbers("123,abc", ",")
	if !errors.As(err, &numError) {
		t.Errorf("Expected error to wrap a strconv.NumError, but it didn't")
	}
}
