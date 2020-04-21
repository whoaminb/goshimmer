package tcrypto

import (
	"github.com/magiconair/properties/assert"
	"testing"
)

func TestNewRandomDKShare(t *testing.T) {
	_, err := NewRndDKShare(67, 100, 5)
	assert.Equal(t, err, nil)

	_, err = NewRndDKShare(1, 1, 0)
	assert.Equal(t, err, nil)

	_, err = NewRndDKShare(5, 4, 0)
	assert.Equal(t, err != nil, true)

	_, err = NewRndDKShare(4, 5, 6)
	assert.Equal(t, err != nil, true)
}
