package genesis

import "testing"

func TestTNIntelchainAccounts(t *testing.T) {
	testDeployAccounts(t, TNIntelchainAccounts)
}

func TestTNFoundationalAccounts(t *testing.T) {
	testDeployAccounts(t, TNFoundationalAccounts)
}
