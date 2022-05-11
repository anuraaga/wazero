package main

import (
	"github.com/tetratelabs/wazero/internal/testing/maintester"
	"github.com/tetratelabs/wazero/internal/testing/require"
	"testing"
)

// Test_main ensures the following will work:
//
//	go run temperature.go 92
func Test_main(t *testing.T) {
	stdout, _ := maintester.TestMain(t, main, "temperature", "92")
	require.Equal(t, `92F is 33C
`, stdout)
}
