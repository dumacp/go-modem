package nmea

import (
	"errors"
	"fmt"
	"time"

	"github.com/dumacp/go-modem/appliance/business/messages"
	"github.com/dumacp/go-modem/appliance/crosscutting/logs"
	"github.com/dumacp/gpsnmea"
	"github.com/looplab/fsm"
)

const (
	sStart   = "sStart"
	sConnect = "sConnect"
	sRun     = "sRun"
	sReset   = "sReset"
	sStop    = "sStop"
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

func (act *actornmea) initFSM() {
	act.fsm = fsm.NewFSM(
		sStart,
		fsm.Events{
			{Name: startEvent, Src: []string{sStart, sStop}, Dst: sConnect},
			{Name: connectOKEvent, Src: []string{sConnect}, Dst: sRun},
			{Name: connectFailEvent, Src: []string{sConnect}, Dst: sReset},
			{Name: timeoutEvent, Src: []string{sConnect}, Dst: sConnect},
			{Name: readFailEvent, Src: []string{sRun}, Dst: sReset},
			{Name: readStopEvent, Src: []string{sStart, sRun, sConnect}, Dst: sStop},
			{Name: resetEvent, Src: []string{sReset}, Dst: sConnect},
		},
		fsm.Callbacks{
			"enter_state": func(e *fsm.Event) {
				logs.LogBuild.Printf("FSM NMEA state Src: %v, state Dst: %v", e.Src, e.Dst)
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

func (act *actornmea) startfsm(chQuit chan int) {
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
					errx = errors.New("Unknown panic")
				}
			}
		}()
		countFail := 0
		act.fsm.SetState(sStart)
		for {
			// log.Println(m.Current())
			switch act.fsm.Current() {
			case sStart:
				dev, err := gpsnmea.NewDevice(act.portNmea, act.baudRate, filter...)
				if err != nil {
					logs.LogWarn.Println(err)
					act.fsm.Event(connectFailEvent)
					time.Sleep(3 * time.Second)
					break
				}
				act.dev = dev
				act.fsm.Event(readStopEvent)
			case sConnect:
				act.dev.Close()
				if err := act.dev.Open(); err != nil {
					logs.LogWarn.Println(err)
					time.Sleep(5 * time.Second)
					if countFail > 2 {
						act.context.Send(act.context.Self(), &msgStop{})
						act.fsm.Event(readStopEvent)
					} else {
						countFail++
						act.fsm.Event(connectFailEvent)
					}
					break
				}
				act.fsm.Event(connectOKEvent)
			case sReset:
				act.context.Request(act.modemPID, &messages.ModemReset{})
				time.Sleep(30 * time.Second)
				act.fsm.Event(resetEvent)
			case sRun:
				funcRun := func() {
					if err := act.run(chQuit, act.timeout, act.distanceMin); err != nil {
						logs.LogError.Println(err)
						select {
						case <-chQuit:
						default:
						}
						act.fsm.Event(readFailEvent)
						return
					}
					act.context.Send(act.context.Self(), &msgStop{})
					act.fsm.Event(readStopEvent)
				}
				funcRun()
				logs.LogWarn.Println("stop run function in nmea")
			case sStop:
				time.Sleep(3 * time.Second)
			default:
				time.Sleep(3 * time.Second)
			}
		}
	}
	go func() {
		for {
			if err := funcRutine(); err != nil {
				act.context.Send(act.context.Self(), &msgFatal{err: err})
			}
		}
	}()
}
