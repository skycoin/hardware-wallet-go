# CLI Documentation

Skycoin command line interface

The CLI command APIs can be used directly from a Go application, see [Skycoin CLI Godoc](https://godoc.org/github.com/skycoin/skycoin/src/cli).

<!-- MarkdownTOC autolink="true" bracket="round" levels="1,2,3" -->

- [Usage](#usage)
    - [Update firmware](#update-firmware)
    - [Ask device to generate addresses](#ask-device-to-generate-addresses)
    - [Configure device mnemonic](#configure-device-mnemonic)
    - [Configure device PIN code](#configure-device-pin-code)
    - [Ask device to sign message](#ask-device-to-sign-message)
    - [Ask device to check signature](#ask-device-to-check-signature)
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
     deviceAddressGen               Generate skycoin addresses using the firmware
     deviceFirmwareUpdate           Update device's firmware.
     deviceSignMessage              Ask the device to sign a message using the secret key at given index.
     deviceCheckMessageSignature    Check a message signature matches the given address.
     emulatorSetMnemonic            Configure an emulated device with a mnemonic.
     emulatorAddressGen             Generate skycoin addresses using an emulated device.
     emulatorSignMessage            Ask the emulated device to sign a message using the secret key at given index.
     emulatorCheckMessageSignature  Check a message signature matches the given address.
     help, h               Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h     show help
   --version, -v  print the version
ENVIRONMENT VARIABLES:
    RPC_ADDR: Address of RPC node. Must be in scheme://host format. Default "http://127.0.0.1:6420"
    COIN: Name of the coin. Default "skycoin"
    USE_CSRF: Set to 1 or true if the remote node has CSRF enabled. Default false (unset)
    WALLET_DIR: Directory where wallets are stored. This value is overriden by any subcommand flag specifying a wallet filename, if that filename includes a path. Default "$HOME/.$COIN/wallets"
    WALLET_NAME: Name of wallet file (without path). This value is overriden by any subcommand flag specifying a wallet filename. Default "$COIN_cli.wlt"
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
