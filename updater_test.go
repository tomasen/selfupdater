package selfupdater

import "testing"

func TestProviderImplementation(t *testing.T) {
	var _ UpdateProvider = &S3UpdateProvider{}
}
