package device

import (
	"bufio"
	"errors"
	"fmt"
	"time"

	"github.com/dumacp/go-logs/pkg/logs"
	"github.com/dumacp/go-modem/internal/nmea/process"
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

func initFSM() *fsm.FSM {
	f := fsm.NewFSM(
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
	return f
}

func (a *actornmea) startfsm(chQuit chan int) {
	// log.Println(m.Current())
	funcRutine := func() (errx error) {

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
		}()

		current := ""

		portReadTimeout := 3 * time.Second
		var port *serial.Port

		var reader *bufio.Reader
		countFail := 0
		countEmpty := 0
		a.fsm.SetState(sStart)
		for {
			// log.Println(m.Current())
			select {

			case <-chQuit:
				return errors.New("close chQuit")
			default:
				if current != a.fsm.Current() {
					logs.LogInfo.Printf("current state NMEA: %v", a.fsm.Current())
					current = a.fsm.Current()
				}
				switch a.fsm.Current() {

				case sStart:
					a.fsm.Event(startEvent)

				case sConnect:
					time.Sleep(1 * time.Second)
					config := &serial.Config{
						Name:        a.portNmea,
						Baud:        a.baudRate,
						ReadTimeout: portReadTimeout,
					}
					if port != nil {
						port.Close()
					}
					var err error
					port, err = serial.OpenPort(config)
					if err != nil {
						logs.LogError.Printf("nmea serial error open: %s", err)
						time.Sleep(3 * time.Second)
						break
					}
					reader = bufio.NewReader(port)

					time.Sleep(1 * time.Second)
					a.fsm.Event(connectOKEvent)
				case sReset:
					a.context.Send(a.context.Self(), &msgFatal{err: errors.New("many errors")})
					time.Sleep(3 * time.Second)
					a.fsm.Event(resetEvent)
				case sRun:
					data, err := listen(reader)
					if err != nil {

						countFail++
						if countFail > 6 {
							logs.LogWarn.Printf("error listen port: %s", err)
							a.fsm.Event(connectFailEvent)
						}
						break
					}
					if len(data) <= 0 {
						countEmpty++
						if countEmpty > 120 {
							a.fsm.Event(connectFailEvent)
						}
						break
					}
					countEmpty = 0
					countFail = 0
					if a.processPID != nil {
						a.context.Send(a.processPID, &process.MsgData{Data: data})
					}
				case sClose:
					a.fsm.Event(startEvent)
				case sStop:
					countFail++
					if countFail > 60 {
						a.fsm.Event(startEvent)
						countFail = 0
					}
				default:
					time.Sleep(1 * time.Second)
				}
				time.Sleep(30 * time.Millisecond)
			}
		}
	}
	go func() {
		if err := funcRutine(); err != nil {
			a.context.Send(a.context.Self(), &msgFatal{err: err})
		}
	}()
}
