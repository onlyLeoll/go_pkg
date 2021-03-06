package avatarbuilder

import (
	"bufio"
	"bytes"
	"errors"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io/ioutil"
	"os"

	"github.com/golang/freetype"
	"golang.org/x/image/font"
)

type FontCenterCalculator interface {
	// CalculateCenterLocation used to calculate center location in different font style
	CalculateCenterLocation(string, *AvatarBuilder) (int, int)
}

const (
	defaultHigh  = 200
	defaultWidth = 200
	defaultFont  = 80
)

type AvatarBuilder struct {
	W        int
	H        int
	fontFile string
	fontSize float64
	bg       color.Color
	fg       color.Color
	ctx      *freetype.Context
	calc     FontCenterCalculator
}

func NewAvatarBuilder(fontFile string, calc FontCenterCalculator) *AvatarBuilder {
	ab := &AvatarBuilder{}
	ab.fontFile = fontFile
	ab.bg, ab.fg = color.White, color.Black
	ab.W, ab.H = defaultHigh, defaultWidth
	ab.fontSize = defaultFont
	ab.calc = calc

	return ab
}

func NewAvatarBuilderWithOption(fontFile string, calc FontCenterCalculator, opt ...BuilderOption) *AvatarBuilder {
	ab := &AvatarBuilder{}
	ab.fontFile = fontFile
	ab.bg, ab.fg = color.White, color.Black
	for _, f := range opt {
		f(ab)
	}
	ab.calc = calc

	return ab
}

func (ab *AvatarBuilder) SetFrontGroundColor(c color.Color) {
	ab.fg = c
}

func (ab *AvatarBuilder) SetBackGroundColor(c color.Color) {
	ab.bg = c
}

func (ab *AvatarBuilder) SetFrontGroundColorHex(hex uint32) {
	ab.fg = ab.hexToRGBA(hex)
}

func (ab *AvatarBuilder) SetBackGroundColorHex(hex uint32) {
	ab.bg = ab.hexToRGBA(hex)
}

func (ab *AvatarBuilder) SetFontSize(size float64) {
	ab.fontSize = size
}

func (ab *AvatarBuilder) SetAvatarSize(w int, h int) {
	ab.W = w
	ab.H = h
}

func (ab *AvatarBuilder) GenerateImageAndSave(s string, outName string) error {
	bs, err := ab.GenerateImage(s)
	if err != nil {
		return err
	}

	// Save that RGBA image to disk.
	outFile, err := os.Create(outName)
	if err != nil {
		return errors.New("create file: " + err.Error())
	}
	defer outFile.Close()

	b := bufio.NewWriter(outFile)
	if _, err := b.Write(bs); err != nil {
		return errors.New("write bytes to file: " + err.Error())

	}
	if err = b.Flush(); err != nil {
		return errors.New("flush image: " + err.Error())
	}

	return nil
}

func (ab *AvatarBuilder) GenerateImage(s string) ([]byte, error) {
	rgba := ab.buildColorImage()
	if ab.ctx == nil {
		if err := ab.buildDrawContext(rgba); err != nil {
			return nil, err
		}
	}

	x, y := ab.calc.CalculateCenterLocation(s, ab)
	pt := freetype.Pt(x, y)
	if _, err := ab.ctx.DrawString(s, pt); err != nil {
		return nil, errors.New("draw string: " + err.Error())
	}

	buf := &bytes.Buffer{}
	if err := png.Encode(buf, rgba); err != nil {
		return nil, errors.New("png encode: " + err.Error())
	}

	return buf.Bytes(), nil
}

func (ab *AvatarBuilder) buildColorImage() *image.RGBA {
	bg := image.NewUniform(ab.bg)
	rgba := image.NewRGBA(image.Rect(0, 0, ab.W, ab.H))
	draw.Draw(rgba, rgba.Bounds(), bg, image.ZP, draw.Src)
	return rgba
}

func (ab *AvatarBuilder) hexToRGBA(h uint32) *color.RGBA {
	rgba := &color.RGBA{
		R: uint8(h >> 16),
		G: uint8((h & 0x00ff00) >> 8),
		B: uint8(h & 0x0000ff),
		A: 255,
	}

	return rgba
}

func (ab *AvatarBuilder) buildDrawContext(rgba *image.RGBA) error {
	// Read the font data.
	fontBytes, err := ioutil.ReadFile(ab.fontFile)
	if err != nil {
		return errors.New("error when open font file:" + err.Error())
	}

	f, err := freetype.ParseFont(fontBytes)
	if err != nil {
		return errors.New("error when parse font file:" + err.Error())
	}

	c := freetype.NewContext()
	c.SetDPI(72)
	c.SetFont(f)
	c.SetFontSize(ab.fontSize)
	c.SetClip(rgba.Bounds())
	c.SetDst(rgba)
	c.SetSrc(image.NewUniform(ab.fg))
	c.SetHinting(font.HintingNone)

	ab.ctx = c
	return nil
}

func (ab *AvatarBuilder) GetFontWidth() int {
	return int(ab.ctx.PointToFixed(ab.fontSize) >> 6)
}
