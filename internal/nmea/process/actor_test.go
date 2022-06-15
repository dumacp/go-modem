package process

import (
	"bufio"
	"os"
	"testing"
	"time"

	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/dumacp/go-logs/pkg/logs"
)

func TestNewActor(t *testing.T) {
	type args struct {
		timeout     int
		distanceMin int
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.

		{
			name: "test1",
			args: args{
				timeout:     10,
				distanceMin: 60,
			},
		},
	}

	logs.LogInfo = logs.New(os.Stderr, "", 0)
	logs.LogBuild = logs.New(os.Stderr, "", 0)
	logs.LogWarn = logs.New(os.Stderr, "", 0)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewActor(tt.args.timeout, tt.args.distanceMin)
			if _, ok := got.(actor.Actor); !ok {
				t.Errorf("NewActor() = %T, want %T", got, new(actor.Actor))
			}

			sys := actor.NewActorSystem()
			props := actor.PropsFromFunc(got.Receive)
			pid, err := sys.Root.SpawnNamed(props, "test")
			if err != nil {
				t.Errorf("error = %s", err)
			}

			f, err := os.OpenFile("./test_data/gps.txt", os.O_RDONLY, 0666)
			if err != nil {
				t.Errorf("error = %s", err)
			}

			scn := bufio.NewScanner(f)

			for scn.Scan() {
				sys.Root.Send(pid, &MsgData{
					Data: scn.Text(),
					// Data: "$GPRMC,144135.0,A,0609.894786,N,07536.099610,W,0.0,0.0,120522,4.7,W,A*05",
				})
				time.Sleep(1 * time.Second)
			}

		})

	}

	time.Sleep(3 * time.Second)
}
