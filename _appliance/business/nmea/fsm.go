package nmea

import (
	"bufio"
	"errors"
	"fmt"
	"time"

	"github.com/dumacp/go-logs/pkg/logs"
	"github.com/dumacp/go-modem/appliance/business/messages"
	"github.com/looplab/fsm"
	"github.com/tarm/serial"
)

const (
	sStart   = "sStart"
	sConnect = "sConnect"
	sRun     = "sRun"
	sReset   = "sReset"
	sStop    = "sStop"
	sClose   = "sClose"
)

const (
	startEvent       = "startEvent"
	readFailEvent    = "readFailEvent"
	readStopEvent    = "readStopEvent"
	connectOKEvent   = "connectOKEvent"
	connectFailEvent = "connectFailEvent"
	timeoutEvent     = "tomeoutEvent"
	resetEvent       = "resetEvent"
)

func beforeEvent(event string) string {
	return fmt.Sprintf("before_%s", event)
}
func enterState(state string) string {
	return fmt.Sprintf("enter_%s", state)
}
func leaveState(state string) string {
	return fmt.Sprintf("leave_%s", state)
}

func (a *actornmea) initFSM() {
	a.fsm = fsm.NewFSM(
		sStart,
		fsm.Events{
			{Name: startEvent, Src: []string{sStart, sStop, sClose}, Dst: sConnect},
			{Name: connectOKEvent, Src: []string{sConnect}, Dst: sRun},
			{Name: connectFailEvent, Src: []string{sConnect, sRun}, Dst: sReset},
			{Name: timeoutEvent, Src: []string{sConnect}, Dst: sConnect},
			{Name: readFailEvent, Src: []string{sRun}, Dst: sClose},
			{Name: readStopEvent, Src: []string{sStart, sRun, sConnect}, Dst: sStop},
			{Name: resetEvent, Src: []string{sReset}, Dst: sStop},
		},
		fsm.Callbacks{
			"enter_state": func(e *fsm.Event) {
				fmt.Printf("FSM NMEA state Src: %v, state Dst: %v\n", e.Src, e.Dst)
			},
			"leave_state": func(e *fsm.Event) {
				if e.Err != nil {
					e.Cancel(e.Err)
				}
			},
			"before_event": func(e *fsm.Event) {
				if e.Err != nil {
					e.Cancel(e.Err)
				}
			},
		},
	)
}

func (a *actornmea) startfsm(chQuit chan int) {
	// log.Println(m.Current())
	funcRutine := func() (errx error) {

		quit := make(chan int)
		chErr := make(chan error)
		chData := make(chan string)

		defer func() {
			if r := recover(); r != nil {
				logs.LogError.Println("Recovered in \"startfsm()\", ", r)
				switch x := r.(type) {
				case string:
					errx = errors.New(x)
				case error:
					errx = x
				default:
					errx = errors.New("unknown panic")
				}
			}
			close(quit)
		}()

		queue := NewQueue()
		lastPoint := []float64{0.0, 0.0}
		portReadTimeout := 10 * time.Second
		var port *serial.Port

		var reader *bufio.Reader
		countFail := 0
		a.fsm.SetState(sStart)
		for {
			// log.Println(m.Current())
			switch a.fsm.Current() {
			case sStart:
				// dev, err := gpsnmea.NewDevice(a.portNmea, a.baudRate, filter...)
				// if err != nil {
				// 	logs.LogWarn.Println(err)
				// 	a.fsm.Event(connectFailEvent)
				// 	time.Sleep(3 * time.Second)
				// 	break
				// }
				// a.dev = dev
				// a.fsm.Event(readStopEvent)
			case sConnect:
				time.Sleep(1 * time.Second)
				config := &serial.Config{
					Name:        a.portNmea,
					Baud:        a.baudRate,
					ReadTimeout: portReadTimeout,
				}
				succ := false
				for range []int{0, 1, 2} {
					if port != nil {
						port.Close()
					}
					var err error
					port, err = serial.OpenPort(config)
					if err != nil {
						logs.LogError.Printf("QR error open: %s", err)
						time.Sleep(3 * time.Second)
						continue
					}
					succ = true
				}
				if !succ {
					break
				}
				reader = bufio.NewReader(port)

				select {
				case _, ok := <-quit:
					if ok {
						close(quit)
					}
				default:
					close(quit)
				}
				time.Sleep(1 * time.Second)
				quit = make(chan int)

				go listen(reader, quit, chErr, chData)
				countFail = 0
				a.fsm.Event(connectOKEvent)
			case sReset:
				a.context.Request(a.modemPID, &messages.ModemReset{})
				time.Sleep(3 * time.Second)
				a.fsm.Event(resetEvent)
			case sRun:
				select {
				case data := <-chData:
					pData, err := processData(data, a.distanceMin, lastPoint, queue)
					if err != nil {

						if countFail > 10 {
							logs.LogWarn.Println(err)
							return err
						}
						countFail += 1
						break
					}
					a.context.Send(a.pubsubPID, &msgGPS{data: pData.raw})
					lastPoint[0] = pData.lat
					lastPoint[1] = pData.lgt
				case <-chErr:
				case <-chQuit:
					return nil
				}
			case sClose:
				if a.dev != nil {
					a.dev.Close()
				}
				a.fsm.Event(startEvent)
			case sStop:
				countFail++
				if countFail > 30 {
					a.fsm.Event(startEvent)
					countFail = 0
				}
			default:
				time.Sleep(3 * time.Second)
			}
			time.Sleep(1 * time.Second)
		}
	}
	go func() {
		for {
			if err := funcRutine(); err != nil {
				a.context.Send(a.context.Self(), &msgFatal{err: err})
			}
		}
	}()
}
