package datastructure

import (
	"bytes"
	"math/rand"
	"testing"

	"github.com/iotaledger/iota.go/trinary"
	"github.com/stretchr/testify/assert"
)

func BenchmarkByteTrie_Insert(b *testing.B) {
	trie := &PrefixTrie{
		value:    nil,
		children: make(map[byte]*PrefixTrie),
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		token := make([]byte, 50)
		rand.Read(token)

		trie.Insert(token)
	}
}

func TestPrefixTrie_GetPrefix(t *testing.T) {
	trie := &PrefixTrie{
		value:    nil,
		children: make(map[byte]*PrefixTrie),
	}

	var token []byte
	for i := 0; i < 64000; i++ {
		token = make([]byte, 50)
		rand.Read(token)

		trie.Insert(token)
	}

	assert.True(t, len(trie.GetPrefix(token)) <= 3)
}

func TestPrefixTrie_Insert(t *testing.T) {
	trie := NewPrefixTree()

	tx1Hash := trinary.MustTrytesToBytes("NDMXEXQFVRJKHVRGYURKMRMUUYUNPCREQHEZKNSEHK9SWSIQDU9IF9IHXTOIVVHWXGLJHHUR9NIGMAUQC")
	tx2Hash := trinary.MustTrytesToBytes("HWTXYVLGPKUUFXGBBGKLMRCI9VOW9MEJRCJPRMUGCKHOVCCMFSZNAFEWOOVYUEHGYDWPWBEKWJPAZ9999")
	tx3Hash := trinary.MustTrytesToBytes("IPVMHNESBZUKAIYFJCYEVXDBICITTLDUQIOVZODAISWCNEMBLWHVBDTEHNWRENIDEGVVODLXHTXGA9999")

	assert.Equal(t, 0, trie.GetSize())

	assert.Equal(t, true, trie.Insert(tx1Hash))
	assert.Equal(t, true, trie.Insert(tx2Hash))
	assert.Equal(t, true, trie.Insert(tx3Hash))
	assert.Equal(t, false, trie.Insert(tx1Hash))

	assert.Equal(t, 3, trie.GetSize())

	txFound := false
	for _, hash := range trie.Get(trie.GetPrefix(tx1Hash)) {
		if bytes.Equal(hash, tx1Hash) {
			txFound = true
		}
	}
	assert.True(t, txFound)

	assert.Equal(t, true, trie.Delete(tx1Hash))
	assert.Equal(t, false, trie.Delete(tx1Hash))
	assert.Equal(t, true, trie.Delete(tx2Hash))

	assert.Equal(t, 1, trie.GetSize())
}

func TestPrefixTrie_Get(t *testing.T) {
	trie := NewPrefixTree()

	tx1Hash := trinary.MustTrytesToBytes("NDMXEXQFVRJKHVRGYURKMRMUUYUNPCREQHEZKNSEHK9SWSIQDU9IF9IHXTOIVVHWXGLJHHUR9NIGMAUQC")
	tx2Hash := trinary.MustTrytesToBytes("HWTXYVLGPKUUFXGBBGKLMRCI9VOW9MEJRCJPRMUGCKHOVCCMFSZNAFEWOOVYUEHGYDWPWBEKWJPAZ9999")

	trie.Insert(tx1Hash)
	trie.Insert(tx2Hash)

	prefix := trie.GetPrefix(tx1Hash)

	resultsByPrefix := trie.Get(prefix)
	resultsByFullHash := trie.Get(tx1Hash)

	assert.Equal(t, len(resultsByPrefix), len(resultsByFullHash))
	for i, foundHashCandidate := range resultsByPrefix {
		assert.True(t, bytes.Equal(foundHashCandidate, resultsByFullHash[i]))
	}
}

func TestPrefixTrie_GetSize(t *testing.T) {
	trie := NewPrefixTree()

	assert.Equal(t, 0, trie.GetSize())

	tx1Hash := trinary.MustTrytesToBytes("NDMXEXQFVRJKHVRGYURKMRMUUYUNPCREQHEZKNSEHK9SWSIQDU9IF9IHXTOIVVHWXGLJHHUR9NIGMAUQC")
	tx2Hash := trinary.MustTrytesToBytes("HWTXYVLGPKUUFXGBBGKLMRCI9VOW9MEJRCJPRMUGCKHOVCCMFSZNAFEWOOVYUEHGYDWPWBEKWJPAZ9999")
	tx3Hash := trinary.MustTrytesToBytes("IPVMHNESBZUKAIYFJCYEVXDBICITTLDUQIOVZODAISWCNEMBLWHVBDTEHNWRENIDEGVVODLXHTXGA9999")

	trie.Insert(tx1Hash)
	trie.Insert(tx2Hash)
	assert.Equal(t, 2, trie.GetSize())

	trie.Insert(tx3Hash)
	assert.Equal(t, 3, trie.GetSize())

	trie.Delete(tx1Hash)
	assert.Equal(t, 2, trie.GetSize())

	trie.Delete(tx2Hash)
	assert.Equal(t, 1, trie.GetSize())

	trie.Delete(tx3Hash)
	assert.Equal(t, 0, trie.GetSize())

	trie.Delete(tx3Hash)
	assert.Equal(t, 0, trie.GetSize())
}

func TestPrefixTrie_Delete(t *testing.T) {
	trie := NewPrefixTree()

	tx1Hash := trinary.MustTrytesToBytes("NDMXEXQFVRJKHVRGYURKMRMUUYUNPCREQHEZKNSEHK9SWSIQDU9IF9IHXTOIVVHWXGLJHHUR9NIGMAUQC")
	tx2Hash := trinary.MustTrytesToBytes("HWTXYVLGPKUUFXGBBGKLMRCI9VOW9MEJRCJPRMUGCKHOVCCMFSZNAFEWOOVYUEHGYDWPWBEKWJPAZ9999")
	tx3Hash := trinary.MustTrytesToBytes("IPVMHNESBZUKAIYFJCYEVXDBICITTLDUQIOVZODAISWCNEMBLWHVBDTEHNWRENIDEGVVODLXHTXGA9999")

	trie.Insert(tx1Hash)
	trie.Insert(tx2Hash)

	assert.Equal(t, true, trie.Delete(tx1Hash))
	assert.Equal(t, false, trie.Delete(tx1Hash))

	assert.Equal(t, true, trie.Delete(tx2Hash))
	assert.Equal(t, false, trie.Delete(tx3Hash))

	assert.Equal(t, 0, trie.GetSize())
}
