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
	act.behavior.Become(act.Wait)
	act.state = WaitState
	act.chQuit = make(chan int, 0)
	act.chQuitTick = make(chan int, 0)
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
		select {
		case act.chQuit <- 1:
		case <-time.After(10 * time.Second):
			//close(act.chQuit)
			//act.chQuit = make(chan int, 0)
		}
		logs.LogWarn.Printf("nmea read modemReset \"%s\"", ctx.Self().Id)
		//time.Sleep(3 * time.Second)
		//panic(fmt.Errorf("modemReset arrive in nmea"))
		//ctx.Send(ctx.Self(), &msgStart{})
	case *msgFatal:
		logs.LogError.Println(msg.err)
		act.behavior.Become(act.Wait)
		act.state = WaitState
		select {
		case _, ok := <-act.chQuit:
			if ok {
				close(act.chQuit)
			}
		default:
			close(act.chQuit)
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
			logs.LogError.Panic(err)
		}
		act.pubsubPID = pidPubSub
		ctx.Watch(pidPubSub)
		go act.checkModem(act.chQuitTick)
	case *actor.Stopping:
		logs.LogInfo.Printf("actor stopping \"%s\"", ctx.Self().Id)
	case *messages.ModemOnResponse:
		logs.LogWarn.Printf("nmea modemOnResponse \"%s\"", msg)
		if msg.State {
			select {
			case _, ok := <-act.chQuit:
				if !ok {
					logs.LogBuild.Println("act.chQuit is closed")
					act.chQuit = make(chan int, 0)
				}
			default:
				select {
				case act.chQuit <- 1:
				case <-time.After(3 * time.Second):
				}
			}
			act.startfsm(act.chQuit)
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
