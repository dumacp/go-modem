package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/dumacp/go-modem/appliance/business/control"
	"github.com/dumacp/go-modem/appliance/business/nmea"
	"github.com/dumacp/go-modem/appliance/crosscutting/logs"
)

var debug bool
var logstd bool
var mqtt bool
var port int
var ipTest string
var apnConn string

var timeout int
var baudRate int
var portNmea string
var portModem string
var distanceMin int

var version bool
var reset bool

const (
	pathudev      = "/etc/udev/rules.d/local.rules"
	ipTestInitial = "8.8.8.8"
	versionString = "1.0.27"
)

func init() {
	flag.BoolVar(&debug, "debug", false, "debug")
	flag.BoolVar(&logstd, "logStd", false, "logs in stderr")
	flag.BoolVar(&version, "version", false, "swho version")
	flag.BoolVar(&reset, "disablereset", false, "disable remote reset")
	flag.BoolVar(&mqtt, "mqtt", false, "[DEPRECATED] send messages to local broker.")
	flag.IntVar(&port, "port", 8082, "port actor in remote mode")
	flag.IntVar(&timeout, "timeout", 30, "timeout to capture frames.")
	flag.IntVar(&baudRate, "baudRate", 115200, "baud rate to capture nmea's frames.")
	flag.StringVar(&portNmea, "portNmea", "/dev/ttyGPS", "device serial to read.")
	flag.StringVar(&portModem, "portModem", "/dev/ttyMODEM", "device serial to conf modem.")
	flag.StringVar(&apnConn, "apn", "", "APN net")
	flag.StringVar(&ipTest, "testip", ipTestInitial, "test IP (ping test connection)")
	flag.IntVar(&distanceMin, "distance", 30, "minimun distance traveled before to send")
}

func main() {

	flag.Parse()
	if version {
		fmt.Printf("version: %s\n", versionString)
		os.Exit(2)
	}
	initLogs(debug, logstd)

	if strings.ContainsAny(ipTest, ipTestInitial) ||
		len(apnConn) <= 0 {
		if testIP, err := getTestIP(); err == nil {
			log.Printf("new testIP from ENV: %q", testIP)
			ipTest = testIP
		}
		if apn, err := getAPN(); err == nil {
			log.Printf("new APN from ENV: %q", apn)
			apnConn = apn
		}
		// if err1 == nil || err2 == nil {

		// 	rootContext.Send(pidCheck, &messages.ModemCheck{
		// 		Addr: testIP,
		// 		Apn:  apn,
		// 	})
		// }
	}

	if portNmea == "/dev/ttyGPS" {
		if fileenv, err := os.Open(pathudev); err != nil {
			logs.LogWarn.Printf("error: reading file UDEV, %s", err)
		} else {
			scanner := bufio.NewScanner(fileenv)
			succ := false
			for scanner.Scan() {
				line := scanner.Text()
				// log.Println(line)
				if strings.Contains(line, "ttyGPS") {
					succ = true
					break

				}
			}
			if !succ {
				portNmea = "/dev/ttyUSB1"
			}
		}
	}
	logs.LogBuild.Printf("portNmea: %s", portNmea)

	if portModem == "/dev/ttyMODEM" {
		if fileenv, err := os.Open(pathudev); err != nil {
			logs.LogWarn.Printf("error: reading file UDEV, %s", err)
		} else {
			scanner := bufio.NewScanner(fileenv)
			succ := false
			for scanner.Scan() {
				line := scanner.Text()
				// log.Println(line)
				if strings.Contains(line, "ttyMODEM") {
					succ = true
					break

				}
			}
			if !succ {
				portModem = "/dev/ttyUSB2"
			}
		}
	}
	logs.LogBuild.Printf("portModem: %s", portModem)

	// remote.Start(fmt.Sprintf("127.0.0.1:%v", port),
	// 	remote.WithAdvertisedAddress(fmt.Sprintf("localhost:%v", port)))

	rootContext := actor.NewActorSystem().Root

	ctx := context.Background()
	context.WithValue(ctx, "ROOTCONTEX", rootContext)

	// var msgChan chan string
	// pub, err := pubsub.NewConnection("go-netmodem")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// defer pub.Disconnect()
	// msgChan = make(chan string)
	// go pub.Publish("SYSTEM/ACTOR/modem", msgChan)
	// go func() {
	// 	for v := range pub.Err {
	// 		log.Println(v)
	// 	}
	// }()

	// pubsubMiddleware := func(next actor.SenderFunc) actor.SenderFunc {
	// 	return func(ctx actor.SenderContext, target *actor.PID, envelope *actor.MessageEnvelope) {
	// 		switch msg := envelope.Message.(type) {
	// 		case *messages.ModemReset:
	// 			msgS, err := json.Marshal(msg)
	// 			if err != nil {
	// 				break
	// 			}
	// 			select {
	// 			case msgChan <- string(msgS):
	// 			case <-time.After(3 * time.Second):
	// 			}
	// 		default:
	// 			next(ctx, target, envelope)
	// 		}
	// 	}
	// }

	pidCheck := &actor.PID{}

	props := actor.PropsFromFunc(func(c actor.Context) {
		switch msg := c.Message().(type) {
		case *actor.Started:
			nmeaA := nmea.NewNmeaActor(
				debug,
				portNmea,
				baudRate,
				timeout,
				distanceMin,
			)
			propsNmea := actor.PropsFromFunc(nmeaA.Receive)
			controlA := control.NewCheckModemActor(debug, reset, portModem, ipTest, apnConn)
			propsCheck := actor.PropsFromFunc(controlA.Receive)
			pidNmea, err := c.SpawnNamed(propsNmea, "nmeaGPS")
			if err != nil {
				logs.LogError.Panic(err)
			}
			pidCheck, err = c.SpawnNamed(propsCheck, "checkmodem")
			if err != nil {
				logs.LogError.Panic(err)
			}
			c.Watch(pidNmea)
			c.Watch(pidCheck)
			c.Send(pidNmea, &nmea.AddressModem{
				Addr: pidCheck.GetAddress(),
				ID:   pidCheck.GetId(),
			})
			// c.Send(pidCheck, &messages.ModemCheck{
			// 	Addr: "8.8.8.8",
			// 	Apn:  "",
			// })

		case *actor.Terminated:
			logs.LogError.Printf("actor terminated: %s", msg.Who.GetId())
		}
	})

	_, err := rootContext.SpawnNamed(props, "modem")
	if err != nil {
		logs.LogError.Fatalln(err)
	}
	time.Sleep(100 * time.Millisecond)

	// funcModemAddr := func(msg *messRecepcionist.SubscribeAdvertising) {
	// 	log.Printf("receive Advertising: %v\n", msg)
	// 	msgAddr := &messages.ModemAddr{
	// 		Addr: msg.Addr,
	// 		Id:   msg.Id,
	// 	}
	// 	rootContext.Send(pid, msgAddr)
	// }

	// serviceRecp, err := receptionist.NewReceptionist("127.0.0.1:8081")
	// if err != nil {
	// 	log.Println(err)
	// } else {
	// 	log.Println("START Receptionist")
	// 	serviceRecp.Register("netmodem", funcModemAddr, []string{"ctrlmodem"})
	// }

	// time.Sleep(2 * time.Second)
	// rootContext.Send(pid, &messages.ModemCheck{
	// 	Addr: "127.0.0.1",
	// 	Apn:  "",
	// })
	// rootContext.Send(pid, &messages.TestIP{Addr: "127.0.0.1"})

	finish := make(chan os.Signal, 1)
	signal.Notify(finish, syscall.SIGINT)
	signal.Notify(finish, syscall.SIGTERM)

	for {
		select {
		case v := <-finish:
			logs.LogError.Println(v)
			return
		}
	}

}

func getTestIP() (string, error) {
	testIP := os.Getenv("TEST_IP")
	if len(testIP) <= 0 {
		return "", fmt.Errorf("TEST_IP not found")
	}
	// varSplit := strings.Split(testIP, " ")
	// if len(varSplit) <= 0 {
	// 	return "", fmt.Errorf("TEST_IP not found")
	// }
	// testIP = varSplit[0]
	if strings.Contains(testIP, ipTest) && len(ipTest) > 0 {
		return "", fmt.Errorf("already testIP")
	}
	// ipTest = testIP
	return testIP, nil
}

func getAPN() (string, error) {
	apn := os.Getenv("APN")
	if len(apn) <= 0 {
		return "", fmt.Errorf("APN not found")
	}
	// varSplit := strings.Split(apn, " ")
	// if len(varSplit) <= 0 {
	// 	return "", fmt.Errorf("APN not found")
	// }
	// apn = varSplit[0]
	// log.Printf("new APN from ENV: %q", apn)
	if strings.Contains(apn, apnConn) && len(apnConn) > 0 {
		return "", fmt.Errorf("already APN")
	}
	// apnConn = apn
	return apn, nil
}
