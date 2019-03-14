package devicewallet

import (
	"github.com/skycoin/hardware-wallet-go/interfaces"
	"github.com/skycoin/hardware-wallet-go/test/mocks"
	"testing"

	"github.com/stretchr/testify/suite"
)

type devicerSuit struct {
	suite.Suite
}

func (suite *devicerSuit) SetupTest() {
}

func TestDevicerSuitSuit(t *testing.T) {
	suite.Run(t, new(devicerSuit))
}

func (suite *devicerSuit) TestGenerateMnemonic() {
	// NOTE(denisacostaq@gmail.com): Giving
	driver := &mocks.DeviceDriver{}
	driver.On("DeviceType").Return(interfaces.DeviceTypeEmulator)
	device := Device{driver}

	// NOTE(denisacostaq@gmail.com): When
	device.GenerateMnemonic(12, false)

	// NOTE(denisacostaq@gmail.com): Assert
}
