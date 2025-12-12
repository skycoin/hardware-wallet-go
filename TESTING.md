# Testing USB Permissions Fix

## Changes Made

### Branch: fix/usb-permissions-libusb

**Files Modified:**
1. `src/usb/lowlevel/libusb/libusb.go`
   - Added `Get_Device_List_Filtered()` to only open devices matching specific VID/PID
   - Fixed `Get_Config_Descriptor()` to properly handle gousb's map-based Configs structure  
   - Fixed endpoint descriptor population (Endpoints is a map, not slice)

2. `src/skywallet/usb/libusb.go`
   - Updated `Enumerate()` to use filtered device list
   - Fixed `matchVidPid()` to accept Skycoin devices (VID 0x313a) in all modes
   - Added `SetAutoDetach(true)` call in `connect()` for automatic kernel driver handling

## Testing with Replace Directive

In `skycoin/go.mod`, add:
```
replace github.com/skycoin/hardware-wallet-go => ../../skycoin/hardware-wallet-go
```

Then build:
```bash
cd skycoin
go build -mod=mod -o hw ./cmd/hardware-wallet/skycoin.go
./hw cli features
```

## Current Status

- ✅ Config descriptor parsing fixed
- ✅ VID/PID matching fixed (Skycoin devices now recognized)
- ✅ Filtered device enumeration (avoids permission errors on unrelated USB devices)
- ⚠️  SetAutoDetach() still requires elevated privileges

The "libusb: bad access [code -3]" error persists because `SetAutoDetach()` requires 
privileged USB ioctls that normal users don't have, even with correct udev rules and 
device permissions.

## Next Steps

Need to investigate:
1. Whether SetAutoDetach() can be avoided if kernel driver is not attached
2. Alternative approaches to kernel driver management without root
3. Testing with actual sudo to verify all other fixes work correctly

