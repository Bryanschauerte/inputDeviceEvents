package main

import (
	"gobot.io/x/gobot/drivers/gpio"
        "fmt"
		"gobot.io/x/gobot/platforms/raspi"
        "gobot.io/x/gobot"
		"gobot.io/x/gobot/platforms/keyboard"
		"gobot.io/x/gobot/drivers/i2c"
		"github.com/bryanschauerte/inputDeviceEvents"
		"os"
		"math"
		"time"

)

func getPercentAmountFromJoystickMovement(valueFromInputEvent int32) int32{

	var amount int32
	isReverseDirection:=false

	if(valueFromInputEvent<0){
		isReverseDirection = true
		amount = int32(math.Round((float64(valueFromInputEvent )/-127)*100))

	}

	if(valueFromInputEvent>=0){
		isReverseDirection = false
		amount = int32(math.Round((float64(valueFromInputEvent )/127)*100))

	}

	if (amount> 100){
		amount =  100
	}

	if(isReverseDirection){
		return (-amount)
	}
	return amount

}


func motorMovement(
	rightOne *gpio.LedDriver, rightTwo *gpio.LedDriver,
	leftOne *gpio.LedDriver, leftTwo *gpio.LedDriver,
	pca9685 *i2c.PCA9685Driver,
	sa *inputDeviceEvents.SA){
		leftSpeedAmount := getPercentAmountFromJoystickMovement(int32(sa.LeftStick.Y))
		rightSpeedAmount := getPercentAmountFromJoystickMovement(int32(sa.RightStick.Y))


		if(rightSpeedAmount<0){
			rightNegSpeedAmount:=makePositive(rightSpeedAmount)
			rightNegspeedPercentConvert:= int(float64(4096* rightNegSpeedAmount) * .01) 
			if(rightNegspeedPercentConvert>=4096){
				rightNegspeedPercentConvert = 4000
			}
			rightOne.Off()
			rightTwo.On()
			pca9685.SetPWM(0, 0, uint16(rightNegspeedPercentConvert))
	
			}else {
		
			speedPercentConvert:= int(float64(4096* rightSpeedAmount) * .01) 
			if(speedPercentConvert>=4096){
				speedPercentConvert = 4000
			}

			pca9685.SetPWM(0, 0, uint16(speedPercentConvert))
			fmt.Println(speedPercentConvert, "right")
			rightOne.On()
			rightTwo.Off()
		}

		if(leftSpeedAmount<0){
			leftNegSpeedVal:=makePositive(leftSpeedAmount)
			leftnegspeedconvert:= int(float64(4096* leftNegSpeedVal) * .01) 
			if(leftnegspeedconvert>=4096){
				leftnegspeedconvert = 4000
			}
			
			leftOne.Off()
			leftTwo.On()
			fmt.Println(leftnegspeedconvert, "left")
			pca9685.SetPWM(1, 0, uint16(leftnegspeedconvert))
			
			}else {
		
			positiveLeftSpeedConv:= int(float64(4096* leftSpeedAmount) * .01) 
			if(positiveLeftSpeedConv>4000){
				positiveLeftSpeedConv = 4000
			}
			fmt.Println(positiveLeftSpeedConv, "neg left")
			pca9685.SetPWM(1, 0, uint16(positiveLeftSpeedConv))

			leftOne.On()
			leftTwo.Off()
		}
		

}

func makePositive(x int32) int32 {
	if x < 0 {
		return -x
	}
	return x
}

func watchPsThreeChange(
	rightOne *gpio.LedDriver, rightTwo *gpio.LedDriver,
	leftOne *gpio.LedDriver, leftTwo *gpio.LedDriver,
	pca9685 *i2c.PCA9685Driver){
	f, err := os.Open("/dev/input/event0")
	// stateOfControllerSticks:= TankState{movement: make(map[string]int32)}
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	playstationState := inputDeviceEvents.New(f)
	go playstationState.Run()

	// loop to check the controller state every half second
	c := time.Tick(500 * time.Millisecond)
	for _ = range c {

		motorMovement(
			rightOne, rightTwo,
			leftOne, leftTwo,
			pca9685,
			playstationState)

	}
}

func main() {
        keys := keyboard.NewDriver()
		r := raspi.NewAdaptor()
		pca9685 := i2c.NewPCA9685Driver(r, i2c.WithBus(1),
		i2c.WithAddress(0x70))
		leftOne := gpio.NewLedDriver(r, "40")
		leftTwo := gpio.NewLedDriver(r, "38")	
		rightOne := gpio.NewLedDriver(r, "37")
		rightTwo := gpio.NewLedDriver(r, "35")

		go watchPsThreeChange(rightOne, rightTwo,
			leftOne, leftTwo,
			pca9685)

        work := func() {

                keys.On(keyboard.Key, func(data interface{}) {
						key := data.(keyboard.KeyEvent)
						fmt.Println("key pressed", data)

                        if key.Key == keyboard.A {


							fmt.Println("keyboard event! A", key, key.Char)
                        } else {
                                fmt.Println("keyboard event!", key, key.Char)
                        }
				})

			
        }

        robot := gobot.NewRobot("keyboardbot",
                []gobot.Connection{r},
                []gobot.Device{keys, leftOne, leftTwo, pca9685, rightOne, rightTwo},
                work,
        )

        robot.Start()
}