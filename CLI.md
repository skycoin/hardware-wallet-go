# CLI Documentation

Skycoin command line interface

The CLI command APIs can be used directly from a Go application, see [Skycoin CLI Godoc](https://godoc.org/github.com/skycoin/skycoin/src/cli).

<!-- MarkdownTOC autolink="true" bracket="round" levels="1,2,3" -->

- [Usage](#usage)
    - [Update firmware](#update-firmware)
    - [Ask device to generate addresses](#ask-device-to-generate-addresses)
    - [Configure device mnemonic](#configure-device-mnemonic)
    - [Ask device to generate mnemonic](#generate-mnemonic)
    - [Configure device PIN code](#configure-device-pin-code)
    - [Get firmware version](#get-version)
    - [Ask device to sign message](#ask-device-to-sign-message)
    - [Ask device to check signature](#ask-device-to-check-signature)
    - [Wipe device](#wipe-device)
    - [Ask the device to perform the seed backup procedure](#backup-device)
    - [Ask the device to perform the seed recovery procedure](#recovery-device)
- [Note](#note)

<!-- /MarkdownTOC -->


## Install

```bash
$ cd $GOPATH/src/github.com/skycoin/hardware-wallet-go/
$ ./install.sh
```

## Usage

After the installation, you can run `skycoin-cli` to see the usage:

```
$ skycoin-cli

NAME:
   skycoin-cli - the skycoin command line interface

USAGE:
   skycoin-cli [global options] command [command options] [arguments...]

VERSION:
   0.24.1

COMMANDS:
     deviceSetMnemonic              Configure the device with a mnemonic.
     deviceFeatures                 Ask the device Features.
     deviceGenerateMnemonic         Ask the device to generate a mnemonic and configure itself with it.
     deviceAddressGen               Generate skycoin addresses using the firmware
     deviceFirmwareUpdate           Update device's firmware.
     deviceSignMessage              Ask the device to sign a message using the secret key at given index.
     deviceCheckMessageSignature    Check a message signature matches the given address.
     deviceSetPinCode               Configure a PIN code on a device.
     deviceWipe                     Ask the device to wipe clean all the configuration it contains.
     deviceBackup                   Ask the device to perform the seed backup procedure.
     deviceGetVersion               Ask firmware version.
     deviceRecovery                 Ask the device to perform the seed recovery procedure.
     emulatorSetMnemonic            Configure an emulated device with a mnemonic.
     emulatorFeatures               Ask the emulator Features.
     emulatorGenerateMnemonic       Ask the device to generate a mnemonic and configure itself with it.
     emulatorAddressGen             Generate skycoin addresses using an emulated device.
     emulatorSignMessage            Ask the emulated device to sign a message using the secret key at given index.
     emulatorCheckMessageSignature  Check a message signature matches the given address.
     emulatorSetPinCode             Configure a PIN code on an emulated device.
     emulatorWipe                   Ask the emulator to wipe clean all the configuration it contains.
     emulatorBackup                 Ask the emulator to perform the seed backup procedure.
     emulatorGetVersion             Ask firmware version.
     emulatorRecovery               Ask the device to perform the seed recovery procedure.
     sandbox                        Sandbox.
     help, h                        Shows a list of commands or help for one command



GLOBAL OPTIONS:
   --help, -h     show help
   --version, -v  print the version
```

### Update firmware

To update firmware from a usb message, the device needs to be in "bootloader mode". To turn on on "bootloader mode" unplug your device, hold both buttons at the same time and plug it back on.

The use this command:


```bash
$skycoin-cli deviceFirmwareUpdate --file=[your firmware .bin file]
```

```
OPTIONS:
        --file string            Path to your firmware file
```

### Ask device to generate addresses

Generate skycoin addresses using the firmware

```bash
$skycoin-cli deviceAddressGen [number of addresses] [start index]
```

```
OPTIONS:
        --addressN value            Number of addresses to generate (default: 1)
        --startIndex value          Start to genereate deterministic addresses from startIndex (default: 0)
```

#### Examples
##### Text output

```bash
$skycoin-cli deviceAddressGen --addressN=2 --startIndex=0
```
<details>
 <summary>View Output</summary>

```
MessageSkycoinAddress 117! array size is 2
MessageSkycoinAddress 117! Answer is: 2EU3JbveHdkxW6z5tdhbbB2kRAWvXC2pLzw
MessageSkycoinAddress 117! Answer is: zC8GAQGQBfwk7vtTxVoRG7iMperHNuyYPs
```
</details>

### Configure device mnemonic

Configure the device with a mnemonic.

```bash
$skycoin-cli deviceSetMnemonic [mnemonic]
```

```
OPTIONS:
        --mnemonic value            Mnemonic that will be stored in the device to generate addresses.
```

#### Examples
##### Text output

```bash
$skycoin-cli deviceSetMnemonic --mnemonic="cloud flower upset remain green metal below cup stem infant art thank"
```
<details>
 <summary>View Output</summary>

```
MessageButtonAck Answer is: 2 / 
Ecloud flower upset remain green metal below cup stem infant art thank
```
</details>


### Get version

Ask firmware version.

```bash
$skycoin-cli deviceGetVersion
```

#### Examples
##### Text output

```bash
$skycoin-cli deviceGetVersion
```
<details>
 <summary>View Output</summary>

```
Firmware version is Firmware Version 1.6.1
```
</details>


### Generate mnemonic

Ask the device to generate a mnemonic and configure itself with it.

```bash
$skycoin-cli deviceGenerateMnemonic
```

#### Examples
##### Text output

```bash
$skycoin-cli deviceGenerateMnemonic
```
<details>
 <summary>View Output</summary>

```
2018/11/06 14:41:50 MessageButtonAck Answer is: 2 /
 Mnemonic successfully configured
```
</details>

### Configure device PIN code

Configure the device with a pin code.

```bash
$skycoin-cli deviceSetPinCode
```

#### Examples
##### Text output

```bash
$skycoin-cli deviceSetPinCode
```
<details>
 <summary>View Output</summary>

```
MessageButtonAck Answer is: 18 /

PinMatrixRequest response: 5757
Setting pin: 5757

MessagePinMatrixAck Answer is: 18 /

PinMatrixRequest response: 4343
Setting pin: 4343

MessagePinMatrixAck Answer is: 18 /

PinMatrixRequest response: 6262
Setting pin: 6262

MessagePinMatrixAck Answer is: 2 /

PIN changed
```

</details>

### Ask device to sign message

Ask the device to sign a message using the secret key at given index.

```bash
$skycoin-cli deviceSignMessage [address index] [message to sign]
```

```
OPTIONS:
        --addressN value            Index of the address that will issue the signature. (default: 0)
        --message value             The message that the signature claims to be signing.
```

#### Examples
##### Text output

```bash
$skycoin-cli deviceSignMessage  --addressN=2 --message="Hello World!"
```
<details>
 <summary>View Output</summary>

```
Success 2! address that issued the signature is: DEK8o3Dnnp8UfTZrZCcCPCA6oRLqDeuKKy85YoTmCjfR2xDcZCz1j6tC4nmaAxHH15wgff88R2xPatT4MRvGHz9nf
```
</details>

### Ask device to check signature

Check a message signature matches the given address.

```bash
$skycoin-cli deviceCheckMessageSignature [address] [signed message] [signature]
```

```
OPTIONS:
        --address value            Address that issued the signature.
        --message value            The message that the signature claims to be signing.
        --signature value          Signature of the message.
```

#### Examples
##### Text output

```bash
$skycoin-cli deviceCheckMessageSignature  --address=2EU3JbveHdkxW6z5tdhbbB2kRAWvXC2pLzw --message="Hello World!" --signature=GvKS4S3CA2YTpEPFA47yFdC5CP3y3qB18jwiX1URXqWQTvMjokd3A4upPz4wyeAyKJEtRdRDGUvUgoGASpsTTUeMn
```
<details>
 <summary>View Output</summary>

```
Success 2! address that issued the signature is: 
#2EU3JbveHdkxW6z5tdhbbB2kRAWvXC2pLzw

```
</details>

## Note

The `[option]` in subcommand must be set before the rest of the values, otherwise the `option` won't
be parsed. For example:

If we want to specify a `change address` in `send` command, we can use `-c` option, if you run
the command in the following way:

```bash
$ skycoin-cli send $RECIPIENT_ADDRESS $AMOUNT -c $CHANGE_ADDRESS
```

The change coins won't go to the address as you wish, it will go to the
default `change address`, which can be by `from` address or the wallet's
coinbase address.

The right script should look like this:

```bash
$ skycoin-cli send -c $CHANGE_ADDRESS $RECIPIENT_ADDRESS $AMOUNT
```

### Wipe device

Ask the device to generate a mnemonic and configure itself with it.

```bash
$skycoin-cli deviceWipe
```

#### Examples
##### Text output

```bash
$skycoin-cli deviceWipe
```
<details>
 <summary>View Output</summary>

```
2018/11/06 16:00:28 Wipe device 26! Answer is: 0806
2018/11/06 16:00:31 MessageButtonAck Answer is: 2 /

Device wiped
```
</details>


### Backup device

Ask the device to perform the seed backup procedure.

```bash
$skycoin-cli deviceBackup
```

#### Examples
##### Text output

```bash
$skycoin-cli deviceBackup
```
<details>
 <summary>View Output</summary>

```
2018/11/15 17:13:40 Backup device 26! Answer is:
2018/11/15 17:14:58 Success 2! Answer is: Seed successfully backed up
```
</details>


### Recovery device

Ask the device to perform the seed recovery procedure.

```bash
$skycoin-cli deviceRecovery
```

#### Examples
##### Text output

```bash
$skycoin-cli deviceRecovery
```
<details>
 <summary>View Output</summary>

```
2018/12/07 17:50:26 Recovery device 46! Answer is: 
Word: market
Word: gaze
Word: crouch
Word: enforce
Word: green
Word: art
Word: stem
Word: infant
Word: host
Word: metal
Word: flower
Word: cup
Word: exit
Word: thank
Word: upset
Word: cloud
Word: below
Word: body
Word: remain
Word: vocal
Word: team
Word: discover
Word: core
Word: abuse
Failed with code:  The seed is valid but does not match the one in the device
```
</details>