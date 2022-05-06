package process

import (
	"fmt"
	"regexp"
	"time"

	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/dumacp/go-logs/pkg/logs"
	"github.com/dumacp/go-modem/appliance/crosscutting/comm/pubsub"
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

var (
	re = regexp.MustCompile(`\$[a-zA-Z]+,`)
)

type actorprocess struct {
	context          actor.Context
	timeout          time.Duration
	distanceMin      int
	lastSendFrames   map[string]*processedData
	lastFrames       map[string]*processedData
	queue            *queue
	countValidFrames int
}

func NewActor(timeout, distanceMin int) actor.Actor {
	//initLogs(debug)
	act := &actorprocess{}

	return act
}

func (a *actorprocess) Receive(ctx actor.Context) {
	a.context = ctx
	switch msg := ctx.Message().(type) {
	case *actor.Started:
		logs.LogInfo.Printf("actor started \"%s\"", ctx.Self().Id)
		a.lastFrames = make(map[string]*processedData)
	case *MsgData:

		if err := func() error {

			data := msg.Data
			gtype := re.FindString(data)
			gtype = gtype[:len(gtype)-1]
			if len(gtype) <= 0 {
				return fmt.Errorf("invalid frame \"%s\"", data)
			}

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
			tn := time.Now()
			timeStamp := float64(tn.UnixNano()) / 1000000000
			if !result.valided {
				return fmt.Errorf("invalid frame in process: %q", result.raw)
			}
			if gtype != GPGGA.Value() {
				if v, ok := a.lastFrames[GPGGA.Value()]; ok {
					if !v.valided {
						return fmt.Errorf("frame %q, lastframe $GPGGA is invalid: %q", data, v.raw)
					}
				}
			}
			if result.distance*0.90 > float64(a.distanceMin) {
				a.lastSendFrames[gtype] = result
				event := []byte(fmt.Sprintf("{\"timeStamp\": %f, \"value\": %q, \"type\": %q}",
					timeStamp, result.raw, result.prefix.Value()[1:]))
				pubsub.Publish(topicEventGPS, event)
				return nil
			}
			if time.Since(result.timeStamp) >= a.timeout {
				a.lastSendFrames[gtype] = result
				event := []byte(fmt.Sprintf("{\"timeStamp\": %f, \"value\": %q, \"type\": %q}",
					timeStamp, result.raw, result.prefix.Value()[1:]))
				pubsub.Publish(topicEventGPS, event)
				return nil
			}
			return nil
		}(); err != nil {
			logs.LogWarn.Println(err)
			fmt.Println(err)
			if a.countValidFrames > 30 {
				if ctx.Sender() != nil {
					ctx.Respond(&MsgInvalidFrames{})
				}
			} else {
				a.countValidFrames += 1
			}
		}

	case *actor.Stopping:

		logs.LogInfo.Printf("actor stopping \"%s\"", ctx.Self().Id)
	}
}
