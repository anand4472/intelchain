package rpc

import (
	"testing"
)

func TestRpcMethodFilter(t *testing.T) {
	method_filters_toml := `
		Allow = [ 
			"itc_method1",
			"wildcard:itcv2_method?",
			"eth*",
			"itc_getNetworkInfo",
			"regex:^itc_send[a-zA-Z]+"
		]

		Deny = [ 
			"*staking*",
			"eth_get*",
			"itc_getNetworkInfo",
			"exact:itc_sendTx"
		]
	`
	b := []byte(method_filters_toml)

	var rmf RpcMethodFilter
	rmf.LoadRpcMethodFilters(b)

	tests := []struct {
		name     string
		exposure bool
	}{
		0:  {"itc_method1", true},         // auto detected exact match which exists (case-insensitive)
		1:  {"itc_MeThoD1", true},         // check case-insensitive
		2:  {"itc_method2", false},        // not exist in allows
		3:  {"itcv2_method5", true},       // wildcard
		4:  {"itcv2_method", false},       // false case for wild card
		5:  {"eth_chainID", true},         // auto detected wild card in allow filters
		6:  {"eth_getValidator", false},   // auto detected wild card in deny filters
		7:  {"itc_getStakingInfo", false}, // deny wild card
		8:  {"abc", false},                // not exist pattern
		9:  {"itc_getNetworkInfo", false}, // case-insensitive normal word match
		10: {"itc_sendTx", false},         // exact match (case-sensitive)
		11: {"itc_sendtx", true},          // exact match (case-sensitive)
	}

	for i, test := range tests {
		mustExpose := rmf.Expose(test.name)

		if mustExpose != test.exposure {
			t.Errorf("Test %d got unexpected value, want %t, got %t", i, test.exposure, mustExpose)
		}
	}
}

func TestRpcMethodAllowAllFilter(t *testing.T) {
	method_filters_toml := `
		Allow = [ 
			"*"
		]

		Deny = [ 
			"mtd1",
			"*staking*",
			"eth_get*",
			"^itc_[a-z]+"
		]
	`
	b := []byte(method_filters_toml)

	var rmf RpcMethodFilter
	rmf.LoadRpcMethodFilters(b)

	tests := []struct {
		name     string
		exposure bool
	}{
		0: {"mtd1", false},
		1: {"itc_method1", false},
		2: {"itcv2_method5", true},
		3: {"itcv2_method", true},
		4: {"eth_chainID", true},
		5: {"eth_getValidator", false},
		6: {"itc_getStakingInfo", false},
		7: {"abc", true},
		8: {"itc_getStakingNetworkInfo", false},
	}

	for i, test := range tests {
		mustExpose := rmf.Expose(test.name)

		if mustExpose != test.exposure {
			t.Errorf("Test %d got unexpected value, want %t, got %t", i, test.exposure, mustExpose)
		}
	}
}

func TestRpcMethodDenyAllFilter(t *testing.T) {
	method_filters_toml := `
		Allow = [ 
			"mtd1",
			"*staking*",
			"eth_get*",
			"regex:^itc_[a-z]+"
		]

		Deny = [ 
			"*"
		]
	`
	b := []byte(method_filters_toml)

	var rmf RpcMethodFilter
	rmf.LoadRpcMethodFilters(b)

	tests := []struct {
		name     string
		exposure bool
	}{
		0: {"mtd1", false},
		1: {"itc_method1", false},
		2: {"itcv2_method5", false},
		3: {"itcv2_method", false},
		4: {"eth_chainID", false},
		5: {"eth_getValidator", false},
		6: {"itc_getStakingInfo", false},
		7: {"abc", false},
		8: {"itc_getStakingNetworkInfo", false},
	}

	for i, test := range tests {
		mustExpose := rmf.Expose(test.name)

		if mustExpose != test.exposure {
			t.Errorf("Test %d got unexpected value, want %t, got %t", i, test.exposure, mustExpose)
		}
	}
}

func TestEmptyRpcMethodFilter(t *testing.T) {

	b := []byte("")
	var rmf RpcMethodFilter
	rmf.LoadRpcMethodFilters(b)

	tests := []struct {
		name     string
		exposure bool
	}{
		0: {"itc_method1", true},
		1: {"itc_method2", true},
		2: {"itcv2_method5", true},
		3: {"itcv2_method", true},
		4: {"eth_chainID", true},
		5: {"eth_getValidator", true},
		6: {"itc_getStakingInfo", true},
		7: {"abc", true},
		8: {"itc_getNetworkInfo", true},
	}

	for i, test := range tests {
		mustExpose := rmf.Expose(test.name)

		if mustExpose != test.exposure {
			t.Errorf("Test %d got unexpected value, want %t, got %t", i, test.exposure, mustExpose)
		}
	}
}

func TestFilter(t *testing.T) {
	tests := []struct {
		input             string
		pattern           string
		expectedAllowance bool
	}{
		0:  {"abc", "abc", true},
		1:  {"AbC", "abc", true},        // case-insensitive
		2:  {"AbC", "exact:AbC", true},  // case-insensitive
		3:  {"AbC", "exact:abc", false}, // case-insensitive
		4:  {"abcd", "*", true},         // check * to pass everything
		5:  {"abc", "simple:abc", true}, // check simple matching
		6:  {"abcd", "simple:abc", false},
		7:  {"abcd", "regex:^a([a-z]+)d$", true}, // check regex
		8:  {"abcde", "regex:^a([a-z]+)d$", false},
		9:  {"abcd", "^a([a-z]+)d$", true}, // auto detected regex
		10: {"abc", "wildcard:abc*", true}, // check wild card
		11: {"abc", "abc*", true},          // auto detected wild card
		12: {"abcdef", "abc*", true},
		13: {"dabcd", "?abc*", true}, // check * and ? for wild card
		14: {"abc", "*abc?", false},  // ? can't be empty
		15: {"abcdef", "*a?c*", true},
		16: {"defabc", "*ab?*", true},
		17: {"defabcghi", "*abc*", true},
		18: {"ab", "*abc*", false},
		19: {"defghabc", "*abc*", true},
	}

	for i, test := range tests {
		isAllowed := Match(test.pattern, test.input)

		if isAllowed != test.expectedAllowance {
			t.Errorf("Test %d got unexpected value, want %t, got %t", i, test.expectedAllowance, isAllowed)
		}
	}
}
