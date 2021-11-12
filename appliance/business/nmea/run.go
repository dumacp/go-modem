/*
Package implements a binary for read serial portNmea nmea.

*/
package nmea

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/dumacp/go-modem/appliance/crosscutting/logs"
	"github.com/dumacp/gpsnmea"
)

const (
	timeoutMax       = 30 * time.Second
	timeoutBadFrames = 10 * time.Minute
	badframeUmbral   = 29
)

var filter = []string{"$GPRMC", "$GPGGA"}

func (act *actornmea) run(chFinish chan int, timeout, distanceMin int) error {

	select {
	case _, ok := <-chFinish:
		if !ok {
			logs.LogWarn.Println("chFinish is closed")
		}
		return nil
	default:
	}

	logs.LogBuild.Println("========== RUN NMEA ============")
	logs.LogBuild.Println("========== ======== ============")

	chQuit := make(chan int)
	defer close(chQuit)

	re := regexp.MustCompile(`\$[a-zA-Z]+,`)

	mapCaptures := make(map[string]string)

	long0 := float64(0)
	lat0 := float64(0)
	long0_t := float64(0)
	lat0_t := float64(0)
	sendDistance := false

	chDist := make(chan int, 0)
	t1 := time.NewTimer(time.Duration(timeout) * time.Second)
	defer t1.Stop()

	go func() {
		defer close(chDist)
		for {
			//only publish if frame GPRMC is not quiet
			select {
			case <-chDist:
				for _, v := range mapCaptures {
					logs.LogBuild.Printf("distance, EVENT -> %s\n", v)
					act.context.Send(act.pubsubPID, &eventGPS{event: v})
				}
				mapCaptures = make(map[string]string)
				if !t1.Stop() {
					select {
					case <-t1.C:
					default:
					}
				}
				t1.Reset(time.Duration(timeout) * time.Second)
			case <-t1.C:
				for _, v := range mapCaptures {
					logs.LogBuild.Printf("timeout, EVENT -> %s\n", v)
					lat0 = lat0_t
					long0 = long0_t
					act.context.Send(act.pubsubPID, &eventGPS{event: v})
				}
				mapCaptures = make(map[string]string)
				t1.Reset(time.Duration(timeout) * time.Second)
			case <-chQuit:
				logs.LogWarn.Println("chQuit nil")
				return
			}
		}
	}()

	funcListen := func() error {
		// defer act.dev.Close()

		defer func() {
			if r := recover(); r != nil {
				logs.LogError.Println("Recovered in funcListen, ", r)
			}
		}()

		ch := act.dev.Read()
		defer act.dev.Close()

		queue := NewQueue()
		type validFrame struct {
			timeStamp string
			raw       string
		}
		mapFrames := make(map[string]*validFrame)
		badFrames := new(validFrame)
		badFrameCount := 0

		tbadcount := time.NewTicker(timeoutBadFrames)
		defer tbadcount.Stop()
		tmax := time.NewTicker(timeoutMax)
		defer tmax.Stop()

		tn := time.Now()
		for {
			select {

			case <-chFinish:
				return errors.New("chQuit nil funcListen in run")
			case <-tbadcount.C:
				rateBad := float64(badFrameCount) / timeoutBadFrames.Minutes()
				badgps := fmt.Sprintf("{\"timeStamp\": %d, \"value\": %.2f, \"type\": %q}", time.Now().Unix(), rateBad, "GPSERROR")

				logs.LogInfo.Printf("last bad GPS frame -> %q", badFrames.raw)
				logs.LogInfo.Printf("rate bad GPS -> %.2f", rateBad)
				act.context.Send(act.pubsubPID, &msgBadGPS{data: badgps})
				if rateBad > badframeUmbral {
					return fmt.Errorf("reset modem NMEA, %w, umbral: %d", umbralError, badframeUmbral)
				}
				badFrameCount = 0
			case <-time.After(60 * time.Second):
				return fmt.Errorf("reset modem NMEA, timeout read frame")
			case frame, ok := <-ch:
				if !ok {
					return fmt.Errorf("device channel error")
				}
				tn = time.Now()
				timeStamp := float64(tn.UnixNano()) / 1000000000
				logs.LogBuild.Printf("frame: %s\n", frame)
				if len(frame) > 34 {
					gtype := re.FindString(frame)
					if strings.Count(frame, "$") > 1 {
						logs.LogWarn.Printf("frame bad format %s", frame)
						continue
					}
					if len(gtype) > 3 {

						switch {
						case strings.HasPrefix(frame, "$GPGGA"):
							sendDistance = false
							if vg := gpsnmea.ParseGGA(frame); vg != nil {
								if !isValidFrame(queue, vg) {
									badFrames = &validFrame{timeStamp: vg.TimeStamp, raw: vg.Raw}
									badFrameCount++
									break
								}
								mapFrames["$GPGGA"] = &validFrame{raw: frame, timeStamp: vg.TimeStamp}

								for k, v := range mapFrames {
									if strings.Contains(v.timeStamp, vg.TimeStamp) {
										act.context.Send(act.pubsubPID, &msgGPS{data: v.raw})
										mapCaptures[k] = fmt.Sprintf("{\"timeStamp\": %f, \"value\": %q, \"type\": %q}", timeStamp, v.raw, k[1:])
									}
								}

								long1 := gpsnmea.LatLongToDecimalDegree(vg.Long, vg.LongCord)
								lat1 := gpsnmea.LatLongToDecimalDegree(vg.Lat, vg.LatCord)
								distance := gpsnmea.Distance(lat0, long0, lat1, long1, "K")
								if distance > float64(distanceMin)*0.90/1000 {
									logs.LogBuild.Printf("distance: %v K\n", distance)
									long0 = long1
									lat0 = lat1
									sendDistance = true
									chDist <- 1
								}
								long0_t = long1
								lat0_t = lat1

							}
						case strings.HasPrefix(frame, "$GPRMC"):
							if vg := gpsnmea.ParseRMC(frame); vg != nil {
								vgga, ok := mapFrames["$GPGGA"]
								if !ok || !strings.Contains(vgga.timeStamp, vg.TimeStamp) {
									mapFrames["$GPRMC"] = &validFrame{raw: frame, timeStamp: vg.TimeStamp}
									break
								}

								mssg := fmt.Sprintf("{\"timeStamp\": %f, \"value\": %q, \"type\": %q}", timeStamp, frame, gtype[1:len(gtype)-1])
								mapCaptures["$GPRMC"] = mssg
								act.context.Send(act.pubsubPID, &msgGPS{data: frame})

								//verify last verify distance
								if sendDistance {
									sendDistance = false
									logs.LogBuild.Printf("distance, EVENT -> %s\n", mssg)
									// act.context.Send(act.pubsubPID, &eventGPS{event: mssg})
								}
							}
						}
					} else {
						badFrames = &validFrame{timeStamp: "", raw: frame}
						badFrameCount++
					}
				} else {
					badFrames = &validFrame{timeStamp: "", raw: frame}
					badFrameCount++
				}
			}
		}
	}
	err := funcListen()
	if err != nil {
		// errlog.Println(err)
		return err
	}
	logs.LogWarn.Printf("funcListen terminated in run nmea")
	return nil
}
