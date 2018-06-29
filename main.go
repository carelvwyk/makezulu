package main

import (
	"errors"
	"log"
	"os"

	term "github.com/nsf/termbox-go"
	"gobot.io/x/gobot"
	"gobot.io/x/gobot/platforms/ble"
	"gobot.io/x/gobot/platforms/sphero/sprkplus"
)

const deviceName = "Sphero Spark+"

func main() {
	bleAdaptor := ble.NewClientAdaptor(os.Args[1])
	// Required for OSX so the adapter doesn't get stuck waiting for a response
	// after the first command is sent:
	bleAdaptor.WithoutResponses(true)

	sprk := sprkplus.NewDriver(bleAdaptor)
	sprk.SetName(deviceName)

	evts := sprk.Subscribe()
	go func() {
		for e := range evts {
			log.Printf("BOT EVENT: %+v", e)
		}
	}()

	robot := gobot.NewRobot("sprk",
		[]gobot.Connection{bleAdaptor},
		[]gobot.Device{sprk},
		nil,
	)

	// Non-blocking start:
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
