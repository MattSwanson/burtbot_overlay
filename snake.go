package main

import (
	"fmt"
	"image/color"
	"math/rand"
	"time"

	"github.com/MattSwanson/ebiten/v2"
	"github.com/MattSwanson/ebiten/v2/ebitenutil"
	"github.com/MattSwanson/ebiten/v2/inpututil"
	"github.com/MattSwanson/ebiten/v2/text"
)

const (
	gridSize     = 30
	xNumInScreen = screenWidth / gridSize
	yNumInScreen = screenHeight / gridSize
)

const (
	dirNone = iota
	dirLeft
	dirRight
	dirDown
	dirUp
)

type Position struct {
	X int
	Y int
}

type Snake struct {
	moveDirection int
	snakeBody     []Position
	apple         Position
	timer         int
	moveTime      int
	score         int
	bestScore     int
	level         int
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

func newSnake() *Snake {
	s := &Snake{
		apple:     Position{X: gridSize, Y: gridSize},
		moveTime:  7,
		snakeBody: make([]Position, 1),
	}
	s.snakeBody[0].X = xNumInScreen / 2
	s.snakeBody[0].Y = yNumInScreen / 2
	return s
}

func (s *Snake) collidesWithApple() bool {
	return s.snakeBody[0].X == s.apple.X &&
		s.snakeBody[0].Y == s.apple.Y
}

func (s *Snake) collidesWithSelf() bool {
	for _, v := range s.snakeBody[1:] {
		if s.snakeBody[0].X == v.X &&
			s.snakeBody[0].Y == v.Y {
			return true
		}
	}
	return false
}

func (s *Snake) collidesWithWall() bool {
	return s.snakeBody[0].X < 0 ||
		s.snakeBody[0].Y < 0 ||
		s.snakeBody[0].X >= xNumInScreen ||
		s.snakeBody[0].Y >= yNumInScreen
}

func (s *Snake) needsToMoveSnake() bool {
	return s.timer%s.moveTime == 0
}

func (s *Snake) reset() {
	s.apple.X = gridSize
	s.apple.Y = gridSize
	s.moveTime = 7
	s.snakeBody = s.snakeBody[:1]
	s.snakeBody[0].X = xNumInScreen / 2
	s.snakeBody[0].Y = yNumInScreen / 2
	s.score = 0
	s.level = 1
	s.moveDirection = dirNone
}

func (s *Snake) Update(currentInput ebiten.Key) error {
	if inpututil.IsKeyJustPressed(ebiten.KeyLeft) || currentInput == ebiten.KeyLeft {
		if s.moveDirection != dirRight {
			s.moveDirection = dirLeft
		}
	} else if inpututil.IsKeyJustPressed(ebiten.KeyRight) || currentInput == ebiten.KeyRight {
		if s.moveDirection != dirLeft {
			s.moveDirection = dirRight
		}
	} else if inpututil.IsKeyJustPressed(ebiten.KeyDown) || currentInput == ebiten.KeyDown {
		if s.moveDirection != dirUp {
			s.moveDirection = dirDown
		}
	} else if inpututil.IsKeyJustPressed(ebiten.KeyUp) || currentInput == ebiten.KeyUp {
		if s.moveDirection != dirDown {
			s.moveDirection = dirUp
		}
	} else if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		s.reset()
	}

	if s.needsToMoveSnake() {
		if s.collidesWithWall() || s.collidesWithSelf() {
			s.reset()
		}

		if s.collidesWithApple() {
			s.apple.X = rand.Intn(xNumInScreen - 1)
			s.apple.Y = rand.Intn(yNumInScreen - 1)
			s.snakeBody = append(s.snakeBody, Position{
				X: s.snakeBody[len(s.snakeBody)-1].X,
				Y: s.snakeBody[len(s.snakeBody)-1].Y,
			})
			if len(s.snakeBody) > 10 && len(s.snakeBody) < 20 {
				s.level = 2
				s.moveTime = 5
			} else if len(s.snakeBody) > 20 {
				s.level = 3
				s.moveTime = 4
			} else {
				s.level = 1
			}
			s.score++
			if s.bestScore < s.score {
				s.bestScore = s.score
			}
		}

		for i := int64(len(s.snakeBody)) - 1; i > 0; i-- {
			s.snakeBody[i].X = s.snakeBody[i-1].X
			s.snakeBody[i].Y = s.snakeBody[i-1].Y
		}
		switch s.moveDirection {
		case dirLeft:
			s.snakeBody[0].X--
		case dirRight:
			s.snakeBody[0].X++
		case dirDown:
			s.snakeBody[0].Y++
		case dirUp:
			s.snakeBody[0].Y--
		}
	}

	s.timer++

	return nil
}

func (s *Snake) Draw(screen *ebiten.Image) {
	for _, v := range s.snakeBody {
		ebitenutil.DrawRect(screen, float64(v.X*gridSize), float64(v.Y*gridSize), gridSize, gridSize, color.RGBA{0x00, 0xff, 0x00, 0xff})
	}
	ebitenutil.DrawRect(screen, float64(s.apple.X*gridSize), float64(s.apple.Y*gridSize), gridSize, gridSize, color.RGBA{0xFF, 0x00, 0x00, 0xff})

	if s.moveDirection == dirNone {
		instStr := "w, a, s or d to start"
		b := text.BoundString(myFont, instStr)
		textX := screenWidth/2 - b.Dx()/2
		textY := screenHeight/2 - b.Dy()/2
		text.Draw(screen, instStr, myFont, textX, textY, color.RGBA{0xFF, 0xFF, 0xFF, 0xFF})
		text.Draw(screen, instStr, myFont, textX+1, textY+1, color.RGBA{0, 0xFF, 0, 0xFF})
	} else {
		s := fmt.Sprintf("Level: %d Score: %d Best Score: %d", s.level, s.score, s.bestScore)
		text.Draw(screen, s, myFont, 25, 25, color.RGBA{0xFF, 0xFF, 0xFF, 0xFF})
		text.Draw(screen, s, myFont, 25, 25, color.RGBA{0, 0xFF, 0, 0xFF})
	}

}
