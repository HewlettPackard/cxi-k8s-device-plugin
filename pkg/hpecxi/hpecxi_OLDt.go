package hpecxi

import (
	"testing"
)

func hasHPECXI(t *testing.T) bool {
	devices := GetHPECXIs()
	if len(devices) <= 0 {
		return false
	}
	return true
}
