package payment

import (
	"os"
	"testing"

	"github.com/cucumber/godog"
)

func TestBDD(t *testing.T) {
	opts := &godog.Options{
		Format: "pretty",
		Strict: true,
		// run all *.feature under this directory
		Paths: []string{"."},
		// skip repo-dependent scenarios (e.g., @repo)
		Tags: "~@repo",
	}
	status := godog.TestSuite{
		Name:                "payment-bdd",
		ScenarioInitializer: InitializeScenario,
		Options:             opts,
	}.Run()

	if status != 0 {
		t.FailNow()
	}
}

// allow "go test -v ./..." to pass custom options
func init() {
	godog.BindCommandLineFlags("godog.", &godog.Options{})
	_ = os.Setenv("GODOG_NO_COLOR", "1")
}
