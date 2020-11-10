/*
Package implements a binary for read serial portNmea nmea.

*/
package nmea

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/dumacp/go-modem/appliance/crosscutting/logs"
	"github.com/dumacp/gpsnmea"
)

const (
	timeoutMax       = 30 * time.Second
	timeoutBadFrames = 15 * time.Minute
)

var filter = []string{"$GPRMC", "$GPGGA"}

func (act *actornmea) run(timeout, distanceMin int) error {

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

	chQuit := make(chan int, 0)
	defer close(chQuit)

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

	chReset := make(chan int, 0)
	chResetStop := make(chan int, 0)

	timerModem := struct {
		timer  *time.Timer
		active bool
	}{}

	timerModem.timer = time.NewTimer(0)
	if !timerModem.timer.Stop() {
		select {
		case <-timerModem.timer.C:
		default:
		}
	}
	timerModem.active = false

	go func() {
		defer close(chReset)

		for {
			select {
			case <-timerModem.timer.C:
				timerModem.active = false
				chResetStop <- 0
			case v := <-chReset:
				if v <= 0 {
					// logs.LogWarn.Println("modem reset timer off, with GPS data")
					if !timerModem.timer.Stop() {
						select {
						case <-timerModem.timer.C:
						default:
						}
					}
					timerModem.active = false
					break
				}
				if timerModem.active {
					break
				}
				logs.LogWarn.Println("modem reset timer on, without GPS data")
				if !timerModem.timer.Stop() {
					select {
					case <-timerModem.timer.C:
					default:
					}
				}
				timerModem.timer.Reset(time.Duration(v) * time.Second)
				timerModem.active = true
			case <-chQuit:
				logs.LogWarn.Println("chQuit nil")
				return
			}
		}
	}()
	modemVerify := func(value int) {
		select {
		case chReset <- value:
		default:
		}
	}

	funcListen := func() error {
		// defer act.dev.Close()

		defer func() {
			if r := recover(); r != nil {
				errlog.Println("Recovered in funcListen, ", r)
			}
		}()

		ch := act.dev.Read()

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
			case <-chResetStop:
				return fmt.Errorf("reset modem NMEA, timer stop")
			case <-tbadcount.C:
				rateBad := float64(badFrameCount) / timeoutBadFrames.Minutes()
				badgps := fmt.Sprintf("{\"timeStamp\": %d, \"value\": %.2f, \"type\": %q}", time.Now().Unix(), rateBad, "GPSERROR")
				badFrameCount = 0
				logs.LogInfo.Printf("last bad GPS frame -> %q", badFrames.raw)
				logs.LogInfo.Printf("rate bad GPS -> %.2f", rateBad)
				act.context.Send(act.pubsubPID, &msgBadGPS{data: badgps})
			case <-tmax.C:
				// if len(mapCaptures) <= 0 {
				// 	break break_for
				// }
				if time.Now().Unix() > tn.Add(timeoutMax).Unix() {
					return fmt.Errorf("reset modem NMEA, timeout read frame")
				}
			case frame, ok := <-ch:
				if !ok {
					return fmt.Errorf("device channel error")
				}
				tn = time.Now()
				timeStamp := float64(tn.UnixNano()) / 1000000000
				logs.LogBuild.Printf("frame: %s\n", frame)
				if len(frame) > 34 {
					if timerModem.active {
						modemVerify(0)
					}
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
									// logs.LogBuild.Printf("bad GPS frame -> %q", vg.Raw)
									badFrames = &validFrame{timeStamp: vg.TimeStamp, raw: vg.Raw}
									break
								}

								// mapCaptures["$GPGGA"] = fmt.Sprintf("{\"timeStamp\": %f, \"value\": %q, \"type\": %q}", timeStamp, frame, gtype[1:len(gtype)-1])
								mapFrames["$GPGGA"] = &validFrame{raw: frame, timeStamp: vg.TimeStamp}
								// act.context.Send(act.pubsubPID, &msgGPS{data: frame})

								for k, v := range mapFrames {
									if strings.Contains(v.timeStamp, vg.TimeStamp) {
										act.context.Send(act.pubsubPID, &msgGPS{data: v.raw})
										mapCaptures[k] = fmt.Sprintf("{\"timeStamp\": %f, \"value\": %q, \"type\": %q}", timeStamp, v.raw, k[1:])
									}
								}

								long1 := gpsnmea.LatLongToDecimalDegree(vg.Long, vg.LongCord)
								lat1 := gpsnmea.LatLongToDecimalDegree(vg.Lat, vg.LatCord)
								distance := gpsnmea.Distance(lat0, long0, lat1, long1, "K")
								if distance > float64(distanceMin)/1000 {
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
									logs.LogBuild.Printf("distance, EVENT -> %s\n", mssg)
									act.context.Send(act.pubsubPID, &eventGPS{event: mssg})
								}

							}
						}
					}
				} else {
					if !timerModem.active {
						modemVerify(180)
					}
				}
			}
		}
	}
	err := funcListen()
	if err != nil {
		// errlog.Println(err)
		return err
	}
	return fmt.Errorf("funcListen terminated")
}

// func (act *actornmea) resetModem() {
// 	logs.LogWarn.Printf()
// 	act.context.Send(act.modemPID, &messages.ModemReset{})
// 	//TODO:
// }
