package control

import (
	"github.com/dumacp/omvz/sierra"
)

const (
	portModem = "/dev/ttyUSB2"
	baudModem = 115200
)

// NewModem function to connect with sierra modem
func newModem(port string) sierra.SierraModem {
	opts := &sierra.PortOptions{
		Port: port,
		Baud: baudModem,
	}
	modSierra := sierra.NewModem(opts)
	return modSierra
}

func verifyModem(m sierra.SierraModem) int {
	if !m.Open() {
		return 1
	}
	defer m.Close()
	if !m.Verify() {
		return 2
	}
	return 0
}

func verifySIM(m sierra.SierraModem) bool {
	if !m.Open() {
		return false
	}
	defer m.Close()
	return m.IsSimOK()
}

func verifyWANT(m sierra.SierraModem) bool {
	if !m.Open() {
		return false
	}
	defer m.Close()
	return m.IsWANT()
}

func setWANT(m sierra.SierraModem) bool {
	if !m.Open() {
		return false
	}
	defer m.Close()
	return m.SetWANT()
}

func powerOffModem(m sierra.SierraModem) bool {
	return m.PowerOff()
}

func powerOnModem(m sierra.SierraModem) bool {
	return m.PowerOn()
}

func resetModem(m sierra.SierraModem) bool {
	return m.ResetHw()
}

func resetSWModem(m sierra.SierraModem) bool {
	return m.ResetSw()
}

func resetUSBHost(m sierra.SierraModem) bool {
	return m.ResetUSBHost()
}

func reConnect(m sierra.SierraModem, apn ...string) bool {
	if !m.Open() {
		return false
	}
	defer m.Close()

	if len(apn) <= 0 {
		if ok, _ := m.ConnectToApn(""); ok {
			return true
		}
		return false
	}

	for _, apn_ := range apn {
		if ok, _ := m.ConnectToApn(apn_); ok {
			return true
		}
	}
	return false
}
