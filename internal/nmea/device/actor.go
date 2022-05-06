package device

import (
	"time"

	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/dumacp/go-logs/pkg/logs"
	"github.com/dumacp/go-modem/pkg/messages"
	"github.com/looplab/fsm"
)

type actornmea struct {
	context    actor.Context
	fsm        *fsm.FSM
	portNmea   string
	baudRate   int
	modemPID   *actor.PID
	processPID *actor.PID
	chQuit     chan int
}

func NewNmeaActor(portNmea string, baudRate int) actor.Actor {
	//initLogs(debug)
	act := &actornmea{}
	act.portNmea = portNmea
	act.baudRate = baudRate
	act.fsm = initFSM()

	return act
}

func (act *actornmea) Receive(ctx actor.Context) {
	act.context = ctx
	switch msg := ctx.Message().(type) {
	case *actor.Started:
		logs.LogInfo.Printf("actor started \"%s\"", ctx.Self().Id)

		if act.chQuit != nil {
			select {
			case _, ok := <-act.chQuit:
				if ok {
					close(act.chQuit)
				}
			default:
				close(act.chQuit)
			}
			time.Sleep(300 * time.Millisecond)
		}
		act.chQuit = make(chan int)
		go act.checkModem(act.chQuit)
		act.startfsm(act.chQuit)
	case *actor.Stopping:
		logs.LogInfo.Printf("actor stopping \"%s\"", ctx.Self().Id)
		if act.chQuit != nil {
			select {
			case _, ok := <-act.chQuit:
				if ok {
					close(act.chQuit)
				}
			default:
				close(act.chQuit)
			}
			time.Sleep(300 * time.Millisecond)
		}
	case *messages.ModemReset:
		logs.LogWarn.Printf("nmea msg modemReset")
		act.fsm.Event(readStopEvent)
	case *messages.ModemOnResponse:
		logs.LogWarn.Printf("nmea modemOnResponse \"%s\"", msg)
		if msg.State {
			logs.LogInfo.Println("startfsm nmea")
			act.fsm.Event(startEvent)
		}
	case *msgFatal:
		logs.LogError.Printf("nmead read failed: %s", msg.err)
		if act.modemPID != nil {
			act.context.Request(act.modemPID, &messages.ModemOnRequest{})
		}
	case *MsgSubscribeModem:
		if ctx.Sender() != nil {
			act.modemPID = ctx.Sender()
		}
	case *MsgSubscribeProcess:
		if ctx.Sender() != nil {
			act.processPID = ctx.Sender()
		}
	case *actor.Terminated:
		logs.LogError.Printf("actor terminated: %s", msg.Who.GetId())
	}
}

func (act *actornmea) checkModem(chQuit chan int) {
	tick := time.NewTicker(10 * time.Second)
	defer tick.Stop()
	for range tick.C {
		if act.fsm.Current() == sRun {
			continue
		}
		if act.modemPID != nil && act.fsm.Current() == sStop {
			act.context.Request(act.modemPID, &messages.ModemOnRequest{})
		} else if act.modemPID == nil {
			logs.LogWarn.Println("act.modemPID is empty")
		}
	}
}
