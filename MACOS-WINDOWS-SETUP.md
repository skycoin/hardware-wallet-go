# macOS and Windows Setup Guide for Skycoin Hardware Wallet

This guide explains how to set up the required USB libraries for the Skycoin hardware wallet on macOS and Windows.

## Overview

The hardware wallet uses two USB communication methods:
- **libusb-1.0**: For WebUSB/vendor-specific interfaces (all platforms)
- **HIDAPI**: For HID interfaces (macOS and Windows only, Linux uses libusb)

Both libraries must be installed on your system before building or running the hardware wallet software.

---

## macOS Setup

### Prerequisites
- Homebrew package manager: https://brew.sh/
- Xcode Command Line Tools: `xcode-select --install`

### Install Required Libraries

```bash
# Install libusb-1.0
brew install libusb

# Install HIDAPI
brew install hidapi
```

### Verify Installation

```bash
# Check libusb
pkg-config --modversion libusb-1.0

# Check hidapi  
brew list hidapi
ls -la /usr/local/lib/libhidapi.dylib
```

### Build the Hardware Wallet CLI

```bash
cd hardware-wallet-go
make build
```

The binary will be created in `$GOPATH/bin/skycoin-hw-cli`.

### Troubleshooting

**Issue**: `ld: library not found for -lhidapi`

**Solution**: Ensure hidapi is properly linked:
```bash
brew link hidapi
export CGO_LDFLAGS="-L/usr/local/lib"
export CGO_CFLAGS="-I/usr/local/include"
```

**Issue**: `permission denied` when accessing device

**Solution**: macOS requires no special permissions for USB HID devices. If using libusb mode, you may need to run with sudo or create a codeless kext (advanced).

---

## Windows Setup

### Prerequisites
- Git for Windows: https://git-scm.com/download/win
- Go 1.21 or later: https://golang.org/dl/
- MinGW-w64 (for CGO): https://www.mingw-w64.org/
- Microsoft Visual C++ Build Tools (alternative to MinGW)

### Install MinGW-w64

1. Download from: https://sourceforge.net/projects/mingw-w64/
2. During installation, select:
   - Architecture: x86_64
   - Threads: posix
   - Exception: seh
3. Add `C:\mingw-w64\x86_64-8.1.0-posix-seh-rt_v6-rev0\mingw64\bin` to PATH

### Install libusb-1.0

**Option A: Pre-built DLL (Recommended)**

1. Download latest Windows binaries from: https://github.com/libusb/libusb/releases
2. Extract `libusb-1.0.dll` from `MinGW64\dll\` folder
3. Copy to one of:
   - `C:\Windows\System32\` (system-wide)
   - Same directory as your executable (app-specific)
   - Any directory in your PATH

**Option B: Build from source**

```cmd
git clone https://github.com/libusb/libusb.git
cd libusb
./configure --prefix=/mingw64
make
make install
```

### Install HIDAPI

**Option A: Pre-built DLL (Recommended)**

1. Download latest Windows binaries from: https://github.com/libusb/hidapi/releases
2. Extract `hidapi.dll` (or `hidapi-win.dll`)
3. Copy to same locations as libusb-1.0.dll above

**Option B: Build from source**

```cmd
git clone https://github.com/libusb/hidapi.git
cd hidapi
mkdir build && cd build
cmake .. -G "MinGW Makefiles" -DCMAKE_INSTALL_PREFIX=/mingw64
mingw32-make
mingw32-make install
```

### Build the Hardware Wallet CLI

```cmd
cd hardware-wallet-go
set CGO_ENABLED=1
go build -o skycoin-hw-cli.exe ./cmd/cli
```

### Verify DLL Dependencies

```cmd
# Check required DLLs
dumpbin /dependents skycoin-hw-cli.exe

# Or use Dependency Walker: https://www.dependencywalker.com/
```

### Troubleshooting

**Issue**: `undefined reference to 'hid_init'`

**Solution**: Ensure hidapi library is in linker path:
```cmd
set CGO_LDFLAGS=-L/mingw64/lib
set CGO_CFLAGS=-I/mingw64/include
```

**Issue**: `The code execution cannot proceed because hidapi.dll was not found`

**Solution**: 
- Copy `hidapi.dll` to the same directory as `skycoin-hw-cli.exe`
- OR add the directory containing `hidapi.dll` to your PATH

**Issue**: `The code execution cannot proceed because libusb-1.0.dll was not found`

**Solution**:
- Copy `libusb-1.0.dll` to the same directory as `skycoin-hw-cli.exe`  
- OR add the directory containing `libusb-1.0.dll` to your PATH

**Issue**: Windows Defender blocks the executable

**Solution**: Add exception or sign the binary with a code signing certificate

---

## For End Users (Binary Distribution)

### macOS

1. Install dependencies:
   ```bash
   brew install libusb hidapi
   ```

2. Download and run the wallet:
   ```bash
   chmod +x skycoin-hw-cli
   ./skycoin-hw-cli features
   ```

### Windows

1. Download libusb-1.0.dll and hidapi.dll:
   - libusb: https://github.com/libusb/libusb/releases
   - hidapi: https://github.com/libusb/hidapi/releases

2. Place DLLs in same folder as `skycoin-hw-cli.exe`

3. Run the wallet:
   ```cmd
   skycoin-hw-cli.exe features
   ```

---

## Library Versions

Tested with:
- **libusb-1.0**: >= 1.0.24
- **HIDAPI**: >= 0.12.0

Newer versions should work but haven't been extensively tested.

---

## Platform-Specific Notes

### macOS (IOKit Backend)

- HIDAPI on macOS uses native IOKit framework (no extra library needed at runtime)
- libusb-1.0 must still be installed via Homebrew
- USB HID devices are accessible without special permissions
- WebUSB devices may require detaching kernel driver (requires root)

### Windows (WinAPI Backend)

- HIDAPI on Windows uses native Windows HID API (hid.dll, setupapi.dll)
- DLLs must be in PATH or same directory as executable
- Windows 10/11 have native HID support, no drivers needed for most devices
- Some devices may require Zadig to install WinUSB driver: https://zadig.akeo.ie/

### Linux (Reference)

- Uses libusb-1.0 for both HID and WebUSB (no HIDAPI needed)
- Static musl builds bundle libusb-1.0 for portability
- Requires udev rules for non-root access (see LINUX-SETUP.md)

---

## Support

For issues:
- macOS/Windows build problems: Check homebrew/mingw installation
- Runtime DLL errors: Verify library installation and PATH
- USB device not found: Check drivers and permissions
- Hardware wallet specific: https://github.com/skycoin/hardware-wallet-go/issues
