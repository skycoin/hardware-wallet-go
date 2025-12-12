# USB Permissions Fix for Skycoin Hardware Wallet

## Summary

This branch fixes the "libusb: bad access [code -3]" error when accessing the Skycoin hardware wallet without sudo.

## Fixes Applied

### 1. Filtered Device Enumeration
**Problem:** Code tried to open ALL USB devices, triggering permission errors on inaccessible devices  
**Solution:** `Get_Device_List_Filtered()` only opens devices matching VID/PID 313a:0001

###2. VID/PID Matching
**Problem:** `matchVidPid()` only checked for Trezor2 devices (VID 0x1209), rejecting Skycoin wallets (VID 0x313a)  
**Solution:** Accept both Trezor and Skycoin devices in all modes

### 3. Config Descriptor Parsing
**Problem:** Treated gousb's `map[int]ConfigDesc` as array, causing index out of bounds  
**Solution:** Properly lookup by config number (1) not array index (0)

### 4. Endpoint Iteration
**Problem:** Treated `map[EndpointAddress]EndpointDesc` as slice  
**Solution:** Iterate over map keys correctly

## Usage

### Quick Test (Temporary)

1. **Unbind the HID driver:**
   ```bash
   # Find interface (usually 1-1.5:1.0 or similar)
   ls -la /sys/bus/usb/devices/*/driver 2>/dev/null | grep usbhid | grep 313a
   
   # Unbind (replace 1-1.5:1.0 with your interface)
   echo "1-1.5:1.0" | sudo tee /sys/bus/usb/drivers/usbhid/unbind
   ```

2. **Test the wallet:**
   ```bash
   ./hw cli features  # Should work without sudo!
   ```

### Permanent Solution (udev Rule)

Create `/etc/udev/rules.d/51-skywallet.rules`:

```udev
# Skycoin Hardware Wallet
ACTION=="add|bind", SUBSYSTEM=="usb", ENV{DEVTYPE}=="usb_interface", \
  ATTRS{idVendor}=="313a", ATTRS{idProduct}=="0001", ATTR{bInterfaceClass}=="03", \
  RUN+="/bin/sh -c 'echo $kernel > /sys/bus/usb/drivers/usbhid/unbind 2>/dev/null || true'"

SUBSYSTEM=="usb", ATTRS{idVendor}=="313a", ATTRS{idProduct}=="0001", MODE="0666"
```

Then reload:
```bash
sudo udevadm control --reload-rules && sudo udevadm trigger
```

Replug the device - it should work immediately!

## Technical Details

**Why unbind is required:**
- Skywallet interface uses HID class (bInterfaceClass=03)
- Linux automatically binds `usbhid` driver to HID devices
- Only one driver can claim an interface at a time
- Kernel driver detachment requires CAP_SYS_ADMIN (even with correct file permissions)
- udev can unbind on device insertion without requiring application privileges

**Alternatives considered:**
- `SetAutoDetach(true)` - requires CAP_SYS_ADMIN ✗
- Manual `detachKernelDriver()` - requires CAP_SYS_ADMIN ✗  
- `setcap cap_sys_admin+ep` on binary - security risk ✗
- udev unbind rule - works perfectly ✓

## Files Changed

- `src/usb/lowlevel/libusb/libusb.go` - Core USB wrapper fixes
- `src/skywallet/usb/libusb.go` - Device enumeration and matching
- `TESTING.md` - Detailed testing instructions
- `README-USB-FIX.md` - This file

## Next Steps

1. Test with unbind script: `./unbind-skywallet.sh` (in skycoin repo)
2. Verify works: `./hw cli features`
3. If successful, create PR to merge to develop
4. Update skycoin vendor after merge

