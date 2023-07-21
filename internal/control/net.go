package control

import (
	"fmt"
	"os/exec"
)

func ifUp() error {
	cmd1 := exec.Command("ifup", "eth1")
	if res, err := cmd1.Output(); err != nil {
		return fmt.Errorf("%s", res)
	}
	return nil
}

func ifDown() error {
	cmd1 := exec.Command("ifdown", "eth1")
	if res, err := cmd1.Output(); err != nil {
		return fmt.Errorf("%s", res)
	}
	return nil
}

//func getAPN() string {
//	apn := os.Getenv("APN")
//	if len(apn) <= 0 {
//		apn = ""
//	}
//	return apn
//}

func pingFunc(testIP []string) (err error) {
	var res []byte
	for _, v := range testIP {
		for range []int{1, 2, 3} {
			cmd1 := exec.Command("ping", v, "-c", "2", "-W", "2", "-q")
			res, err = cmd1.Output()
			if err == nil {
				return nil
			}
		}
	}
	return fmt.Errorf("%s", res)
}
