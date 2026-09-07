package tests

import (
	"anantashahane/BLADE_db/internal/config"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	// setup — runs before the tests
	state := config.GetContext("test")
	preTestClear(state)
	code := m.Run()
	os.Exit(code)
}
