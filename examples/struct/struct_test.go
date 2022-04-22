package main

import (
	"testing"

	"github.com/tetratelabs/wazero/internal/testing/maintester"
	"github.com/tetratelabs/wazero/internal/testing/require"
)

// Test_main ensures the following will work:
//
//	go run struct.go
func Test_main(t *testing.T) {
	stdout, _ := maintester.TestMain(t, main)
	require.Equal(t, `name:Hello Yogi
greeting:Yabba Dabba Doo!
`, stdout)
}
