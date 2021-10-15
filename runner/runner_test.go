package runner

import (
	"blackJack/config"
	"testing"
)

func TestRunner(t *testing.T) {
	options := config.DefaultOption
	r, _ := New(&options)
	r.CreateRunner()
}
