package datastructure

import (
	"bytes"
	"fmt"
	"sync"
)

type PrefixTrie struct {
	value    []byte
	children map[byte]*PrefixTrie
	size     int
	mutex    sync.RWMutex
}

func NewPrefixTree() *PrefixTrie {
	return &PrefixTrie{
		value:    nil,
		children: make(map[byte]*PrefixTrie),
	}
}

func (prefixTrie *PrefixTrie) Get(byteSequenceOrPrefix []byte) (result [][]byte) {
	prefixTrie.mutex.RLock()
	defer prefixTrie.mutex.RUnlock()

	result = make([][]byte, 0)

	currentNode := prefixTrie

	for currentLevel := 0; currentLevel < len(byteSequenceOrPrefix); currentLevel++ {
		if currentNode.value != nil && bytes.HasPrefix(currentNode.value, byteSequenceOrPrefix) {
			result = append(result, currentNode.value)

			return
		}

		if existingNode, exists := currentNode.children[byteSequenceOrPrefix[currentLevel]]; exists {
			currentNode = existingNode
		} else {
			// error tried to inflate non-existing entry
			return
		}
	}

	if currentNode.value != nil {
		result = append(result, currentNode.value)
	} else {
		// traverse child elements
		if false {
			fmt.Println("WAS")
		}
	}

	return
}

func (prefixTrie *PrefixTrie) GetPrefix(insertedBytes []byte) []byte {
	prefixTrie.mutex.RLock()
	defer prefixTrie.mutex.RUnlock()

	currentNode := prefixTrie
	currentLevel := 0

	for {
		// if we have reached our target
		if bytes.Equal(currentNode.value, insertedBytes) {
			return insertedBytes[:currentLevel]
		}

		// if we have arrived at the wrong node: return nil
		if currentNode.value != nil {
			return nil
		}

		if childNode, exists := currentNode.children[insertedBytes[currentLevel]]; !exists {
			return nil
		} else {
			currentNode = childNode
		}

		// increase level counter
		currentLevel++
	}
}

func (prefixTrie *PrefixTrie) Insert(byteSequence []byte) bool {
	prefixTrie.mutex.Lock()
	defer prefixTrie.mutex.Unlock()

	currentNode := prefixTrie
	currentLevel := 0

	for {
		// if we have reached our target node: insert value
		if currentNode.value == nil && len(currentNode.children) == 0 {
			currentNode.value = byteSequence

			prefixTrie.size++

			return true
		}

		// if we have reached a previous leaf
		if currentNode.value != nil {
			// return if same element
			if bytes.Equal(currentNode.value, byteSequence) {
				return false
			}

			// move current value to correct sub element
			currentNode.children[currentNode.value[currentLevel]] = &PrefixTrie{
				value:    currentNode.value,
				children: make(map[byte]*PrefixTrie),
			}

			// set the value to nil
			currentNode.value = nil
		}

		// traverse or create correct child element
		if existingChildNode, exists := currentNode.children[byteSequence[currentLevel]]; exists {
			currentNode = existingChildNode
		} else {
			newNode := &PrefixTrie{
				children: make(map[byte]*PrefixTrie),
			}

			currentNode.children[byteSequence[currentLevel]] = newNode

			currentNode = newNode
		}

		// increase level counter
		currentLevel++
	}
}

func (prefixTrie *PrefixTrie) Delete(byteSequence []byte) bool {
	prefixTrie.mutex.Lock()
	defer prefixTrie.mutex.Unlock()

	currentNode := prefixTrie

	// trivial case: delete from root
	if bytes.Equal(currentNode.value, byteSequence) {
		currentNode.value = nil

		prefixTrie.size--

		return true
	}

	// non-trivial case: delete leaf
	for currentLevel := 0; currentLevel < len(byteSequence); currentLevel++ {
		if existingNode, exists := currentNode.children[byteSequence[currentLevel]]; exists {
			if bytes.Equal(existingNode.value, byteSequence) {
				delete(currentNode.children, byteSequence[currentLevel])

				if len(currentNode.children) == 1 {
					for index, currentChild := range currentNode.children {
						if currentChild.value != nil {
							currentNode.value = currentChild.value
							delete(currentNode.children, index)
						}
					}
				}

				prefixTrie.size--

				return true
			}

			currentNode = existingNode
		} else {
			return false
		}
	}

	return false
}

func (prefixTrie *PrefixTrie) GetSize() int {
	prefixTrie.mutex.RLock()
	defer prefixTrie.mutex.RUnlock()

	return prefixTrie.size
}
