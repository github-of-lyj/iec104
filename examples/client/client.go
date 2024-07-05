package main

import (
	"fmt"
	"time"

	"github.com/github-of-lyj/iec104"
	"github.com/sirupsen/logrus"
)

const (
	serverAddress = "192.168.23.129:2404"
)

type handler struct{}

func (h handler) GeneralInterrogationHandler(apdu *iec104.APDU) error {
	for _, signal := range apdu.Signals {
		fmt.Printf("%f ", signal.Value)
	}
	fmt.Println("a")
	return nil
}

func (h handler) CounterInterrogationHandler(apdu *iec104.APDU) error {
	for _, signal := range apdu.Signals {
		fmt.Printf("%f ", signal.Value)
	}
	fmt.Println("b")
	return nil
}

func (h handler) ReadCommandHandler(apdu *iec104.APDU) error {
	fmt.Println("c")
	return nil
}

func (h handler) ClockSynchronizationHandler(apdu *iec104.APDU) error {
	fmt.Println("d")
	return nil
}

func (h handler) TestCommandHandler(apdu *iec104.APDU) error {
	fmt.Println("e")
	return nil
}

func (h handler) ResetProcessCommandHandler(apdu *iec104.APDU) error {
	fmt.Println("f")
	return nil
}

func (h handler) DelayAcquisitionCommandHandler(apdu *iec104.APDU) error {
	fmt.Println("g")
	return nil
}

func (h handler) APDUHandler(apdu *iec104.APDU) error {
	fmt.Println("h")
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

	option, err := iec104.NewClientOption(serverAddress, &handler{}, 10*time.Second)
	if err != nil {
		panic(any(err))
	}
	client := iec104.NewClient(option)
	if err := client.Connect(); err != nil {
		panic(any(err))
	}
	defer client.Close()

	// go func() {
	// 	time.Sleep(5 * time.Second)
	// 	client.SendTestFrame()
	// }()

	go func() {
		// for {
		// 	time.Sleep(1 * time.Second)
		// 	client.SendReadCommand(0x000001)
		// }
		// time.Sleep(1 * time.Second)
		// client.SendReadCommand(0x000001)

	}()

	// go func() {
	// 	for {
	// 		if client.Signals != nil {
	// 			// 提取 map 中的所有键
	// 			keys := make([]iec104.IOA, 0, len(client.Signals))
	// 			for key := range client.Signals {
	// 				keys = append(keys, key)
	// 			}
	// 			// 对键进行排序
	// 			n := len(keys)
	// 			for i := 0; i < n; i++ {
	// 				for j := 0; j < n-i-1; j++ {
	// 					if keys[j] > keys[j+1] {
	// 						// 交换相邻元素
	// 						keys[j], keys[j+1] = keys[j+1], keys[j]
	// 					}
	// 				}
	// 			}

	// 			for i := 0; i < n; i++ {
	// 				fmt.Printf("地址为：%d,传输的值为：%f", keys[i], client.Signals[keys[i]])
	// 				fmt.Println()
	// 			}

	// 		}
	// 		time.Sleep(5 * time.Second)
	// 	}
	// }()

	// go func() {
	// 	time.Sleep(2 * time.Second)
	// 	client.SendCounterInterrogation()
	// }()

	// go func() {
	// 	time.Sleep(3 * time.Second)
	// 	if err := client.SendSingleCommand(iec104.IOA(1), true /* close */); err != nil {
	// 		panic(any(err))
	// 	}
	// 	if err := client.SendSingleCommand(iec104.IOA(1), false /* close */); err != nil {
	// 		panic(any(err))
	// 	}
	// 	if err := client.SendDoubleCommand(iec104.IOA(1), true /* close */); err != nil {
	// 		panic(any(err))
	// 	}
	// 	if err := client.SendDoubleCommand(iec104.IOA(1), false /* close */); err != nil {
	// 		panic(any(err))
	// 	}
	// }()

	// go func() {
	// 	time.Sleep(5 * time.Second)
	// 	fmt.Printf("Connected: %v\n", client.IsConnected())
	// }()

	time.Sleep(30 * time.Minute)
}
