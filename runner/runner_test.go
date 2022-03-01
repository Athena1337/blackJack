package runner

import (
	"github.com/Athena1337/blackJack/config"
	"testing"
)

func TestRunner(t *testing.T) {
	options := config.DefaultOption
	r, _ := New(&options)
	r.CreateRunner()
}
