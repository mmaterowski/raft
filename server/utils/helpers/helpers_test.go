package helpers

import (
	"strings"
	"testing"
)

func TestTrimSuffix(t *testing.T) {
	stringWithSuffix := "string,"
	stringWithSuffix = TrimSuffix(stringWithSuffix, ",")
	if strings.Contains(stringWithSuffix, ",") {
		t.Errorf("String contains a suffix, got: %s", stringWithSuffix)
	}
}
