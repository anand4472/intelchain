package genesis

// IntelchainAccounts are the accounts for the initial genesis nodes hosted by Intelchain.
var IntelchainAccounts = []DeployAccount{
	{Index: " 0 ", Address: "", BLSPublicKey: ""},
	{Index: " 1 ", Address: "", BLSPublicKey: ""},
}

// IntelchainAccountsPostHIP30 are the accounts after shard and node count reduction.
// It is calculated by removing shard 2 / 3 keys from IntelchainAccounts.
// There is no need to remove 10% (40) keys from the bottom because they will simply be unelected
var IntelchainAccountsPostHIP30 = []DeployAccount{
	{Index: " 0 ", Address: "", BLSPublicKey: ""},
	{Index: " 1 ", Address: "", BLSPublicKey: ""},
}
