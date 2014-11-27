package i2c

import (
	"fmt"
	"syscall"
	"unsafe"
)

const (
	I2C_SLAVE       = 0x0703 /* Use this slave address */
	I2C_SLAVE_FORCE = 0x0706
	I2C_SMBUS       = 0x0720 /* SMBus transfer */
)

/* i2c_smbus_xfer read or write markers */
const (
	I2C_SMBUS_READ  = 1
	I2C_SMBUS_WRITE = 0
)

/* SMBus transaction types (size parameter in the above functions)
   Note: these no longer correspond to the (arbitrary) PIIX4 internal codes! */
const (
	I2C_SMBUS_QUICK            = 0
	I2C_SMBUS_BYTE             = 1
	I2C_SMBUS_BYTE_DATA        = 2
	I2C_SMBUS_WORD_DATA        = 3
	I2C_SMBUS_PROC_CALL        = 4
	I2C_SMBUS_BLOCK_DATA       = 5
	I2C_SMBUS_I2C_BLOCK_BROKEN = 6
	I2C_SMBUS_BLOCK_PROC_CALL  = 7 /* SMBus 2.0 */
	I2C_SMBUS_I2C_BLOCK_DATA   = 8
)

/*
func smbusWriteByteData(device *Device, command uint8, value uint8) error {
	data := &[4]uint8{value}
	return smbusAccess(device,
		I2C_SMBUS_WRITE, value,
		I2C_SMBUS_BYTE_DATA, uintptr(unsafe.Pointer(data)))
}
*/

type Device struct {
	fd int
}

func (device *Device) Close() error {
	return syscall.Close(device.fd)
}

func ioctl(a1, a2, a3 uintptr) (r1, r2 uintptr, err error) {
	err = nil
	r1, r2, errno := syscall.Syscall(syscall.SYS_IOCTL, a1, a2, a3)
	if errno != 0 {
		err = errno
	}
	return
}

type smbusIoctlData struct {
	readWrite uint8
	command   uint8
	size      uint32
	data      uintptr
}

func (device *Device) smbusAccess(readWrite uint8, command uint8, size int, data uintptr) error {
	ioctlData := &smbusIoctlData{readWrite, command, uint32(size), data}

	_, _, err := ioctl(
		uintptr(device.fd),
		uintptr(I2C_SMBUS),
		uintptr(unsafe.Pointer(ioctlData)))
	return err
}

func (device *Device) setSlaveAddr(addr int) error {
	_, _, err := ioctl(uintptr(device.fd), uintptr(I2C_SLAVE), uintptr(addr))
	return err
}

func (device *Device) WriteByteData(slave int, command uint8, value uint8) error {
	if err := device.setSlaveAddr(slave); err != nil {
		return err
	}
	data := [4]uint8{value, 0, 0, 0}
	dataPtr := uintptr(unsafe.Pointer(&data[0]))
	if err := device.smbusAccess(I2C_SMBUS_WRITE, command,
		I2C_SMBUS_BYTE_DATA, dataPtr); err != nil {
		return err
	}
	return nil
}

func open(path string) (device *Device, err error) {
	fd, err := syscall.Open(path, syscall.O_RDWR, 0x644)
	if err != nil {
		return nil, err
	}
	device = &Device{fd}
	return
}

func Open(busno int) (device *Device, err error) {
	// first try
	device, err = open(fmt.Sprintf("/dev/i2c/%d", busno))
	if err != nil {
		// second try
		device, err = open(fmt.Sprintf("/dev/i2c-%d", busno))
	}
	return
}
