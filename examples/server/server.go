package main

import (
	"fmt"

	iec104 "github.com/github-of-lyj/IEC104"
	"github.com/sirupsen/logrus"
)

type handler struct{}

func (h handler) GeneralInterrogationHandler(apdu *iec104.APDU) error {
	for _, signal := range apdu.Signals {
		fmt.Printf("%f ", signal.Value)
	}
	fmt.Println()
	return nil
}

func (h handler) CounterInterrogationHandler(apdu *iec104.APDU) error {
	for _, signal := range apdu.Signals {
		fmt.Printf("%f ", signal.Value)
	}
	fmt.Println()
	return nil
}

func (h handler) ReadCommandHandler(apdu *iec104.APDU) error {
	fmt.Println("a")
	return nil
}

func (h handler) ClockSynchronizationHandler(apdu *iec104.APDU) error {
	fmt.Println("a")
	return nil
}

func (h handler) TestCommandHandler(apdu *iec104.APDU) error {
	fmt.Println("a")
	return nil
}

func (h handler) ResetProcessCommandHandler(apdu *iec104.APDU) error {
	fmt.Println("a")
	return nil
}

func (h handler) DelayAcquisitionCommandHandler(apdu *iec104.APDU) error {
	fmt.Println("a")
	return nil
}

func (h handler) APDUHandler(apdu *iec104.APDU) error {
	for _, signal := range apdu.Signals {
		fmt.Printf("%f ", signal.Value)
	}
	fmt.Println()
	return nil
}

func main() {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)
	iec104.SetLogger(logger)

	server := iec104.NewServer(":2404", nil, logger)
	if err := server.Serve(&handler{}); err != nil {
		panic(any(err))
	}
}
