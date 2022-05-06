package device

import (
	"bufio"
	"fmt"
	"strings"
)

var filter = []string{"$GPRMC", "$GPGGA"}

func isSentence(s1 string, filter []string) bool {
	if len(s1) > 8 {
		for _, v := range filter {
			if strings.HasPrefix(s1, v) {
				//if s1[1:8] != "GPRMC,," {
				return true
				//}
			}
		}
	}
	return false
}

func listen(reader *bufio.Reader) (string, error) {

	// re := regexp.MustCompile(`\$[a-zA-Z]+,`)

	v, err := reader.ReadBytes(0x0D)
	if err != nil {
		return "", err
	}
	if len(v) <= 0 {
		return "", fmt.Errorf("empty frame")
	}
	frame := string(v)
	// // tn := time.Now()
	// // timeStamp := float64(tn.UnixNano()) / 1000000000
	// fmt.Printf("frame: %s\n", frame)
	// if len(frame) <= 18 {
	// 	return "", fmt.Errorf("frame bad format %s", frame)
	// }
	// gtype := re.FindString(frame)
	// if strings.Count(frame, "$") > 1 {
	// 	return "", fmt.Errorf("frame bad format %s", frame)
	// }
	// if len(gtype) <= 3 {
	// 	return "", fmt.Errorf("frame bad format %s", frame)
	// }
	// if !isSentence(frame, filter) {
	// 	return "", nil
	// }

	return strings.TrimSpace(frame), nil
}
