# Testing USB Permissions Fix

## Branch: fix/usb-permissions-libusb

### Changes Made

**1. src/usb/lowlevel/libusb/libusb.go**
- Added `Get_Device_List_Filtered()` - only opens devices matching specific VID/PID to avoid permission errors
- Fixed `Get_Config_Descriptor()` - properly handles gousb's map-based Configs (key=config number, not array index)
- Fixed endpoint iteration - Endpoints is `map[EndpointAddress]EndpointDesc`, not a slice

**2. src/skywallet/usb/libusb.go**
- Updated `Enumerate()` to use filtered device enumeration
- Fixed `matchVidPid()` - now accepts Skycoin devices (VID 0x313a, PID 0x0001) in all modes
- Removed `SetAutoDetach()` call - requires CAP_SYS_ADMIN even with correct udev rules

### Testing Instructions

**1. Add replace directive to skycoin/go.mod:**
```
replace github.com/skycoin/hardware-wallet-go => ../../skycoin/hardware-wallet-go
```

**2. Build:**
```bash
cd skycoin
go build -mod=mod -o hw ./cmd/hardware-wallet/skycoin.go
```

**3. Unbind the usbhid kernel driver:**

The Skywallet interface (bInterfaceClass=03 HID) gets automatically bound to the `usbhid` kernel driver. This must be unbound before the application can claim it.

**Find the interface path:**
```bash
# Find your device
lsusb | grep 313a

# If it shows "Bus 001 Device 012", check:
ls -la /sys/bus/usb/devices/*/driver | grep -B2 -A2 313a
# This will show something like "1-1.5:1.0"
```

**Unbind the driver (replace with your interface path):**
```bash
echo "1-1.5:1.0" | sudo tee /sys/bus/usb/drivers/usbhid/unbind
```

**Verify unbind:**
```bash
ls -la /sys/bus/usb/devices/1-1.5\:1.0/driver 2>/dev/null || echo "Successfully unbound"
```

**4. Test:**
```bash
./hw cli features
```

### Permanent Solution via udev

Create `/etc/udev/rules.d/51-skywallet.rules`:

```udev
# Skycoin Hardware Wallet - Unbind HID driver and set permissions
# VID=313a PID=0001

# Unbind HID driver when interface appears
ACTION=="add|bind", SUBSYSTEM=="usb", ENV{DEVTYPE}=="usb_interface", \
  ATTRS{idVendor}=="313a", ATTRS{idProduct}=="0001", \
  ATTR{bInterfaceClass}=="03", \
  RUN+="/bin/sh -c 'echo $kernel > /sys/bus/usb/drivers/usbhid/unbind 2>/dev/null || true'"

# Set device permissions  
SUBSYSTEM=="usb", ATTRS{idVendor}=="313a", ATTRS{idProduct}=="0001", MODE="0666", GROUP="plugdev"
KERNEL=="hidraw*", ATTRS{idVendor}=="313a", ATTRS{idProduct}=="0001", MODE="0666", GROUP="plugdev"
```

Reload udev rules:
```bash
sudo udevadm control --reload-rules
sudo udevadm trigger
```

Then replug the device.

### Current Status

✅ **Fixed:**
- Config descriptor parsing (gousb map structure)
- Device enumeration (filtered by VID/PID to avoid permission errors)  
- VID/PID matching for Skycoin hardware wallet (0x313a:0x0001)

✅ **Works without sudo** (after unbinding usbhid driver)

### Technical Notes

**Why unbind is needed:**
- Interface class 03 (HID) is auto-claimed by the `usbhid` kernel driver
- gousb can't access an interface claimed by another driver
- Both `SetAutoDetach()` and manual `detachKernelDriver()` require CAP_SYS_ADMIN
- udev rules can unbind automatically on device insertion

**Alternative approach** (not implemented):
- Build application with CAP_SYS_ADMIN capability: `sudo setcap cap_sys_admin+ep ./hw`
- Security risk - grants broad system privileges

