package nmea

import (
	"container/list"
	"fmt"
	"sync"
	"time"

	"github.com/dumacp/gpsnmea"
	"github.com/golang/geo/s2"
)

type queue struct {
	mux  sync.Mutex
	list *list.List
}

func NewQueue() *queue {
	q := &queue{}
	q.mux = sync.Mutex{}
	q.list = list.New()
	return q
}

func (q *queue) push(v interface{}) *list.Element {
	if v == nil {
		return nil
	}
	q.mux.Lock()
	defer q.mux.Unlock()
	if q.list.Len() > 5 {
		if e := q.list.Front(); e != nil {
			q.list.Remove(e)
		}
	}
	return q.list.PushBack(v)
}

type frameElem struct {
	prefix    string
	timeStamp string
	frame     interface{}
}

func isValidFrame(q *queue, frame *gpsnmea.Gpgga) bool {

	v1 := verifyHDOP(frame)
	v2 := verifySatellites(v1)
	v3 := verifyAltitude(v2)
	v4 := verifyDistance(q, v3)

	if v4 != nil {
		return true
	}
	return false
}

// 	if v4 != nil {
// 		select {
// 		case ch <- &frameElem{
// 			prefix:    v4.Fileds[0],
// 			frame:     v4,
// 			timeStamp: v4.TimeStamp,
// 		}:
// 		case <-time.After(1 * time.Second):
// 		}

// 		for k, vf := range mapFrames {
// 			if strings.Contains(vf.timeStamp, v4.TimeStamp) {
// 				select {
// 				case ch <- vf:
// 				case <-time.After(1 * time.Second):
// 				}
// 			}
// 			delete(mapFrames, k)
// 		}
// 	}

// case vg := <-chRMC:
// 	if v, ok := mapFrames["$GPGGA"]; ok {
// 		if strings.Contains(v.timeStamp, vg.TimeStamp) {
// 			select {
// 			case ch <- &frameElem{
// 				prefix:    vg.Fields[0],
// 				frame:     vg,
// 				timeStamp: vg.TimeStamp,
// 			}:
// 			case <-time.After(1 * time.Second):
// 			}
// 		}
// 		delete(mapFrames, vg.Fields[0])
// 	}
// }

// 	}
// }()

// 	go func
// 	for {
// 		select {
// 		case v := <-frames:
// 			switch {
// 			case strings.HasPrefix(v, "$GPRMC"):
// 				if vg := gpsnmea.ParseRMC(v); vg != nil {
// 					select {
// 					case chRMC <- vg:
// 						mapFrames[vg.Fields[0]] = &frameElem{prefix: vg.Fields[0], frame: vg, timeStamp: vg.TimeStamp}
// 					default:
// 					}
// 				}
// 			case strings.HasPrefix(v, "$GPGGA"):
// 				if vg := gpsnmea.ParseGGA(v); vg != nil {
// 					select {
// 					case chGGA <- vg:
// 						mapFrames[vg.Fileds[0]] = &frameElem{prefix: vg.Fileds[0], frame: vg, timeStamp: vg.TimeStamp}
// 					default:
// 					}
// 				}
// 			}
// 		case <-quit:
// 			return
// 		}
// 	}

// 	return ch
// }

func verifyHDOP(g1 *gpsnmea.Gpgga) *gpsnmea.Gpgga {
	if g1 == nil {
		return nil
	}
	if g1.HDop > 1.6 && g1.NumberSat < 5 {
		return nil
	}

	if g1.HDop > 1.7 {
		return nil
	}

	return g1
}

func verifySatellites(g1 *gpsnmea.Gpgga) *gpsnmea.Gpgga {
	if g1 == nil {
		return nil
	}
	if g1.NumberSat < 3 {
		return nil
	}
	return g1
}

func verifyAltitude(g1 *gpsnmea.Gpgga) *gpsnmea.Gpgga {
	if g1 == nil {
		return nil
	}
	if g1.Altitude > 3500 {
		return nil
	}
	if g1.Altitude < 1300 {
		return nil
	}
	return g1
}

func verifyDistance(q *queue, g1 *gpsnmea.Gpgga) *gpsnmea.Gpgga {

	defer func() {
		if g1 != nil {
			q.push(g1)
		}
	}()
	if q == nil || g1 == nil || q.list.Len() < 3 {
		return nil
	}

	funCompare := func(gga1, gga0 *gpsnmea.Gpgga) bool {

		lat0 := gpsnmea.LatLongToDecimalDegree(gga0.Lat, gga0.LatCord)
		lat1 := gpsnmea.LatLongToDecimalDegree(gga1.Lat, gga1.LatCord)
		lng0 := gpsnmea.LatLongToDecimalDegree(gga0.Long, gga0.LongCord)
		lng1 := gpsnmea.LatLongToDecimalDegree(gga1.Long, gga1.LongCord)

		t0, err := time.Parse("150405", fmt.Sprintf("%s", gga0.TimeStamp))
		if err != nil {
			return false
		}
		t1, err := time.Parse("150405", fmt.Sprintf("%s", gga1.TimeStamp))
		if err != nil {
			return false
		}

		if t0.After(t1) {
			return false
		}

		tDiff := t1.Sub(t0).Hours()

		p0 := s2.LatLngFromDegrees(lat0, lng0)
		p1 := s2.LatLngFromDegrees(lat1, lng1)

		dDiff := p0.Distance(p1).Degrees() * 111.139

		vel := dDiff / tDiff

		if vel > 120 {
			return false
		}

		return true
	}

	e := q.list.Front()
	for e != nil {
		g0, ok := e.Value.(*gpsnmea.Gpgga)
		if !ok {
			return nil
		}
		if !funCompare(g1, g0) {
			return nil
		}
		e = e.Next()
	}

	return g1
}
