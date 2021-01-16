package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"log"
	"os"
)

func max(ints ...uint32) uint32 {
	max := uint32(0)
	for _, i := range ints {
		if i > max {
			max = i
		}
	}
	return max
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), `%s converts a black and white png image into a java int array.
It is suitable for embedding simple bitmap images, but hits limits
on the length of java source code when used with larger images.

NOTE: this program only examines the alpha of each pixel in the image.

Use: %s [file.png] >> SomeClass.java

You'll then need to tweak the java file so that these fields are
declared inside of a class definition.

Additionally, this program will create a preview of the embedded image
in the file out.png in the current working directory.
`, os.Args[0], os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()
	imgFile, err := os.Open(os.Args[1])
	if err != nil {
		log.Fatalf("Failed reading file: %v", err)
	}
	defer imgFile.Close()
	pngData, err := png.Decode(imgFile)
	if err != nil {
		log.Fatalf("Failed decoding image: %v", err)
	}
	bounds := pngData.Bounds()
	fmt.Printf("\tprivate static int height = %d;\n", bounds.Dy())
	fmt.Printf("\tprivate static int width = %d;\n", bounds.Dx())
	fmt.Printf("\tprivate static int[] data = {")
	out := image.NewRGBA(bounds)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			c := pngData.At(x, y)
			_, _, _, a := c.RGBA()
			ratio := (float64(a) / 0xffff)
			asByte := uint8(256 * ratio)
			scaled := int32(asByte) << 24 // ARGB, so alpha needs to be first
			fmt.Printf("%d", scaled)
			if x != bounds.Max.X-1 || y != bounds.Max.Y-1 {
				fmt.Printf(",")
			}
			out.SetRGBA(x, y, color.RGBA{A: asByte})
		}
	}
	fmt.Printf("};")
	outFile, err := os.Create("out.png")
	if err != nil {
		log.Fatalf("failed opening output file: %v", err)
	}
	defer outFile.Close()
	if err := png.Encode(outFile, out); err != nil {
		log.Fatalf("failed encoding output image: %v", err)
	}
}
