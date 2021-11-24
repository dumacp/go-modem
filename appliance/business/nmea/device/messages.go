package device

type msgFatal struct {
	err error
}

type msgStop struct{}
type msgStart struct{}

type msgGPS struct {
	data string
}

type eventGPS struct {
	event string
}

type msgBadGPS struct {
	data string
}

type StopNmea struct{}

type AddressModem struct {
	Addr string
	ID   string
}

type AddressPubSub struct {
	Addr string
	ID   string
}
