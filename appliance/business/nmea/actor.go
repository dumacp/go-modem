package nmea

import (
	"time"

	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/dumacp/go-modem/appliance/business/messages"
	"github.com/dumacp/go-modem/appliance/crosscutting/logs"
	"github.com/dumacp/gpsnmea"
	"github.com/looplab/fsm"
)

const (
	RunState int = iota
	WaitState
)

type actornmea struct {
	debug       bool
	context     actor.Context
	behavior    actor.Behavior
	state       int
	fsm         *fsm.FSM
	dev         *gpsnmea.Device
	portNmea    string
	timeout     int
	distanceMin int
	baudRate    int
	modemPID    *actor.PID
	pubsubPID   *actor.PID
	chQuit      chan int
	chQuitTick  chan int
	chTick      *time.Timer
}

func NewNmeaActor(debug bool, portNmea string, baudRate, timeout, distanceMin int) actor.Actor {
	//initLogs(debug)
	act := &actornmea{}
	act.debug = debug
	act.portNmea = portNmea
	act.timeout = timeout
	act.baudRate = baudRate
	act.distanceMin = distanceMin
	act.initFSM()
	act.chQuit = make(chan int, 0)
	act.startfsm(act.chQuit)
	act.behavior.Become(act.Wait)
	act.state = WaitState
	act.chQuitTick = make(chan int, 0)
	go act.checkModem(act.chQuitTick)
	return act
}

func (act *actornmea) Receive(ctx actor.Context) {
	act.context = ctx
	act.behavior.Receive(ctx)
}

func (act *actornmea) Run(ctx actor.Context) {
	switch msg := ctx.Message().(type) {
	case *actor.Stopping:
		logs.LogInfo.Printf("actor stopping \"%s\"", ctx.Self().Id)
	case *msgStop:
		logs.LogWarn.Printf("nmea read stopped \"%s\"", ctx.Self().Id)
		act.behavior.Become(act.Wait)
		act.state = WaitState
		//time.Sleep(3 * time.Second)
		//panic(fmt.Errorf("msgStop arrive in nmea"))
	case *messages.ModemReset:
		logs.LogWarn.Printf("nmea msg modemReset")
		select {
		case act.chQuit <- 1:
			logs.LogWarn.Printf("stopping RUN nmea function")
		case <-time.After(30 * time.Second):
			logs.LogWarn.Printf("error stopping RUN nmea function")
			act.behavior.Become(act.Wait)
			act.state = WaitState
			// panic("error stopping RUN nmea function")
		}
		logs.LogWarn.Printf("nmea read modemReset \"%s\"", ctx.Self().Id)

		//time.Sleep(3 * time.Second)
		//panic(fmt.Errorf("modemReset arrive in nmea"))
		//ctx.Send(ctx.Self(), &msgStart{})
	case *msgFatal:
		logs.LogError.Printf("nmead read failed: %s", msg.err)
		act.behavior.Become(act.Wait)
		act.state = WaitState
		select {
		case act.chQuit <- 1:
			logs.LogWarn.Printf("stopping RUN nmea function")
		default:
		}
		//time.Sleep(3 * time.Second)
		//panic(msg.err)
	case *AddressModem:
		act.modemPID = actor.NewPID(msg.Addr, msg.ID)
	case *AddressPubSub:
		act.pubsubPID = actor.NewLocalPID(msg.ID)
	case *actor.Terminated:
		logs.LogError.Printf("actor terminated: %s", msg.Who.GetId())
	}
}

func (act *actornmea) Wait(ctx actor.Context) {
	switch msg := ctx.Message().(type) {
	case *actor.Started:
		logs.LogInfo.Printf("actor started \"%s\"", ctx.Self().Id)

		propsPubSub := actor.PropsFromFunc(NewPubSubActor(act.debug).Receive)
		pidPubSub, err := ctx.SpawnNamed(propsPubSub, "nmeaPubSub")
		if err != nil {
			time.Sleep(3 * time.Second)
			logs.LogError.Panic(err)
		}
		act.pubsubPID = pidPubSub
		ctx.Watch(pidPubSub)

	case *actor.Stopping:
		logs.LogInfo.Printf("actor stopping \"%s\"", ctx.Self().Id)
	case *messages.ModemOnResponse:
		logs.LogWarn.Printf("nmea modemOnResponse \"%s\"", msg)
		if msg.State {
			select {
			case _, ok := <-act.chQuit:
				if !ok {
					act.chQuit = make(chan int, 0)
					logs.LogWarn.Printf("error chQuit in RUN nmea is closed")
					panic("error chQuit in RUN nmea is closed")
				}
			default:
				select {
				case act.chQuit <- 1:
				case <-time.After(10 * time.Second):
				}
			}
			logs.LogInfo.Println("startfsm nmea")
			// act.startfsm(act.chQuit)
			act.fsm.Event(startEvent)
			act.behavior.Become(act.Run)
			act.state = RunState
		}
	case *AddressModem:
		act.modemPID = actor.NewPID(msg.Addr, msg.ID)
	case *AddressPubSub:
		act.pubsubPID = actor.NewLocalPID(msg.ID)
	case *actor.Terminated:
		logs.LogError.Printf("actor terminated: %s", msg.Who.GetId())
	}
}

func (act *actornmea) checkModem(chQuit chan int) {
	tick := time.NewTicker(10 * time.Second)
	defer tick.Stop()
	for range tick.C {
		if act.state == RunState {
			continue
		}
		select {
		case <-chQuit:
			return
		default:
		}
		if act.modemPID != nil {
			act.context.Request(act.modemPID, &messages.ModemOnRequest{})
		} else {
			logs.LogWarn.Println("act.modemPID is empty")
		}
	}
}
