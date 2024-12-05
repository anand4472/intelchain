// Copyright 2022 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package rawdb

// The list of table names of chain freezer.
// This variables is NOT used, just ported over from the Ethereum
const (
	// ChainFreezerHeaderTable indicates the name of the freezer header table.
	ChainFreezerHeaderTable = "headers"

	// ChainFreezerHashTable indicates the name of the freezer canonical hash table.
	ChainFreezerHashTable = "hashes"

	// ChainFreezerBodiesTable indicates the name of the freezer block body table.
	ChainFreezerBodiesTable = "bodies"

	// ChainFreezerReceiptTable indicates the name of the freezer receipts table.
	ChainFreezerReceiptTable = "receipts"

	// ChainFreezerDifficultyTable indicates the name of the freezer total difficulty table.
	ChainFreezerDifficultyTable = "diffs"
)

// chainFreezerNoSnappy configures whether compression is disabled for the ancient-tables.
// Hashes and difficulties don't compress well.
// This function is NOT used, just ported over from the Ethereum
var chainFreezerNoSnappy = map[string]bool{
	ChainFreezerHeaderTable:     false,
	ChainFreezerHashTable:       true,
	ChainFreezerBodiesTable:     false,
	ChainFreezerReceiptTable:    false,
	ChainFreezerDifficultyTable: true,
}

// The list of identifiers of ancient stores.
var (
	chainFreezerName = "chain" // the folder name of chain segment ancient store.
)

// freezers the collections of all builtin freezers.
var freezers = []string{chainFreezerName}