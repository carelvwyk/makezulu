package main

import (
	"aws_iot"
	"errors"
	"flag"
	"log"
	"os"
	"time"

	term "github.com/nsf/termbox-go"
	"gobot.io/x/gobot"
	"gobot.io/x/gobot/platforms/ble"
	"gobot.io/x/gobot/platforms/sphero/sprkplus"
)

const deviceName = "Sphero Spark+"

var (
	privKey   = flag.String("iot_privkey", "", "AWS IoT Thing Private Key")
	cert      = flag.String("iot_cert", "", "AWS IoT Thing Certificate")
	thingName = flag.String("iot_thingname", "", "AWS IoT Thing name")
)

func main() {
	bleAdaptor := ble.NewClientAdaptor(os.Args[1])
	// Required for OSX so the adapter doesn't get stuck waiting for a response
	// after the first command is sent:
	bleAdaptor.WithoutResponses(true)

	sprk := sprkplus.NewDriver(bleAdaptor)
	sprk.SetName(deviceName)

	work := func() {
		// Remove that which you do not need
		gobot.Every(1*time.Second, func() {
			r := uint8(gobot.Rand(255))
			g := uint8(gobot.Rand(255))
			b := uint8(gobot.Rand(255))
			sprk.SetRGB(r, g, b)
		})
	}

	go func() {
		evts := sprk.Subscribe()
		for e := range evts {
			log.Printf("BOT EVENT: %+v", e)
		}
	}()

	robot := gobot.NewRobot("sprk",
		[]gobot.Connection{bleAdaptor},
		[]gobot.Device{sprk},
		work,
	)

	// Non-blocking start ("false" parameter):
	if err := robot.Start(false); err != nil {
		log.Printf("Error starting robot: %v", err)
	}

	if err := handleKeyboardInputForever(robot); err != nil {
		log.Printf("Keyboard Input Handler Error: %v", err)
	}
}

func handleKeyboardInputForever(robot *gobot.Robot) error {
	sprk, ok := robot.Device(deviceName).(*sprkplus.SPRKPlusDriver)
	if !ok {
		return errors.New("Unable to cast robot device to Sphero Spark+ Driver")
	}
	// Initialize terminal keyboard input
	if err := term.Init(); err != nil {
		return err
	}
	defer term.Close()

	log.Println("Listening for arrow keys. Press ESC button to stop")

	for {
		switch ev := term.PollEvent(); ev.Type {
		case term.EventKey:
			switch ev.Key {
			case term.KeyEsc:
				log.Println("Bye!")
				// Note: robot.Stop() hangs on connection.Finalize on OSX so
				// we need to kill it somehow
				// Solution - don't bother stopping at all, just die
				return nil
				//return robot.Stop()
			case term.KeyArrowUp:
				log.Println("Arrow Up pressed")
				sprk.Roll(40, 0)
			case term.KeyArrowDown:
				log.Println("Arrow Down pressed")
				sprk.Roll(40, 180)
			case term.KeyArrowLeft:
				log.Println("Arrow Left pressed")
				sprk.Roll(40, 270)
			case term.KeyArrowRight:
				log.Println("Arrow Right pressed")
				sprk.Roll(40, 90)
			default:

			}
		case term.EventError:
			return ev.Err
		}
	}
}

func iotPubSubDemo() {
	flag.Parse()

	iotClient, err := aws_iot.New(*thingName, *privKey, *cert, "us-east-2")
	if err != nil {
		log.Println(err)
		return
	}

	// Handle incoming messages:
	go func() {
		subChan := iotClient.SubChannel()
		for msg := range subChan {
			log.Println(msg.String())
		}
	}()

	// Publish outgoing messages:
	go func() {
		pubChan := iotClient.PubChannel()
		newState := map[string]interface{}{
			"state": map[string]interface{}{
				"reported": map[string]interface{}{
					"red":   187,
					"green": 114,
					"blue":  222,
				},
			},
		}
		log.Printf("Sending: %+v", newState)
		pubChan <- newState
		// Wait to receive and handle a response:
		time.Sleep(time.Second * 5)
		iotClient.Stop()
	}()

	if err := iotClient.Connect(); err != nil {
		log.Printf("iot client error: %v", err)
	}

	log.Println("DONE")
}
