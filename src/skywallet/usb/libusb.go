package usb

import (
	"encoding/hex"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"

	lowlevel "github.com/skycoin/hardware-wallet-go/src/usb/lowlevel/libusb"
)

const (
	libusbPrefix    = "lib"
	usbConfigNum    = 1
	usbConfigIndex  = 0
	transferTimeout = 250
)

type libusbIfaceData struct {
	number     uint8
	altSetting uint8
	epIn       uint8
	epOut      uint8
}

var normalIface = libusbIfaceData{
	number:     0,
	altSetting: 0,
	epIn:       0x81,
	epOut:      0x01,
}

// Old bootloader has different epOut
// We need it here, since on Linux,
// we use libusb instead of hidapi for old BL
var oldBLIface = libusbIfaceData{
	number:     0,
	altSetting: 0,
	epIn:       0x81,
	epOut:      0x02,
}

type LibUSB struct {
	usb    lowlevel.Context
	only   bool
	cancel bool
	detach bool
}

func InitLibUSB(onlyLibusb, allowCancel, detach bool) (*LibUSB, error) {
	var usb lowlevel.Context
	err := lowlevel.Init(&usb)
	if err != nil {
		return nil, fmt.Errorf(`error when initializing LibUSB.
Are you running skywallet in an environment without USB (for example, docker or travis)'.
Original error: %v`, err)
	}

	return &LibUSB{
		usb:    usb,
		only:   onlyLibusb,
		cancel: allowCancel,
		detach: detach,
	}, nil
}

func (b *LibUSB) Close() {
	lowlevel.Exit(b.usb)
}

func hasIface(dev lowlevel.Device, dIface libusbIfaceData, dClass uint8) (bool, error) {
	config, err := lowlevel.Get_Config_Descriptor(dev, usbConfigIndex)
	if err != nil {
		return false, err
	}

	ifaces := config.Interface
	for _, iface := range ifaces {
		for _, alt := range iface.Altsetting {
			if alt.BInterfaceNumber == dIface.number &&
				alt.BAlternateSetting == dIface.altSetting &&
				alt.BNumEndpoints == 2 &&
				alt.BInterfaceClass == dClass &&
				alt.Endpoint[0].BEndpointAddress == dIface.epIn &&
				alt.Endpoint[1].BEndpointAddress == dIface.epOut {
				return true, nil
			}
		}
	}
	return false, nil
}

func detectOldBL(dev lowlevel.Device) (bool, error) {
	return hasIface(dev, oldBLIface, uint8(lowlevel.CLASS_HID))
}

func (b *LibUSB) Enumerate(vendorID, productID uint16) ([]Info, error) {
	list, err := lowlevel.Get_Device_List(b.usb)

	if err != nil {
		return nil, err
	}

	defer func() {
		lowlevel.Free_Device_List(list, 1) // unlink devices
	}()

	var infos []Info

	// There is a bug in libusb that makes
	// device appear twice with the same path.
	// This is already fixed in libusb 2.0.12;
	// however, 2.0.12 has other problems with windows, so we
	// patchfix it here
	paths := make(map[string]bool)

	for _, dev := range list {
		m, t := b.match(dev)
		if m {
			dd, err := lowlevel.Get_Device_Descriptor(dev)
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
						Type:      t,
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

func (b *LibUSB) Has(path string) bool {
	return strings.HasPrefix(path, libusbPrefix)
}

func (b *LibUSB) Connect(path string) (Device, error) {
	list, err := lowlevel.Get_Device_List(b.usb)
	if err != nil {
		return nil, err
	}

	defer func() {
		lowlevel.Free_Device_List(list, 1) // unlink devices
	}()

	// There is a bug in libusb that makes
	// device appear twice with the same path.
	// This is already fixed in libusb 2.0.12;
	// however, 2.0.12 has other problems with windows, so we
	// patchfix it here
	mydevs := make([]lowlevel.Device, 0)
	for _, dev := range list {
		m, _ := b.match(dev)
		if m && b.identify(dev) == path {
			mydevs = append(mydevs, dev)
		}
	}

	err = ErrNotFound
	for _, dev := range mydevs {
		res, errConn := b.connect(dev)
		if errConn == nil {
			return res, nil
		}
		err = errConn
	}
	return nil, err
}

func (b *LibUSB) setConfiguration(d lowlevel.Device_Handle) {
	currConf, err := lowlevel.Get_Configuration(d)
	if err != nil {
		log.Errorf("current configuration err %s", err.Error())
	}

	if currConf != usbConfigNum {
		err = lowlevel.Set_Configuration(d, usbConfigNum)
		if err != nil {
			// don't abort if set configuration fails
			// lowlevel.Close(d)
			// return nil, err
			log.Errorf("Warning: error at configuration set: %s", err)
		}

		currConf, err = lowlevel.Get_Configuration(d)
		if err != nil {
			log.Errorf("current configuration err %s", err.Error())
		}
	}
}

func (b *LibUSB) claimInterface(d lowlevel.Device_Handle) (bool, error) {
	attach := false
	usbIfaceNum := int(normalIface.number)

	if b.detach {
		kernel, errD := lowlevel.Kernel_Driver_Active(d, usbIfaceNum)
		if errD != nil {
			lowlevel.Close(d)
			return false, errD
		}
		if kernel {
			attach = true
			errD = lowlevel.Detach_Kernel_Driver(d, usbIfaceNum)
			if errD != nil {
				lowlevel.Close(d)
				return false, errD
			}
		}
	}
	err := lowlevel.Claim_Interface(d, usbIfaceNum)
	if err != nil {
		lowlevel.Close(d)
		return false, err
	}

	return attach, nil
}

func (b *LibUSB) connect(dev lowlevel.Device) (*LibUSBDevice, error) {
	oldBL, err := detectOldBL(dev)
	if err != nil {
		return nil, err
	}

	d, err := lowlevel.Open(dev)
	if err != nil {
		return nil, err
	}

	b.setConfiguration(d)
	attach, err := b.claimInterface(d)
	if err != nil {
		return nil, err
	}
	return &LibUSBDevice{
		dev:    d,
		closed: 0,
		cancel: b.cancel,
		attach: attach,
		oldBL:  oldBL,
	}, nil
}

func matchType(dd *lowlevel.Device_Descriptor) DeviceType {
	if dd.IdProduct == ProductT1Firmware {
		// this is HID, in platforms where we don't use hidapi (linux, bsd)
		return TypeT1Hid
	}

	if dd.IdProduct == ProductT2Bootloader {
		if int(dd.BcdDevice>>8) == 1 {
			return TypeT1WebusbBoot
		}
		return TypeT2Boot
	}

	if int(dd.BcdDevice>>8) == 1 {
		return TypeT1Webusb
	}

	return TypeT2
}

func (b *LibUSB) match(dev lowlevel.Device) (bool, DeviceType) {
	dd, err := lowlevel.Get_Device_Descriptor(dev)
	if err != nil {
		log.Error("error getting descriptor -" + err.Error())
		return false, 0
	}

	vid := dd.IdVendor
	pid := dd.IdProduct
	if !b.matchVidPid(vid, pid) {
		return false, 0
	}

	c, err := lowlevel.Get_Active_Config_Descriptor(dev)
	if err != nil {
		log.Error("error getting config descriptor " + err.Error())
		return false, 0
	}

	var is bool
	usbIfaceNum := normalIface.number
	usbAltSetting := normalIface.altSetting
	if b.only {

		// if we don't use hidapi at all, keep HID devices
		is = (c.BNumInterfaces > usbIfaceNum &&
			c.Interface[usbIfaceNum].Num_altsetting > int(usbAltSetting))

	} else {

		is = (c.BNumInterfaces > usbIfaceNum &&
			c.Interface[usbIfaceNum].Num_altsetting > int(usbAltSetting) &&
			c.Interface[usbIfaceNum].Altsetting[usbAltSetting].BInterfaceClass == lowlevel.CLASS_VENDOR_SPEC)
	}

	if !is {
		return false, 0
	}
	return true, matchType(dd)
}

func (b *LibUSB) matchVidPid(vid uint16, pid uint16) bool {
	// Note: Trezor1 libusb will actually have the T2 vid/pid
	trezor2 := vid == VendorT2 && (pid == ProductT2Firmware || pid == ProductT2Bootloader)

	if b.only {
		trezor1 := vid == VendorT1 && (pid == ProductT1Firmware)
		return trezor1 || trezor2
	}

	return trezor2
}

func (b *LibUSB) identify(dev lowlevel.Device) string {
	var ports [8]byte
	p, err := lowlevel.Get_Port_Numbers(dev, ports[:])
	if err != nil {
		log.Errorf("error getting port numbers %s", err.Error())
		return ""
	}
	return libusbPrefix + hex.EncodeToString(p)
}

type LibUSBDevice struct {
	dev lowlevel.Device_Handle

	closed              int32 // atomic
	normalTransferMutex sync.Mutex
	debugTransferMutex  sync.Mutex
	// two interrupt_transfers should not happen at the same time

	cancel bool
	attach bool
	oldBL  bool
}

func (d *LibUSBDevice) Close(disconnected bool) error {
	atomic.StoreInt32(&d.closed, 1)

	if d.cancel {
		// libusb close does NOT cancel transfers on close
		// => we are using our own function that we added to libusb/sync.c
		// this "unblocks" Interrupt_Transfer in readWrite

		lowlevel.Cancel_Sync_Transfers_On_Device(d.dev)

		// reading recently disconnected device sometimes causes weird issues
		// => if we *know* it is disconnected, don't finish read queue
		//
		// Finishing read queue is not necessary when we don't allow cancelling
		// (since when we don't allow cancelling, we don't allow session stealing)
		if !disconnected {
			d.finishReadQueue()
		}
	}

	iface := int(normalIface.number)
	err := lowlevel.Release_Interface(d.dev, iface)
	if err != nil {
		// do not throw error, it is just release anyway
		log.Warnf("Warning: error at releasing interface: %s", err)
	}

	if d.attach {
		err = lowlevel.Attach_Kernel_Driver(d.dev, iface)
		if err != nil {
			// do not throw error, it is just re-attach anyway
			log.Warnf("Warning: error at re-attaching driver: %s", err)
		}
	}

	lowlevel.Close(d.dev)
	return nil
}

func (d *LibUSBDevice) transferMutexLock() {
	d.normalTransferMutex.Lock()
}

func (d *LibUSBDevice) transferMutexUnlock() {
	d.normalTransferMutex.Unlock()
}

func (d *LibUSBDevice) finishReadQueue() {
	usbEpIn := normalIface.epIn
	d.transferMutexLock()
	var err error
	var buf [64]byte

	for err == nil {
		// these transfers have timeouts => should not interfer with
		// cancel_sync_transfers_on_device
		_, err = lowlevel.Interrupt_Transfer(d.dev, usbEpIn, buf[:], 50)
	}
	d.transferMutexUnlock()
}

func (d *LibUSBDevice) readWrite(buf []byte, endpoint uint8) (int, error) {
	for {
		closed := (atomic.LoadInt32(&d.closed)) == 1
		if closed {
			return 0, ErrClosedDevice
		}

		d.transferMutexLock()
		p, err := lowlevel.Interrupt_Transfer(d.dev, endpoint, buf, transferTimeout)
		d.transferMutexUnlock()

		if err != nil {
			if isErrorDisconnect(err) {
				return 0, ErrDisconnect
			}

			return 0, err
		}

		// sometimes, empty report is read, skip it
		// TODO: is this still needed with 0 timeouts?
		if len(p) > 0 {
			return len(p), err
		}
		// continue the for cycle if empty transfer
	}
}

func isErrorDisconnect(err error) bool {
	// according to libusb docs, disconnecting device should cause only
	// LIBUSB_ERROR_NO_DEVICE error, but in real life, it causes also
	// LIBUSB_ERROR_IO, LIBUSB_ERROR_PIPE, LIBUSB_ERROR_OTHER

	return (err.Error() == lowlevel.Error_Name(int(lowlevel.ERROR_IO)) ||
		err.Error() == lowlevel.Error_Name(int(lowlevel.ERROR_NO_DEVICE)) ||
		err.Error() == lowlevel.Error_Name(int(lowlevel.ERROR_OTHER)) ||
		err.Error() == lowlevel.Error_Name(int(lowlevel.ERROR_PIPE)))
}

func (d *LibUSBDevice) Write(buf []byte) (int, error) {
	usbEpOut := normalIface.epOut
	if d.oldBL {
		usbEpOut = oldBLIface.epOut
	}
	return d.readWrite(buf, usbEpOut)
}

func (d *LibUSBDevice) Read(buf []byte) (c int, err error) {
	usbEpIn := normalIface.epIn
	for {
		c, err = d.readWrite(buf, usbEpIn)
		if err != nil && err.Error() == "LIBUSB_ERROR_TIMEOUT" {
			continue
		}
		return c, err
	}
	return c, err
}
