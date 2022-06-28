package epaper

import (
	"fmt"
	"image"
	"time"

	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/gpio/gpioreg"
	"periph.io/x/periph/conn/physic"
	"periph.io/x/periph/conn/spi"
	"periph.io/x/periph/conn/spi/spireg"
	"periph.io/x/periph/host"
)

// Epd is basic struc for Waveshare eps2in13bc
type Epd struct {
	Width   int
	Height  int
	port    spi.PortCloser
	spiConn spi.Conn
	rstPin  gpio.PinIO
	dcPin   gpio.PinIO
	csPin   gpio.PinIO
	busyPin gpio.PinIO
}

// CreateEpd is constructor for Epd
func CreateEpd() Epd {
	e := Epd{
		Width:  122,
		Height: 250,
	}

	var err error

	host.Init()

	// SPI
	e.port, err = spireg.Open("")
	if err != nil {
		fmt.Println(err)
	}

	e.spiConn, err = e.port.Connect(3000000*physic.Hertz, 0b00, 8)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(e.spiConn)

	// GPIO - read

	e.rstPin = gpioreg.ByName("GPIO17")  // out
	e.dcPin = gpioreg.ByName("GPIO25")   // out
	e.csPin = gpioreg.ByName("GPIO8")    // out
	e.busyPin = gpioreg.ByName("GPIO24") // in

	return e
}

// Close is closing pariph.io port
func (e *Epd) Close() {
	e.port.Close()
}

// reset epd
func (e *Epd) reset() {
	e.rstPin.Out(true)
	time.Sleep(200 * time.Millisecond)
	e.rstPin.Out(false)
	time.Sleep(5 * time.Millisecond)
	e.rstPin.Out(true)
	time.Sleep(200 * time.Millisecond)
}

// sendCommand sets DC ping low and sends byte over SPI
func (e *Epd) sendCommand(command byte) {
	e.dcPin.Out(false)
	e.csPin.Out(false)
	c := []byte{command}
	r := make([]byte, len(c))
	e.spiConn.Tx(c, r)
	e.csPin.Out(true)
	e.readBusy()
}

// sendData sets DC ping high and sends byte over SPI
func (e *Epd) sendData(data byte) {
	e.dcPin.Out(true)
	e.csPin.Out(false)
	c := []byte{data}
	r := make([]byte, len(c))
	e.spiConn.Tx(c, r)
	e.csPin.Out(true)
	e.readBusy()
}

// ReadBusy waits for epd
func (e *Epd) readBusy() {
	//
	// 0: idle
	// 1: busy
	for e.busyPin.Read() == gpio.High {
		time.Sleep(100 * time.Millisecond)
	}
}

// Sleep powers off the epd
func (e *Epd) Sleep() {
	e.executeCommandAndLog(0x10, "DEEP_SLEEP", []byte{0x03})
	time.Sleep(100 * time.Millisecond)
}

// Display sends an image to epd
func (e *Epd) Display(image []byte) {
	lineWidth := e.Width / 8
	if e.Width%8 != 0 {
		lineWidth++
	}
	e.sendCommand(0x24)
	for j := 0; j < e.Height; j++ {
		for i := 0; i < lineWidth; i++ {
			e.sendData(image[i+j*lineWidth])
		}
	}
	e.TurnDisplayOn()
}

// TurnDisplayOn turn the epd on
func (e *Epd) TurnDisplayOn() {
	e.sendCommand(0x22)
	e.sendData(0xC7)
	e.sendCommand(0x20)
	e.readBusy()
}

var lutFullUpdate = []byte{
	0x80, 0x4A, 0x40, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x40, 0x4A, 0x80, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x80, 0x4A, 0x40, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x40, 0x4A, 0x80, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0xF, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0xF, 0x0, 0x0, 0xF, 0x0, 0x0, 0x2,
	0xF, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x0, 0x0, 0x0,
	0x22, 0x17, 0x41, 0x0, 0x32, 0x36,
}

// Init starts the epd
func (e *Epd) Init() {
	e.reset()

	e.readBusy()
	e.executeCommandAndLog(0x12, "SOFT_RESET", nil)
	e.readBusy()

	e.executeCommandAndLog(0x01, "DRIVER_OUTPUT_CONTROL", []byte{0xF9, 0x00, 0x00})
	e.executeCommandAndLog(0x11, "DATA_ENTRY_MODE", []byte{0x03})

	//Equivalent to Set Window in python driver
	e.executeCommandAndLog(0x44, "SET_X-RAM_START_END_POSITION - Second data byte 0x0C-->(15+1)*8=128", []byte{0x0, 0x0F})
	e.executeCommandAndLog(0x45, "SET_X-RAM_START_END_POSITION - First data byte 0xF9-->(249+1)=250", []byte{0x0, 0x0, 0xF9, 0x0})
	e.setCursor(0, 0)

	e.executeCommandAndLog(0x3C, "BorderWavefrom", []byte{0x05})
	e.executeCommandAndLog(0x21, "Display update control", []byte{0x00, 0x80})
	e.executeCommandAndLog(0x18, "", []byte{0x80})
	e.readBusy()
	e.setLut(lutFullUpdate)

	fmt.Println("INIT DONE")
	time.Sleep(100 * time.Millisecond)
}

func (e *Epd) executeCommandAndLog(command byte, log string, data []byte) {
	fmt.Println(log)
	e.sendCommand(command)
	for i := 0; i < len(data); i++ {
		e.sendData(data[i])
	}
}

// Clear sets epd display to white (0xFF)
func (e *Epd) Clear() {
	lineWidth := e.Width / 8
	if e.Width%8 != 0 {
		lineWidth++
	}
	e.sendCommand(0x24)
	for i := 0; i < e.Height; i++ {
		for j := 0; j < lineWidth; j++ {
			e.sendData(0xFF)
		}
	}
	e.TurnDisplayOff()
}

// TurnDisplayOff turn the display off
func (e *Epd) TurnDisplayOff() {
	e.sendCommand(0x22)
	e.sendData(0xC7)
	e.sendCommand(0x20)
}

// GetBuffer return the buffer from a RGBA image, this buffer
// should be pass to Display method.
func (e *Epd) GetBuffer(image *image.RGBA) []byte {
	lineWidth := e.Width / 8
	if e.Width%8 != 0 {
		lineWidth++
	}

	size := (lineWidth * e.Height)
	data := make([]byte, size)
	for i := 0; i < len(data); i++ {
		data[i] = 0xFF
	}

	imageWidth := image.Rect.Dx()
	imageHeight := image.Rect.Dy()

	if imageWidth == e.Width && imageHeight == e.Height {
		for y := 0; y < imageHeight; y++ {
			for x := 0; x < imageWidth; x++ {
				if isBlack(image, x, y) {
					pos := imageWidth - x
					data[pos/8+y*lineWidth] &= ^(0x80 >> (pos % 8))
				}
			}
		}
		return data
	}

	if imageWidth == e.Height && imageHeight == e.Width {
		for y := 0; y < imageHeight; y++ {
			for x := 0; x < imageWidth; x++ {
				if isBlack(image, x, y) {
					posx := y
					posy := imageWidth - (e.Height - x - 1) - 1
					data[posx/8+posy*lineWidth] &= ^(0x80 >> (y % 8))
				}
			}
		}
		return data
	}
	fmt.Printf("Can't convert image expected %d %d but having %d %d", lineWidth, e.Height, imageWidth, imageHeight)
	return data
}

func isBlack(image *image.RGBA, x, y int) bool {
	r, g, b, a := getRGBA(image, x, y)
	offset := 10
	return r < 255-offset && g < 255-offset && b < 255-offset && a > offset
}
func getRGBA(image *image.RGBA, x, y int) (int, int, int, int) {
	r, g, b, a := image.At(x, y).RGBA()
	r = r / 257
	g = g / 257
	b = b / 257
	a = a / 257

	return int(r), int(g), int(b), int(a)
}

func (e *Epd) lut(lut []byte) {
	e.sendCommand(0x32)

	for i := 0; i <= 153; i++ {
		e.sendData(lut[i])
	}
	e.readBusy()
}

func (e *Epd) setLut(lut []byte) {
	e.lut(lut)
	e.executeCommandAndLog(0x03f, "", []byte{lut[153]})
	e.executeCommandAndLog(0x03, "gate voltage", []byte{lut[154]})
	e.executeCommandAndLog(0x04, "source voltage: VSH, VSH2, VSL", []byte{lut[155], lut[156], lut[157]})
	e.executeCommandAndLog(0x2c, "vcom", []byte{lut[158]})
}

func (e *Epd) setCursor(x byte, y byte) {
	e.executeCommandAndLog(0x4E, "SET_RAM_X_ADDRESS_COUNTER", []byte{x & 0xFF})
	e.executeCommandAndLog(0x4F, "SET_RAM_Y_ADDRESS_COUNTER", []byte{y & 0xFF, (y >> 8) & 0xFF})

}

// ShiftLeft performs a left bit shift operation on the provided bytes.
// If the bits count is negative, a right bit shift is performed.
func ShiftLeft(data []byte, bits int) {
	n := len(data)
	if bits < 0 {
		bits = -bits
		for i := n - 1; i > 0; i-- {
			data[i] = data[i]>>bits | data[i-1]<<(8-bits)
		}
		data[0] >>= bits
	} else {
		for i := 0; i < n-1; i++ {
			data[i] = data[i]<<bits | data[i+1]>>(8-bits)
		}
		data[n-1] <<= bits
	}
}
