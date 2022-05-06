package nmea

import (
	"github.com/dumacp/go-logs/pkg/logs"
	"github.com/dumacp/gpsnmea"
)

func Distance(lastFrame, actualFrame *gpsnmea.Gpgga, distanceMin float64) bool {

	long0 := gpsnmea.LatLongToDecimalDegree(lastFrame.Long, lastFrame.LongCord)
	lat0 := gpsnmea.LatLongToDecimalDegree(lastFrame.Lat, lastFrame.LatCord)

	long1 := gpsnmea.LatLongToDecimalDegree(actualFrame.Long, actualFrame.LongCord)
	lat1 := gpsnmea.LatLongToDecimalDegree(actualFrame.Lat, actualFrame.LatCord)
	distance := gpsnmea.Distance(lat0, long0, lat1, long1, "K")
	if distance > float64(distanceMin)*0.90/1000 {
		logs.LogBuild.Printf("distance: %v K\n", distance)
		return true
	}

	return false
}
