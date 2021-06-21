package main

/*
import (
	"image/color"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
)

type Game struct {
	g             *Client
	width, height int
	count         int

	canvasImage *ebiten.Image
}

func (g *Game) Update() error {
	mx, my := ebiten.CursorPosition()
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		g.g.MouseDown(0, mx, my)
		g.g.MouseUp(0, mx, my)
	}
	return nil
}

// paint draws the brush on the given canvas image at the position (x, y).
func (g *Game) paint(canvas *ebiten.Image, x, y int) {
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(x), float64(y))
	g.canvasImage.DrawImage(canvas, op)
}

func (g *Game) Draw(screen *ebiten.Image) {
	//fmt.Println("Draw:", time.Now())
	screen.Clear()
	screen.DrawImage(g.canvasImage, nil)
	select {
	case bs := <-BitmapCH:
		//t := time.Now()
		for _, b := range bs {
			m := ebiten.NewImage(b.Width, b.Height)
			i := 0
			for y := 0; y < b.Height; y++ {
				for x := 0; x < b.Width; x++ {
					c := color.RGBA{b.Data[i+2], b.Data[i+1], b.Data[i], 255}
					i += 4
					m.Set(x, y, c)
				}
			}
			g.paint(m, b.DestLeft, b.DestTop)
		}
		//fmt.Println("len:", len(bs), ",time:", time.Now().Sub(t))

	default:
	}

}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return g.width, g.height
}

func initUI1(g1 *Client, width, height int) {
	ebiten.SetWindowSize(width, height)
	ebiten.SetWindowTitle("Paint Demo")
	g := &Game{g1, width, height, 0, ebiten.NewImage(width, height)}
	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}*/
