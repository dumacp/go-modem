package device

import (
	"bufio"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/dumacp/go-logs/pkg/logs"
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

func listen(reader *bufio.Reader, quit <-chan int, chError chan error, chData chan string) {

	defer func() {
		close(chError)
		close(chData)
	}()

	funcError := func(err error) {
		select {
		case <-quit:
		case chError <- err:
		case <-time.After(time.Second * 30):
		}
	}

	re := regexp.MustCompile(`\$[a-zA-Z]+,`)
	wrongFrameCount := 0
	emptyFrameCount := 0

	for {
		select {
		case <-quit:
			return
		default:
		}
		v, err := reader.ReadBytes(0x0D)
		if err != nil {
			funcError(fmt.Errorf("read fail"))
			continue
		}
		if len(v) <= 0 {
			if emptyFrameCount > 30 {
				funcError(fmt.Errorf("continuos empty frame"))
				return
			}
			emptyFrameCount += 1
			continue
		}
		frame := string(v)
		// tn := time.Now()
		// timeStamp := float64(tn.UnixNano()) / 1000000000
		fmt.Printf("frame: %s\n", frame)
		if len(frame) <= 34 {
			logs.LogWarn.Printf("frame bad format %s", frame)
			if wrongFrameCount > 3 {
				funcError(fmt.Errorf("wrong format frame: %s", frame))
			}
			wrongFrameCount += 1
			continue
		}
		gtype := re.FindString(frame)
		if strings.Count(frame, "$") > 1 {
			logs.LogWarn.Printf("frame bad format %s", frame)
			if wrongFrameCount > 3 {
				funcError(fmt.Errorf("wrong format frame: %s", frame))
			}
			wrongFrameCount += 1
			continue
		}
		if len(gtype) <= 3 {
			logs.LogWarn.Printf("frame bad format %s", frame)
			if wrongFrameCount > 3 {
				funcError(fmt.Errorf("wrong format frame: %s", frame))
			}
			wrongFrameCount += 1
			continue
		}
		if !isSentence(frame, filter) {
			continue
		}
		select {
		case <-quit:
		case chData <- frame:
		case <-time.After(1 * time.Second):
		}
	}
}
