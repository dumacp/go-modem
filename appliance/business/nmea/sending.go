package nmea

import (
	"fmt"
	"time"

	"github.com/AsynkronIT/protoactor-go/actor"
)

func sending(ctx *actor.RootContext, pid *actor.PID,
	distanceMin float64, timeout time.Duration, chData chan *processedData) {

	lastTime := make(map[string]time.Time)
	for v := range chData {
		tn := time.Now()
		timeStamp := float64(tn.UnixNano()) / 1000000000
		if v.valided {
			ctx.Send(pid, &msgGPS{data: v.raw})
		}
		lastStamp, ok := lastTime[v.prefix.Value()]
		if ok {
			switch {
			case v.distance*0.90 > distanceMin:
				lastTime[v.prefix.Value()] = tn
				value := fmt.Sprintf("{\"timeStamp\": %f, \"value\": %q, \"type\": %q}", timeStamp, v.raw, v.prefix.Value())
				ctx.Send(pid, &eventGPS{event: value})
			case time.Since(lastStamp) > timeout:
				lastTime[v.prefix.Value()] = tn
				value := fmt.Sprintf("{\"timeStamp\": %f, \"value\": %q, \"type\": %q}", timeStamp, v.raw, v.prefix.Value())
				ctx.Send(pid, &eventGPS{event: value})
			}
		}
	}
}
