package device

import "errors"

var errReset = errors.New("nmea reset error")
var errNmea = errors.New("nmea error")

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

type MsgSubscribeModem struct {
}

type MsgSubscribeProcess struct {
}
