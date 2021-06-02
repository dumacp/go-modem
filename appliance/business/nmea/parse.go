package nmea

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/dumacp/gpsnmea"
)

func Parse(frame string) (interface{}, error) {

	re := regexp.MustCompile(`\$[a-zA-Z]+,`)

	if len(frame) > 34 {
		gtype := re.FindString(frame)
		if strings.Count(frame, "$") > 1 {
			return nil, fmt.Errorf("frame bad format %s", frame)
		}
		if len(gtype) > 3 {
			switch {
			case strings.HasPrefix(frame, "$GPGGA"):
				if vg := gpsnmea.ParseGGA(frame); vg != nil {
					return vg, nil
				}
				return nil, fmt.Errorf("frame bad format %s", frame)
			case strings.HasPrefix(frame, "$GPRMC"):
				if vg := gpsnmea.ParseRMC(frame); vg != nil {
					return vg, nil
				}
				return nil, fmt.Errorf("frame bad format %s", frame)
			default:
				return nil, errors.New("unkown frame")
			}
		}
		return nil, fmt.Errorf("frame bad format %s", frame)
	}
	return nil, fmt.Errorf("frame bad format %s", frame)
}
