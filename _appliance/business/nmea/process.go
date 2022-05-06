package nmea

import (
	"errors"
	"fmt"
	"strings"

	"github.com/dumacp/gpsnmea"
)

type TypeFrame int

const (
	GPGGA TypeFrame = iota
	GPRMC
)

func (t TypeFrame) Value() string {
	switch t {
	case GPGGA:
		return "GPGGA"
	case GPRMC:
		return "GPRMC"
	default:
		return ""
	}
}

type processedData struct {
	raw       string
	prefix    TypeFrame
	lat       float64
	lgt       float64
	distance  float64
	timeStamp string
	valided   bool
}

func processData(frame string, distanceMin int,
	lastPoint []float64, queue *queue) (*processedData, error) {
	switch {
	case strings.HasPrefix(frame, "$GPGGA"):

		if vg := gpsnmea.ParseGGA(frame); vg != nil {

			if !isValidFrame(queue, vg) {
				return &processedData{
					raw:     frame,
					prefix:  GPGGA,
					valided: false,
				}, errors.New("invalid frame")
			}
			long1 := gpsnmea.LatLongToDecimalDegree(vg.Long, vg.LongCord)
			lat1 := gpsnmea.LatLongToDecimalDegree(vg.Lat, vg.LatCord)

			distance := gpsnmea.Distance(lastPoint[0], lastPoint[1], lat1, long1, "K")
			// if distance > float64(distanceMin)*0.90/1000 {
			fmt.Printf("distance: %v K\n", distance)
			return &processedData{
				raw:       frame,
				prefix:    GPGGA,
				lat:       lat1,
				lgt:       long1,
				valided:   true,
				timeStamp: vg.TimeStamp,
				distance:  distance,
			}, nil
			// }
		}
	case strings.HasPrefix(frame, "$GPRMC"):
		if vg := gpsnmea.ParseRMC(frame); vg != nil {

			if !vg.Validity {
				return &processedData{
					raw:     frame,
					prefix:  GPRMC,
					valided: false,
				}, errors.New("invalid frame")
			}

			long1 := gpsnmea.LatLongToDecimalDegree(vg.Long, vg.LongCord)
			lat1 := gpsnmea.LatLongToDecimalDegree(vg.Lat, vg.LatCord)

			distance := gpsnmea.Distance(lastPoint[0], lastPoint[1], lat1, long1, "K")
			// if distance > float64(distanceMin)*0.90/1000 {
			fmt.Printf("distance: %v K\n", distance)
			return &processedData{
				raw:       frame,
				prefix:    GPRMC,
				lat:       lat1,
				lgt:       long1,
				valided:   true,
				timeStamp: vg.TimeStamp,
				distance:  distance,
			}, nil
			// }
		}
	default:
		return nil, errors.New("frame type not supported")
	}
	return nil, nil
}
