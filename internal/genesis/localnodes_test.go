package genesis

import "testing"

func TestLocalTestAccounts(t *testing.T) {
	for name, accounts := range map[string][]DeployAccount{
		"IntelchainV0":   LocalIntelchainAccounts,
		"IntelchainV1":   LocalIntelchainAccountsV1,
		"IntelchainV2":   LocalIntelchainAccountsV2,
		"FoundationalV0": LocalFnAccounts,
		"FoundationalV1": LocalFnAccountsV1,
		"FoundationalV2": LocalFnAccountsV2,
	} {
		t.Run(name, func(t *testing.T) { testDeployAccounts(t, accounts) })
	}
}
