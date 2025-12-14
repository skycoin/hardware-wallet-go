## Current Status Summary

### Problem

When the  hardware-wallet-go  package is vendored in another repository (like  skycoin/skycoin ), CGO compilation
fails with:

fatal error: os/threads_posix.c: No such file or directory

### Root Cause

The C source files in  src/usb/lowlevel/libusb/c/  and  src/usb/lowlevel/hidapi/c/  subdirectories are NOT being
vendored by  go mod vendor  because Go only vendors files in directories that contain Go source files.

### What We Fixed Successfully

1. ✅ Changed  #cgo CFLAGS: -I./c  to  #cgo CFLAGS: -I${SRCDIR}/c  in:
•  src/usb/lowlevel/libusb/libusb.go  (line 24)
•  src/usb/lowlevel/hidapi/hid.go  (line 16)
2. ✅ Added blank imports to force vendoring:
•  src/usb/lowlevel/libusb/libusb.go  imports  c  and  c/os  packages
•  src/usb/lowlevel/hidapi/hid.go  imports  c ,  c/mac ,  c/windows  packages


### What We Tried (Failed Approaches)

1. ❌ Created stub .go files with build tags - Go complains "C source files not allowed when not using cgo"
2. ❌ Added  import "C"  to stub files - Tries to compile C files standalone (fails)
3. ❌ Used  //go:build ignore  - Prevents vendoring entirely

### Current State

• The stub .go files have been REMOVED
• The blank imports are still in libusb.go and hid.go but have no packages to import
• The C files will NOT be vendored

### Next Steps to Try

1. Create stub .go files that DO use  import "C"  but have proper headers - The C files need proper #include guards
and to be compilable on their own, OR
2. Use a different vendoring mechanism - Maybe create a .c file in the parent directory that includes all the
subdirectory C files, OR
3. Document that this package cannot be vendored and must be used as a direct dependency, OR
4. Copy the C file inclusion approach - Put all C includes directly in libusb.go/hid.go without subdirectories

The most promising approach is probably #4 - flatten the structure or explicitly list all C files in the parent
directory with proper include paths.

Files Modified:

•  src/usb/lowlevel/libusb/libusb.go 
•  src/usb/lowlevel/hidapi/hid.go 

## Conversation Summary

### Original Tasks Completed ✅

1. Fixed libusb.go compilation errors - Corrected missing  return  statements, type conversions, and duplicate code in
 src/usb/lowlevel/libusb/libusb.go 
2. Migrated CLI from urfave/cli to spf13/cobra - Successfully converted all 19 command files in  src/cli/  to use
cobra, updated  go.mod , and verified the build works locally

### New Problem Identified 🔴

When integrating the CLI into the upstream  skycoin/skycoin  repo, vendoring fails with:

fatal error: os/threads_posix.c: No such file or directory

Root cause:  go mod vendor  only vendors directories containing Go source files. The C files in subdirectories ( c/ ,
 c/os/ ,  c/mac/ ,  c/windows/ ) aren't being vendored.

### Changes Made for Vendoring

Files Modified:

•  src/usb/lowlevel/libusb/libusb.go  - Changed  -I./c  to  -I${SRCDIR}/c  (line 24)
•  src/usb/lowlevel/hidapi/hid.go  - Changed  -I./c  to  -I${SRCDIR}/c  (line 16)

Attempted Solutions (all failed):

1. Created stub .go files in C directories → "C source files not allowed when not using cgo"
2. Added  import "C"  to stubs → Tries to compile C files standalone (fails due to missing headers)
3. Used  //go:build ignore  → Prevents vendoring entirely

Current State: Stub files removed, blank imports orphaned, C files won't vendor.

### Next Steps

Most viable approaches:

1. Flatten structure - Move all C files to parent directories or copy them
2. Alternative solution - Use Go embed or asset tools to include C files
3. Document limitation - State package must be direct dependency, not vendored
4. Different approach - Investigate if  go mod vendor  flags or go.mod directives can force C file inclusion

The fundamental issue: Go's vendoring mechanism doesn't handle C-only subdirectories well when using CGO.

Current working directory of the persistent shell

/home/d0mo/go/src/github.com/0pcom/hardware-wallet-go
