package usb

import (
	"encoding/hex"
	"log"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/skycoin/hardware-wallet-go/src/device-wallet/usb/usbhid"
)

const (
	webusbPrefix  = "web"
	webConfigNum  = 1
	webIfaceNum   = 0
	webAltSetting = 0
	webEpIn       = 0x81
	webEpOut      = 0x01
	usbTimeout    = 5000
)

type WebUSB struct {
	usb usbhid.Context
}

func InitWebUSB() (*WebUSB, error) {
	var usb usbhid.Context
	err := usbhid.Init(&usb)
	if err != nil {
		return nil, err
	}
	usbhid.Set_Debug(usb, int(usbhid.LOG_LEVEL_NONE))

	return &WebUSB{
		usb: usb,
	}, nil
}

func (b *WebUSB) Close() {
	usbhid.Exit(b.usb)
}

func (b *WebUSB) Enumerate(vendorID uint16, productID uint16) ([]Info, error) {
	list, err := usbhid.Get_Device_List(b.usb)
	if err != nil {
		return nil, err
	}
	defer usbhid.Free_Device_List(list, 1) // unlink devices

	var infos []Info

	// There is a bug in either Trezor T or libusb that makes
	// device appear twice with the same path
	paths := make(map[string]bool)

	for _, dev := range list {
		if b.match(dev) {
			dd, err := usbhid.Get_Device_Descriptor(dev)
			if err != nil {
				continue
			}
			path := b.identify(dev)
			inset := paths[path]
			if !inset {
				appendInfo := func() {
					infos = append(infos, Info{
						Path:      path,
						VendorID:  int(dd.IdVendor),
						ProductID: int(dd.IdProduct),
					})
					paths[path] = true
				}
				if vendorID != 0 && productID != 0 {
					if dd.IdVendor == vendorID && dd.IdProduct == productID {
						appendInfo()
					}
				} else if vendorID != 0 {
					if dd.IdVendor == vendorID {
						appendInfo()
					}
				} else if productID != 0 {
					if dd.IdProduct == productID {
						appendInfo()
					}
				} else {
					appendInfo()
				}
			}
		}
	}
	return infos, nil
}

func (b *WebUSB) Has(path string) bool {
	return strings.HasPrefix(path, webusbPrefix)
}

func (b *WebUSB) Connect(path string) (Device, error) {
	list, err := usbhid.Get_Device_List(b.usb)
	if err != nil {
		return nil, err
	}
	defer usbhid.Free_Device_List(list, 1) // unlink devices

	for _, dev := range list {
		if b.match(dev) && b.identify(dev) == path {
			return b.connect(dev)
		}
	}
	return nil, ErrNotFound
}

func (b *WebUSB) connect(dev usbhid.Device) (*WUD, error) {
	d, err := usbhid.Open(dev)
	if err != nil {
		return nil, err
	}
	err = usbhid.Reset_Device(d)
	if err != nil {
		// don't abort if reset fails
		// usbhid.Close(d)
		// return nil, err
		log.Printf("Warning: error at device reset: %s", err)
	}
	err = usbhid.Set_Configuration(d, webConfigNum)
	if err != nil {
		// don't abort if set configuration fails
		// usbhid.Close(d)
		// return nil, err
		log.Printf("Warning: error at configuration set: %s", err)
	}
	err = usbhid.Claim_Interface(d, webIfaceNum)
	if err != nil {
		usbhid.Close(d)
		return nil, err
	}
	return &WUD{
		dev:    d,
		closed: 0,
	}, nil
}

func (b *WebUSB) match(dev usbhid.Device) bool {
	dd, err := usbhid.Get_Device_Descriptor(dev)
	if err != nil {
		return false
	}

	vid := dd.IdVendor
	pid := dd.IdProduct
	if !b.matchVidPid(vid, pid) {
		return false
	}

	c, err := usbhid.Get_Active_Config_Descriptor(dev)
	if err != nil {
		return false
	}
	return (c.BNumInterfaces > webIfaceNum &&
		c.Interface[webIfaceNum].Num_altsetting > webAltSetting &&
		c.Interface[webIfaceNum].Altsetting[webAltSetting].BInterfaceClass == usbhid.CLASS_VENDOR_SPEC)
}

func (b *WebUSB) matchVidPid(vid uint16, pid uint16) bool {
	trezor1 := vid == VendorT1 && (pid == ProductT1Firmware)
	trezor2 := vid == VendorT2 && (pid == ProductT2Firmware || pid == ProductT2Bootloader)
	return trezor1 || trezor2
}

func (b *WebUSB) identify(dev usbhid.Device) string {
	var ports [8]byte
	p, err := usbhid.Get_Port_Numbers(dev, ports[:])
	if err != nil {
		return ""
	}
	return webusbPrefix + hex.EncodeToString(p)
}

type WUD struct {
	dev usbhid.Device_Handle

	closed int32 // atomic

	transferMutex sync.Mutex
	// closing cannot happen while interrupt_transfer is hapenning,
	// otherwise interrupt_transfer hangs forever
}

func (d *WUD) Close() error {
	atomic.StoreInt32(&d.closed, 1)

	d.finishReadQueue()

	d.transferMutex.Lock()
	usbhid.Close(d.dev)
	d.transferMutex.Unlock()

	return nil
}

func (d *WUD) finishReadQueue() {
	d.transferMutex.Lock()
	var err error
	var buf [64]byte

	for err == nil {
		_, err = usbhid.Interrupt_Transfer(d.dev, webEpIn, buf[:], 50)
	}
	d.transferMutex.Unlock()
}

func (d *WUD) readWrite(buf []byte, endpoint uint8) (int, error) {
	for {
		closed := (atomic.LoadInt32(&d.closed)) == 1
		if closed {
			return 0, errClosedDevice
		}

		d.transferMutex.Lock()
		p, err := usbhid.Interrupt_Transfer(d.dev, endpoint, buf, usbTimeout)
		d.transferMutex.Unlock()

		if err == nil {
			// sometimes, empty report is read, skip it
			if len(p) > 0 {
				return len(p), err
			}
		}

		if err != nil {
			if err.Error() == usbhid.Error_Name(int(usbhid.ERROR_IO)) ||
				err.Error() == usbhid.Error_Name(int(usbhid.ERROR_NO_DEVICE)) {
				return 0, errDisconnect
			}

			if err.Error() != usbhid.Error_Name(int(usbhid.ERROR_TIMEOUT)) {
				return 0, err
			}
		}

		// continue the for cycle
	}
}

func (d *WUD) Write(buf []byte) (int, error) {
	return d.readWrite(buf, webEpOut)
}

func (d *WUD) Read(buf []byte) (int, error) {
	return d.readWrite(buf, webEpIn)
}
