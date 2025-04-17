package main

import (
	"errors"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"log"
	"os"
	"strings"
)

func setLSB(value uint8, bit byte) uint8 {
	return (value & 0xFE) | bit
}

func byteToBit(data []byte) []byte {
	bits := make([]byte, 0, len(data)*8)
	for _, b := range data {
		for i := 7; i >= 0; i-- {
			bits = append(bits, (b>>i)&1)
		}
	}
	return bits
}

func bitsToByte(bits []byte) byte {
	var b byte
	for i := 0; i < 8; i++ {
		b = (b << 1) | (bits[i] & 1)
	}
	return b
}

func encryption(message []byte, pwd string) []byte {
	if len(pwd) == 0 {
		return message
	}
	res := make([]byte, len(message))
	for i := 0; i < len(message); i++ {
		res[i] = message[i] ^ pwd[i%len(pwd)]
	}
	return res
}

func encode(fileName, outFileName, message, pwd string, encrypt bool) {

	file, err := os.Open(fileName)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	img, err := png.Decode(file)
	if err != nil {
		log.Fatal(err)
	}

	bounds := img.Bounds()
	rgba := image.NewNRGBA(bounds)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, a := img.At(x, y).RGBA()
			rgba.Set(x, y, color.RGBA{
				R: uint8((r >> 8)),
				G: uint8((g >> 8)),
				B: uint8((b >> 8)),
				A: uint8((a >> 8)),
			})
		}
	}
	msg := []byte(message)
	if encrypt {
		fmt.Printf("Original: % x\n", msg)
		msg = encryption(msg, pwd)
		fmt.Printf("Encrypted: % x\n", msg)
	}
	msgBits := byteToBit(msg)

	lengthBits := byteToBit([]byte{
		byte(len(msg)) >> 24,
		byte(len(msg)) >> 16,
		byte(len(msg)) >> 8,
		byte(len(msg)),
	})
	msgBits = append(lengthBits, msgBits...)

	bitIndex := 0

	for y := bounds.Min.Y; y < bounds.Max.Y && bitIndex < len(msgBits); y++ {
		for x := bounds.Min.X; x < bounds.Max.X && bitIndex < len(msgBits); x++ {
			r, g, b, a := rgba.At(x, y).RGBA()
			pixel := color.RGBA{
				R: uint8(r >> 8),
				G: uint8(g >> 8),
				B: uint8(b >> 8),
				A: uint8(a >> 8),
			}

			if bitIndex < len(msgBits) {
				pixel.R = setLSB(uint8(r>>8), msgBits[bitIndex])
				bitIndex++
			}
			if bitIndex < len(msgBits) {
				pixel.G = setLSB(uint8(g>>8), msgBits[bitIndex])
				bitIndex++
			}
			if bitIndex < len(msgBits) {
				pixel.B = setLSB(uint8(b>>8), msgBits[bitIndex])
				bitIndex++
			}

			rgba.Set(x, y, pixel)
		}
	}

	outfile, err := os.Create(outFileName)
	if err != nil {
		log.Fatal(err)
	}
	defer outfile.Close()
	png.Encode(outfile, rgba)
}

func decode(fileName, pwd string, decrypt bool) {
	file, err := os.Open(fileName)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	img, err := png.Decode(file)
	if err != nil {
		log.Fatal(err)
	}
	bounds := img.Bounds()
	var allBits []byte

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			allBits = append(allBits, byte(r&1), byte(g&1), byte(b&1))
		}
	}

	if len(allBits) < 32 {
		log.Fatal("Not enough bits for length prefix")
	}

	var length uint32
	for i := 0; i < 4; i++ {
		length = (length << 8) | uint32(bitsToByte(allBits[i*8:(i+1)*8]))
	}

	neededBits := int(length) * 8
	if len(allBits)-32 < neededBits {
		log.Fatalf("Need %d bits but only have %d", neededBits, len(allBits)-32)
	}

	msgBits := allBits[32 : 32+neededBits]
	message := make([]byte, 0, length)
	for i := 0; i < len(msgBits); i += 8 {
		message = append(message, bitsToByte(msgBits[i:i+8]))
	}

	if decrypt == true {
		fmt.Printf("Before decryption: % x\n", message)
		message = encryption(message, pwd)
		fmt.Printf("After decryption: % x\n", message)
	}

	fmt.Println("Hidden message: ", string(message))
}

func help() {
	fmt.Println("\n\nThis program allows you to encode or decode any message from a png file")
	fmt.Println("You can also either encrypt or decrypt the message (XOR encryption) using password")
	fmt.Println("\nUsage of the program:")
	fmt.Println("\tstpng encode -i INPUT_PNG -o OUTPUT_PNG -m MESSAGE [-p PASSWORD] [-e BOOL]")
	fmt.Println("\tstpng decode -i INPUT_PNG [-e BOOL] [-p PASSWORD]")
	fmt.Println("\nExample:")
	fmt.Println("\tstpng encode -i input.png -o output.png -m hello -p secr3t -e true\n")
	fmt.Println("\nFlags:")
	flag.PrintDefaults()
	fmt.Println("\nNote: boolean flags are false by default\n\n")
	os.Exit(1)
}

type Config struct {
	Encode   bool
	Decode   bool
	Input    string
	Output   string
	Message  string
	Password string
	Encrypt  bool
	Help     bool
}

func parseFlags() *Config {
	cfg := &Config{}
	flag.BoolVar(&cfg.Encode, "encode", false, "Encode hidden message in png")
	flag.BoolVar(&cfg.Decode, "decode", false, "Decode hidden message from png")
	flag.StringVar(&cfg.Input, "i", "", "Input file path")
	flag.StringVar(&cfg.Output, "o", "output.png", "Output file path")
	flag.StringVar(&cfg.Message, "m", "", "Message to encode")
	flag.StringVar(&cfg.Password, "p", "", "Password for encryption")
	flag.BoolVar(&cfg.Encrypt, "e", false, "Encryption for the message (false by default)")
	flag.BoolVar(&cfg.Help, "help", false, "Program usage")

	flag.Parse()

	return cfg

}

func validateCommand(cfg *Config) error {
	if !cfg.Encode && !cfg.Decode {
		return errors.New("invalid flags, must be '-encode' or '-decode'\nUse -help for program usage")
	}
	return nil
}

func validateEncodeConfig(cfg *Config) error {
	if strings.TrimSpace(cfg.Input) == "" {
		return errors.New("input file is required for encoding")
	}
	if strings.TrimSpace(cfg.Output) == "" {
		return errors.New("output file is required for encoding")
	}
	if strings.TrimSpace(cfg.Message) == "" {
		return errors.New("message is required for encoding")
	}
	if cfg.Encrypt && strings.TrimSpace(cfg.Password) == "" {
		return errors.New("password is required for encryption")
	}
	encode(cfg.Input, cfg.Output, cfg.Message, cfg.Password, cfg.Encrypt)
	return nil
}

func validateDecodeConfig(cfg *Config) error {
	if strings.TrimSpace(cfg.Input) == "" {
		return errors.New("input file is required for decoding")
	}
	if cfg.Encrypt && strings.TrimSpace(cfg.Password) == "" {
		return errors.New("password is required for decryption")
	}
	decode(cfg.Input, cfg.Password, cfg.Encrypt)

	return nil
}

func main() {
	cfg := parseFlags()
	if cfg.Help {
		help()
	}

	if err := validateCommand(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if cfg.Encode {
		if err := validateEncodeConfig(cfg); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\nUse -help for program usage\n", err)
			os.Exit(1)
		}
	}
	if cfg.Decode {
		if err := validateDecodeConfig(cfg); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\nUse -help for program usage\n", err)
			os.Exit(1)
		}
	}
}
