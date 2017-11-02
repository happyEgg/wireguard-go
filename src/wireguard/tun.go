package wireguard

import (
	"sync/atomic"
)

const DefaultMTU = 1420

type TUNEvent int

const (
	TUNEventUp = 1 << iota
	TUNEventDown
	TUNEventMTUUpdate
)

type TUNDevice interface {
	Read([]byte) (int, error)  // read a packet from the device (without any additional headers)
	Write([]byte) (int, error) // writes a packet to the device (without any additional headers)
	MTU() (int, error)         // returns the MTU of the device
	Name() string              // returns the current name
	Events() chan TUNEvent     // returns a constant channel of events related to the device
	Close() error              // stops the device and closes the event channel
}

func (device *Device) RoutineTUNEventReader() {
	logInfo := device.Log.Info
	logError := device.Log.Error

	for event := range device.tun.device.Events() {
		if event&TUNEventMTUUpdate != 0 {
			mtu, err := device.tun.device.MTU()
			old := atomic.LoadInt32(&device.tun.mtu)
			if err != nil {
				logError.Println("Failed to load updated MTU of device:", err)
			} else if int(old) != mtu {
				if mtu+MessageTransportSize > MaxMessageSize {
					logInfo.Println("MTU updated:", mtu, "(too large)")
				} else {
					logInfo.Println("MTU updated:", mtu)
				}
				atomic.StoreInt32(&device.tun.mtu, int32(mtu))
			}
		}

		if event&TUNEventUp != 0 {
			if !device.tun.isUp.Get() {
				logInfo.Println("Interface set up")
				device.tun.isUp.Set(true)
				updateUDPConn(device)
			}
		}

		if event&TUNEventDown != 0 {
			if device.tun.isUp.Get() {
				logInfo.Println("Interface set down")
				device.tun.isUp.Set(false)
				closeUDPConn(device)
			}
		}
	}
}
