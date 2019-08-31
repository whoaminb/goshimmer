package stringify

import (
	"fmt"
	"testing"
)

func TestMap(t *testing.T) {
	testMap := make(map[string][]byte)

	testMap["huhu"] = []byte{1, 2}
	testMap["haha"] = []byte{1, 2, 3, 4}

	fmt.Println(Map(testMap))
}
