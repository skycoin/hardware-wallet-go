# Linux Setup Guide for SkyWallet

This guide covers the necessary setup steps to use the SkyWallet hardware wallet on Linux without requiring root privileges.

## Table of Contents
- [Prerequisites](#prerequisites)
- [udev Rules Setup](#udev-rules-setup)
- [Kernel Driver Management](#kernel-driver-management)
- [Verification](#verification)
- [Troubleshooting](#troubleshooting)

## Prerequisites

- Linux kernel 3.x or later
- libusb 1.0.9 or later
- systemd (for TAG+="uaccess" support)

Install dependencies:

**Ubuntu/Debian:**
```bash
sudo apt-get install libusb-1.0-0-dev libudev-dev
```

**Fedora/RHEL:**
```bash
sudo dnf install libusb-devel systemd-devel
```

**Arch:**
```bash
sudo pacman -S libusb systemd
```

## udev Rules Setup

udev rules grant your user account permission to access the SkyWallet USB device without sudo.

### Installation

1. Copy the udev rules file to the system directory:
   ```bash
   sudo cp udev/51-skywallet.rules /etc/udev/rules.d/
   ```

2. Reload udev rules:
   ```bash
   sudo udevadm control --reload-rules
   sudo udevadm trigger
   ```

3. Unplug and replug your SkyWallet device

### Manual Rule Creation

If you prefer to create the file manually:

```bash
sudo tee /etc/udev/rules.d/51-skywallet.rules > /dev/null <<EOF
# SkyWallet Hardware Wallet
SUBSYSTEM=="usb", ATTR{idVendor}=="313a", ATTR{idProduct}=="0001", MODE="0666", TAG+="uaccess"
EOF

sudo udevadm control --reload-rules
sudo udevadm trigger
```

## Kernel Driver Management

### The Issue

Linux automatically binds the `usbhid` kernel driver to HID devices (including SkyWallet). This prevents libusb from accessing the device. You must unbind the kernel driver before using the hardware wallet.

### Option 1: Manual Unbind (Recommended for Testing)

Each time you plug in the device, run:

```bash
# Find the SkyWallet USB interface
INTERFACE=$(for iface in /sys/bus/usb/devices/*:*.*; do
  if [ -L "$iface/driver" ]; then
    driver=$(basename $(readlink "$iface/driver"))
    if [ "$driver" = "usbhid" ]; then
      parent="${iface%:*}"
      if [ -f "$parent/idVendor" ]; then
        vid=$(cat "$parent/idVendor" 2>/dev/null)
        if [ "$vid" = "313a" ]; then
          basename "$iface"
        fi
      fi
    fi
  fi
done | head -1)

# Unbind the kernel driver
echo "$INTERFACE" | sudo tee /sys/bus/usb/drivers/usbhid/unbind
```

**Or use the provided script:**

```bash
# In the skycoin main project (if available)
./unbind-skywallet.sh
```

### Option 2: Automatic Unbind via udev (Permanent Solution)

Create an additional udev rule to automatically unbind the driver on device insertion:

```bash
sudo tee /etc/udev/rules.d/52-skywallet-unbind.rules > /dev/null <<EOF
# Automatically unbind usbhid from SkyWallet
ACTION=="add", SUBSYSTEM=="usb", ATTRS{idVendor}=="313a", ATTRS{idProduct}=="0001", \\
  RUN+="/bin/sh -c 'for i in /sys/bus/usb/devices/*:1.0; do if [ -e \$i/driver ]; then echo \$(basename \$i) > /sys/bus/usb/drivers/usbhid/unbind; fi; done'"
EOF

sudo udevadm control --reload-rules
```

**Note:** This approach may require adjustment based on your specific system configuration.

### Option 3: Kernel Module Parameter (Advanced)

Prevent the HID driver from binding to SkyWallet by adding a quirk:

```bash
echo "options usbhid quirks=0x313A:0x0001:0x0004" | sudo tee /etc/modprobe.d/skywallet.conf
sudo update-initramfs -u  # Ubuntu/Debian
# OR
sudo dracut -f            # Fedora/RHEL
```

Reboot for this to take effect.

## Verification

### Check Device Permissions

1. List USB devices and find SkyWallet:
   ```bash
   lsusb | grep -i skywallet
   # Output: Bus 001 Device 012: ID 313a:0001 SkycoinFoundation SKYWALLET
   ```

2. Check device file permissions (use bus/device numbers from above):
   ```bash
   ls -l /dev/bus/usb/001/012
   # Should show: crw-rw-rw-+ ... (mode 0666)
   ```

3. Verify ACLs grant your user access:
   ```bash
   getfacl /dev/bus/usb/001/012
   # Should show: user:yourusername:rw-
   ```

### Check Kernel Driver Status

```bash
# Find the device
for d in /sys/bus/usb/devices/*/idVendor; do
  if grep -q "313a" "$d" 2>/dev/null; then
    parent=$(dirname "$d")
    echo "Device: $(basename $parent)"
    
    # Check if usbhid is bound
    if [ -L "$parent:1.0/driver" ]; then
      driver=$(basename $(readlink "$parent:1.0/driver"))
      echo "  Driver: $driver"
      if [ "$driver" = "usbhid" ]; then
        echo "  ⚠️  WARNING: usbhid is still bound - run unbind script"
      fi
    else
      echo "  ✓ No driver bound (ready for use)"
    fi
  fi
done
```

### Test Hardware Wallet Access

```bash
# Build and run the CLI tool
go build -o hw cmd/cli/cli.go
./hw features

# Should return device information without errors
```

## Troubleshooting

### "Permission denied" or "Access denied" Errors

**Symptoms:**
- `libusb: bad access [code -3]`
- Cannot open device without sudo

**Solutions:**
1. Verify udev rules are installed correctly
2. Unplug and replug the device
3. Check that you're in the `plugdev` group (some distros):
   ```bash
   sudo usermod -a -G plugdev $USER
   # Log out and back in
   ```

### Device Still Requires Sudo

**Cause:** Kernel driver (usbhid) is still bound to the device

**Solution:** Run the unbind script before accessing the device:
```bash
./unbind-skywallet.sh
```

### "Device not found" Errors

**Check if device is detected:**
```bash
lsusb | grep 313a
```

**If not visible:**
- Try a different USB port
- Check USB cable
- Check device is powered on
- Try `sudo dmesg | tail` after plugging in to see kernel messages

### Communication Hangs or Timeouts

**Cause:** May be related to USB power management

**Solution:** Disable USB autosuspend for the device:
```bash
# Find device path
DEVPATH=$(for d in /sys/bus/usb/devices/*/idVendor; do
  grep -q "313a" "$d" && dirname "$d"
done)

# Disable autosuspend
echo on | sudo tee $DEVPATH/power/control
```

Or add a udev rule:
```bash
echo 'SUBSYSTEM=="usb", ATTR{idVendor}=="313a", ATTR{idProduct}=="0001", ATTR{power/control}="on"' | \
  sudo tee -a /etc/udev/rules.d/51-skywallet.rules
```

### Checking Library Versions

```bash
# Check libusb version
pkg-config --modversion libusb-1.0

# Check systemd version (for uaccess support)
systemctl --version
```

## Additional Resources

- [USB Device IDs](http://www.linux-usb.org/usb-ids.html)
- [udev Rules Guide](https://wiki.archlinux.org/title/Udev)
- [libusb Documentation](https://libusb.info/)
- [SkyWallet Hardware Repository](https://github.com/SkycoinProject/hardware-wallet)

## Security Notes

- The `MODE="0666"` rule makes the device accessible to all users on the system. This is generally safe for hardware wallets as they require physical button confirmation for sensitive operations.
- The `TAG+="uaccess"` rule restricts access to currently logged-in users on systemd systems, providing better security.
- Never modify the udev rules to add `GROUP="users"` or similar unless you understand the security implications.

## Support

If you encounter issues not covered in this guide:

1. Check the [GitHub Issues](https://github.com/SkycoinProject/hardware-wallet-go/issues)
2. Review the [USB Fix Summary](USB-FIX-SUMMARY.md) for advanced troubleshooting
3. Join the Skycoin community channels

---

**Last Updated:** 2025-12-12  
**Applies to:** hardware-wallet-go v1.x on Linux
