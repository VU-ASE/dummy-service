package main

import (
	"time"

	pb_outputs "github.com/VU-ASE/pkg-CommunicationDefinitions/v2/packages/go/outputs"
	pb_systemmanager_messages "github.com/VU-ASE/pkg-CommunicationDefinitions/v2/packages/go/systemmanager"
	servicerunner "github.com/VU-ASE/pkg-ServiceRunner/v2/src"
	zmq "github.com/pebbe/zmq4"
	"github.com/rs/zerolog/log"
	"google.golang.org/protobuf/proto"
)

var tuning *pb_systemmanager_messages.TuningState

func run(
	serviceInfo servicerunner.ResolvedService,
	sysMan servicerunner.SystemManagerInfo,
	initialTuning *pb_systemmanager_messages.TuningState) error {
	tuning = initialTuning

	outputAddr, err := serviceInfo.GetOutputAddress("rpm")
	if err != nil {
		return err
	}

	numOutput, err := servicerunner.GetTuningInt("numSensors", tuning)
	if err != nil {
		return err
	}

	log.Info().Str("output", outputAddr).Msg("pls")
	log.Info().Int("num", numOutput).Msg("pls")

	sock, err := zmq.NewSocket(zmq.PUB)
	if err != nil {
		return err
	}

	err = sock.Bind(outputAddr)
	if err != nil {
		return err
	}

	for {
		dummyMin, err := servicerunner.GetTuningInt("dummyMin", tuning)
		if err != nil {
			return err
		}

		dummyMax, err := servicerunner.GetTuningInt("dummyMax", tuning)
		if err != nil {
			return err
		}

		for sensorId := 1; sensorId < numOutput; sensorId++ {
			for rpm := dummyMin; rpm < dummyMax; rpm++ {

				msg := pb_outputs.SensorOutput{
					SensorId:  uint32(sensorId),
					Status:    0,
					Timestamp: uint64(time.Now().UnixMilli()),
					SensorOutput: &pb_outputs.SensorOutput_ControllerOutput{
						ControllerOutput: &pb_outputs.ControllerOutput{
							LeftThrottle:  float32(rpm),
							RightThrottle: float32(-rpm),
							SteeringAngle: float32(rpm / 2),
						},
					},
				}

				// time.Sleep(5 * time.Second)
				// log.Info().Int("timestamp", int(time.Now().UnixMilli())).Msg("Timestamp")

				time.Sleep(4 * time.Millisecond)

				msgBytes, err := proto.Marshal(&msg)
				if err != nil {
					return err
				}

				_, err = sock.SendBytes(msgBytes, 0)
				if err != nil {
					return err
				}
			}
		}
	}
}

func tuningCallback(newtuning *pb_systemmanager_messages.TuningState) {
	log.Warn().Msg("Tuning state changed")
	log.Info().Msg(newtuning.String())
	tuning = newtuning
}

func main() {
	servicerunner.Run(run, tuningCallback, false)
}
