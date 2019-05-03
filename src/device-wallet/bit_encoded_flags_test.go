package devicewallet

import (
	"github.com/stretchr/testify/suite"
	"math/rand"
	"testing"
)

type bitEncodedFlagsSuit struct {
	suite.Suite
}


func TestBitEncodedFlagsSuit(t *testing.T) {
	suite.Run(t, new(bitEncodedFlagsSuit))
}


func (suite *bitEncodedFlagsSuit) TestOperationsAreReversible() {
	for i := 0; i < 100; i++ {
		// NOTE: Giving
		flags := rand.Uint64() % 32
		ff := NewFirmwareFeatures(flags)
		// NOTE: When
		suite.NoError(ff.Unmarshal())
		f, e := ff.Marshal()
		// NOTE: Assert
		suite.NoError(e)
		suite.Equal(flags, ff.flags)
		suite.Equal(flags, f)
	}
}