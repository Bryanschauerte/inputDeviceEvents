package inputDeviceEvents

// taken from github.com/adammck/sixaxis and adapted. Mostly identifying codes incorrect or misused.

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"strings"
)

const (

// digital
	buttonPressEvent = 1
	stickMovement  = 3 // tAnalog

	buttonAnalog = 4 // not using it but leaving it for the next poor soul to try and use it, key codes in select block

	// Digital event codes 0 or 1
	bcSelect = 288
	bcL3     = 289
	bcR3     = 290
	bcStart  = 291
	bcPS     = 304
	aUp       = 292
	aRight    = 293
	aDown     = 294
	aLeft     = 295
	aL2       = 296
	aR2       = 297
	aL1       = 298
	aR1       = 299
	aTriangle = 300
	aCircle   = 301
	aCross    = 302
	aSquare   = 303

	LeftStickX  = 0
	LeftStickY  = 1
	RightStickX = 2
	RightStickY = 3

	// Gyroscope cant seem to find the event
	GyroX = 5//4 // left/right
	GyroY = 6 // 5// forwards/backwards
	GyroZ = 7 // ???

)

// https://github.com/torvalds/linux/blob/master/include/uapi/linux/time.h#L15
// https://github.com/torvalds/linux/blob/master/include/uapi/asm-generic/posix_types.h#L88

type timeVal struct {
	Sec  int32 // seconds
	Usec int32 // microseconds
}

// https://github.com/torvalds/linux/blob/master/include/uapi/linux/input.h#L24
type inputEvent struct {
	Time  timeVal
	Type  uint16
	Code  uint16
	Value int32
}

type AnalogStick struct {
	X int32
	Y int32
}

func (as *AnalogStick) String() string {
	return fmt.Sprintf("%+04d, %+04d", as.X, as.Y)
}

// Also stored as int32
type Orientation struct {
	RawX int32
	RawY int32
	RawZ int32
}

// X returns the orientation on the X axis (aka roll, or bank), in the range -1
// (90 degrees left) to +1 (90 degrees right).
func (o *Orientation) X() float64 {

	// Note that the scale is inverted for this axis. I have no idea why.
	// See: https://github.com/falkTX/qtsixa/blob/master/sixad/sixaxis.cpp#L97
	return clamp((float64(o.RawX) + 512) / 110)
}

// Y returns the orientation on the Y axis (aka pitch), in the range -1 (90
// degrees forwards, away from the player) to +1 (90 degrees backwards).
func (o *Orientation) Y() float64 {
	return (float64(o.RawY))
}

// Z returns the orientation on the Z axis. It's unclear what this axis
// represents, but it doesn't seem to be the heading.
func (o *Orientation) Z() float64 {
	return clamp((float64(o.RawZ) - 512) / 110)
}

// TODO: Scale to the range of values
func (o *Orientation) String() string {
	return fmt.Sprintf("&Orientation{x=%+04d, y=%+04d, z=%+04d}", o.RawX, o.RawY, o.RawZ)
}

type SA struct {
	r io.Reader

	// Digital Buttons
	Select bool
	L3     bool
	R3     bool
	Start  bool
	PS     bool
	Up       bool
	Right    bool
	Down     bool
	Left     bool
	L2       bool
	R2       bool
	L1       bool
	R1       bool
	Triangle bool
	Circle   bool
	Cross    bool
	Square   bool
	LeftStick  *AnalogStick
	RightStick *AnalogStick

	// Gyro not used
	Orientation *Orientation
}

func New(reader io.Reader) *SA {
	return &SA{
		r:           reader,
		LeftStick:   &AnalogStick{},
		RightStick:  &AnalogStick{},
		Orientation: &Orientation{-512, 512, 512},
	}
}

// String returns the current state of the controller as a string.
func (sa *SA) String() string {
	s := make([]string, 0, 30)

	// sticks
	if sa.LeftStick.X != 0 {
		s = append(s, fmt.Sprintf("LX=%+04d", sa.LeftStick.X))
	}
	if sa.LeftStick.Y != 0 {
		s = append(s, fmt.Sprintf("LY=%+04d", sa.LeftStick.Y))
	}
	if sa.RightStick.X != 0 {
		s = append(s, fmt.Sprintf("RX=%+04d", sa.RightStick.X))
	}
	if sa.RightStick.Y != 0 {
		s = append(s, fmt.Sprintf("RY=%+04d", sa.RightStick.Y))
	}

	// gyro
	if sa.Orientation.RawX != 0 {
		s = append(s, fmt.Sprintf("OX=%+04d", sa.Orientation.RawX))
	}
	if sa.Orientation.RawY != 0 {
		s = append(s, fmt.Sprintf("OY=%+04d", sa.Orientation.RawY))
	}
	if sa.Orientation.RawZ != 0 {
		s = append(s, fmt.Sprintf("OZ=%+04d", sa.Orientation.RawZ))
	}

	// dpad
	if sa.Up {
		s = append(s, fmt.Sprintf("up=%d", sa.Up))
	}
	if sa.Down {
		s = append(s, fmt.Sprintf("down=%d", sa.Down))
	}
	if sa.Left {
		s = append(s, fmt.Sprintf("left=%d", sa.Left))
	}
	if sa.Right {
		s = append(s, fmt.Sprintf("right=%d", sa.Right))
	}

	// other analogs
	if sa.L1 {
		s = append(s, fmt.Sprintf("L1=%d", sa.L1))
	}
	if sa.L2 {
		s = append(s, fmt.Sprintf("L2=%d", sa.L2))
	}
	if sa.R1 {
		s = append(s, fmt.Sprintf("R1=%d", sa.R1))
	}
	if sa.R2 {
		s = append(s, fmt.Sprintf("R2=%d", sa.R2))
	}
	if sa.Triangle {
		s = append(s, fmt.Sprintf("T=%d", sa.Triangle))
	}
	if sa.Circle {
		s = append(s, fmt.Sprintf("C=%d", sa.Circle))
	}
	if sa.Cross {
		s = append(s, fmt.Sprintf("X=%d", sa.Cross))
	}
	if sa.Square {
		s = append(s, fmt.Sprintf("S=%d", sa.Square))
	}

	// digital buttons
	if sa.Select {
		s = append(s, "select")
	}
	if sa.L3 {
		s = append(s, "L3")
	}
	if sa.R3 {
		s = append(s, "R3")
	}
	if sa.Start {
		s = append(s, "start")
	}
	if sa.PS {
		s = append(s, "PS")
	}

	return fmt.Sprintf("&Sixaxis{%s}", strings.Join(s, ", "))
}

// Update changes the state of the controller to reflect the changes in an
// input event. This should be called every time an input event is received.
func (sa *SA) Update(event *inputEvent) {

	switch event.Type {
	case 0:
		// Zero events show up all the time, but never contain any codes or
		// values.

// there are also analog events that correspond between the digital, digital is easier to
// deal with for holding one down vs knowing one was pressed
	case buttonPressEvent:
		v := buttonToBool(event.Value)
		switch event.Code {
		case bcSelect:
			sa.Select = v

		case bcL3:
			sa.L3 = v

		case bcR3:
			sa.R3 = v

		case bcStart:
			sa.Start = v

		case bcPS:
			sa.PS = v

		case aUp:
			sa.Up = v

		case aRight:
			sa.Right = v

		case aDown:
			sa.Down = v

		case aLeft:
			sa.Left = v

		case aL2:
			sa.L2 = v

		case aR2:
			sa.R2 = v

		case aL1:
			sa.L1 = v

		case aR1:
			sa.R1 = v

		case aTriangle:
			sa.Triangle = v

		case aCircle:
			sa.Circle = v

		case aCross:
			sa.Cross = v

		case aSquare:
			sa.Square = v

		default:
			// There are a lot of events which we ignore here, because they're
			// digital representations of the analog buttons.
		}


		// leaving in case of some need though digital button press 1 or 0 is easier to deal with
		// O *sometime 4 4 589838}
		// x *sometime 4 4 589839}
		// Squaure *sometime 4 4 589840
		// triangle *sometime 4 4 589837}
	// case buttonAnalog:
	// 	switch event.Code {

	// 	case 4:
	// 		fmt.Println("its an x", event)

	// 	default:
	// 		fmt.Println("its an UNKNOWN", event)
	// 	}
	//

	case stickMovement:
		switch event.Code {

		case LeftStickX:
			sa.LeftStick.X = event.Value

		case LeftStickY:
			sa.LeftStick.Y = event.Value

		case RightStickX:
			sa.RightStick.X = event.Value

		case RightStickY:
			sa.RightStick.Y = event.Value

		// case GyroX:
		// 	fmt.Println(event,"possibley stick Y")
		// 	sa.Orientation.RawX = event.Value // inverted?

		// case GyroY:
		// 	fmt.Println(event,"GYX stick Y")
		// 	sa.Orientation.RawY = event.Value

		// case GyroZ:
		// 	sa.Orientation.RawZ = event.Value

		default:

			fmt.Println(event, "UNKNOWN stick event")
		}

	default:
		fmt.Println(event,"Unknown group event")
	}
}

// Run loops forever, keeping the state of the controller up to date. This
// should be called in a goroutine.
func (sa *SA) Run() {
	var event inputEvent 

	for {
		binary.Read(sa.r, binary.LittleEndian, &event)

		sa.Update(&event)
	}
}

func buttonToBool(value int32) bool {
	return value == 1
}

func dumpEvent(event *inputEvent) string {
	return fmt.Sprintf("type=%04d, code=%04d, value=%08d\n", event.Type, event.Code, event.Value)
}

func clamp(v float64) float64 {
	return math.Min(math.Max(v, -1), 1)
}
