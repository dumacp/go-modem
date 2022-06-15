package process

import (
	"fmt"
	"regexp"
	"time"

	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/dumacp/go-logs/pkg/logs"
	"github.com/dumacp/go-modem/internal/pubsub"
)

const (
	RunState int = iota
	WaitState
)

const (
	topicGPS      = "GPS"
	topicEventGPS = "EVENTS/gps"
	topicBadGPS   = "EVENTS/badgps"
)

const timeoutBadFrames = 10 * time.Minute

var (
	re = regexp.MustCompile(`\$[a-zA-Z]+,`)
)

type actorprocess struct {
	context            actor.Context
	timeout            time.Duration
	distanceMin        int
	lastSendFrames     map[string]*processedData
	lastFrames         map[string]*processedData
	lastBad            string
	queue              *queue
	countInvalidFrames int
	quit               chan int
}

func NewActor(timeout, distanceMin int) actor.Actor {
	act := &actorprocess{}
	act.distanceMin = distanceMin
	act.timeout = time.Duration(timeout) * time.Second
	return act
}

func (a *actorprocess) Receive(ctx actor.Context) {
	a.context = ctx
	switch msg := ctx.Message().(type) {
	case *actor.Started:
		logs.LogInfo.Printf("actor started \"%s\"", ctx.Self().Id)
		a.lastFrames = make(map[string]*processedData)
		a.lastSendFrames = make(map[string]*processedData)
		a.queue = NewQueue()
		if a.quit != nil {
			select {
			case _, ok := <-a.quit:
				if ok {
					close(a.quit)
				}
			default:
				close(a.quit)
			}
		}
		a.quit = make(chan int)
		go tick(ctx, timeoutBadFrames, a.quit)
	case *MsgData:
		if err := func() error {
			data := msg.Data
			gtype := re.FindString(data)
			gtype = gtype[:len(gtype)-1]
			if len(gtype) <= 0 {
				return fmt.Errorf("invalid frame \"%s\"", data)
			}
			if gtype != GPGGA.Value() && gtype != GPRMC.Value() {
				return nil
			}
			if len(data) < 34 {
				return fmt.Errorf("invalid frame %q", data)
			}
			fmt.Println(msg.Data)

			var lastPoint []float64
			if v, ok := a.lastSendFrames[gtype]; ok {
				lastPoint = []float64{v.lat, v.lgt}
			} else {
				lastPoint = []float64{0.0, 0.0}
			}
			result, err := processData(data, gtype, a.distanceMin, lastPoint, a.queue)
			if err != nil {
				return err
			}
			defer func() {
				a.lastFrames[gtype] = result
			}()
			if !result.valided {
				return fmt.Errorf("invalid frame in process: %q", result.raw)
			}
			if gtype == GPRMC.Value() {
				ctx.Send(ctx.Self(), &MsgSendFrame{
					Topic: topicGPS,
					Data:  []byte(result.raw),
				})
				// fmt.Println("send_0")
			}

			funcSend := func(sendData *processedData) {
				if sendData.distance*1000 > float64(a.distanceMin) {
					a.lastSendFrames[gtype] = result
					tn := time.Now()
					timeStamp := float64(tn.UnixNano()) / 1000000000
					event := []byte(fmt.Sprintf("{\"timeStamp\": %f, \"value\": %q, \"type\": %q}",
						timeStamp, result.raw, result.prefix.Value()[1:]))
					ctx.Send(ctx.Self(), &MsgSendFrame{
						Topic: topicEventGPS,
						Data:  event,
					})
					// fmt.Println("send_1")
					return
				}
				lastTime := time.Now().Add(-30 * time.Minute)
				if last, ok := a.lastSendFrames[gtype]; ok {
					lastTime = last.timeStamp
				}
				if time.Since(lastTime) >= a.timeout {
					a.lastSendFrames[gtype] = result
					tn := time.Now()
					timeStamp := float64(tn.UnixNano()) / 1000000000
					event := []byte(fmt.Sprintf("{\"timeStamp\": %f, \"value\": %q, \"type\": %q}",
						timeStamp, result.raw, result.prefix.Value()[1:]))
					ctx.Send(ctx.Self(), &MsgSendFrame{
						Topic: topicEventGPS,
						Data:  event,
					})
					// fmt.Println("send_2")
					return
				}
			}

			if gtype != GPGGA.Value() {
				if v1, ok := a.lastFrames[GPGGA.Value()]; ok && result.timeDate.Sub(v1.timeDate) < 60*time.Second {
					if v2, ok := a.lastSendFrames[GPGGA.Value()]; ok {
						if tv2 := result.timeDate.Sub(v2.timeDate); tv2 >= 0 && tv2 < 1*time.Second {
							a.lastSendFrames[gtype] = result
							tn := time.Now()
							timeStamp := float64(tn.UnixNano()) / 1000000000
							event := []byte(fmt.Sprintf("{\"timeStamp\": %f, \"value\": %q, \"type\": %q}",
								timeStamp, result.raw, result.prefix.Value()[1:]))
							ctx.Send(ctx.Self(), &MsgSendFrame{
								Topic: topicEventGPS,
								Data:  event,
							})
							// fmt.Println("send_3")
							return nil
						}
					}
					return nil
				} else {
					funcSend(result)
				}
			} else {
				funcSend(result)
				for _, v := range a.lastFrames {
					if tv := result.timeDate.Sub(v.timeDate); tv >= 0 && tv < 1*time.Second {
						a.lastSendFrames[gtype] = v
						tn := time.Now()
						timeStamp := float64(tn.UnixNano()) / 1000000000
						event := []byte(fmt.Sprintf("{\"timeStamp\": %f, \"value\": %q, \"type\": %q}",
							timeStamp, v.raw, v.prefix.Value()[1:]))
						ctx.Send(ctx.Self(), &MsgSendFrame{
							Topic: topicEventGPS,
							Data:  event,
						})
						// fmt.Println("send_4")
					}
				}
			}
			return nil
		}(); err != nil {
			fmt.Println(err)
			a.countInvalidFrames++
			if a.countInvalidFrames%5 == 0 {
				logs.LogWarn.Println(err)
			}
			a.lastBad = msg.Data
		}
	case *MsgTick:
		rateBad := float64(a.countInvalidFrames) / timeoutBadFrames.Minutes()
		badgps := fmt.Sprintf("{\"timeStamp\": %d, \"value\": %.2f, \"type\": %q}", time.Now().Unix(), rateBad, "GPSERROR")
		a.countInvalidFrames = 0
		fmt.Printf("last bad GPS frame -> %q\n", a.lastBad)
		fmt.Printf("rate bad GPS -> %.2f\n", rateBad)
		pubsub.Publish(topicBadGPS, []byte(badgps))
	case *MsgSendFrame:
		// tn := time.Now()
		// timeStamp := float64(tn.UnixNano()) / 1000000000

		// event := []byte(fmt.Sprintf("{\"timeStamp\": %f, \"value\": %q, \"type\": %q}",
		// 	timeStamp, msg.Data.raw, msg.Data.prefix.Value()[1:]))
		// fmt.Printf("publish: %s, topic: %s\n", msg.Data, msg.Topic)
		pubsub.Publish(msg.Topic, msg.Data)
	case *actor.Stopping:
		logs.LogInfo.Printf("actor stopping \"%s\"", ctx.Self().Id)
		if a.quit != nil {
			select {
			case _, ok := <-a.quit:
				if ok {
					close(a.quit)
				}
			default:
				close(a.quit)
			}
		}
	}
}

func tick(ctx actor.Context, timeout time.Duration, quit <-chan int) {
	rootctx := ctx.ActorSystem().Root
	self := ctx.Self()
	t1 := time.NewTicker(timeout)
	defer t1.Stop()
	for {
		select {
		case <-t1.C:
			rootctx.Send(self, &MsgTick{})
		case <-quit:
			return
		}
	}
}
