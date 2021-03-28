package main

import (
	"fmt"
	"image/color"
	"math/rand"
	"time"

	"github.com/MattSwanson/ebiten/v2"
	"github.com/MattSwanson/ebiten/v2/audio"
	"github.com/MattSwanson/ebiten/v2/ebitenutil"
	"github.com/MattSwanson/ebiten/v2/inpututil"
	"github.com/MattSwanson/ebiten/v2/text"
)

const (
	gridSize         = 30
	xNumInScreen     = screenWidth / gridSize
	yNumInScreen     = screenHeight / gridSize
	initialMoveSpeed = 35
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
	nextMove      int
	snakeBody     []Position
	apple         Position
	timer         int
	moveTime      int
	score         int
	bestScore     int
	level         int
	sounds        map[string]*audio.Player
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

func newSnake(sounds map[string]*audio.Player) *Snake {
	s := &Snake{
		apple:     Position{X: gridSize, Y: gridSize},
		moveTime:  initialMoveSpeed,
		snakeBody: make([]Position, 1),
		sounds:    sounds,
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

// func (s *Snake) collidesWithWall() bool {
// 	return s.snakeBody[0].X < 0 ||
// 		s.snakeBody[0].Y < 0 ||
// 		s.snakeBody[0].X >= xNumInScreen ||
// 		s.snakeBody[0].Y >= yNumInScreen
// }

func (s *Snake) needsToMoveSnake() bool {
	return s.timer%s.moveTime == 0
}

func (s *Snake) reset() {
	s.apple.X = gridSize
	s.apple.Y = gridSize
	s.moveTime = initialMoveSpeed
	s.snakeBody = s.snakeBody[:1]
	s.snakeBody[0].X = xNumInScreen / 2
	s.snakeBody[0].Y = yNumInScreen / 2
	s.score = 0
	s.level = 1
	s.moveDirection = dirNone
}

func (s *Snake) Update(currentInput ebiten.Key) error {
	if inpututil.IsKeyJustPressed(ebiten.KeyLeft) || currentInput == ebiten.KeyLeft {
		s.nextMove = dirLeft
	} else if inpututil.IsKeyJustPressed(ebiten.KeyRight) || currentInput == ebiten.KeyRight {
		s.nextMove = dirRight
	} else if inpututil.IsKeyJustPressed(ebiten.KeyDown) || currentInput == ebiten.KeyDown {
		s.nextMove = dirDown
	} else if inpututil.IsKeyJustPressed(ebiten.KeyUp) || currentInput == ebiten.KeyUp {
		s.nextMove = dirUp
	} else if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		s.reset()
	}

	if s.needsToMoveSnake() {
		if s.nextMove == dirDown && s.moveDirection != dirUp {
			s.moveDirection = s.nextMove
		} else if s.nextMove == dirUp && s.moveDirection != dirDown {
			s.moveDirection = s.nextMove
		} else if s.nextMove == dirLeft && s.moveDirection != dirRight {
			s.moveDirection = s.nextMove
		} else if s.nextMove == dirRight && s.moveDirection != dirLeft {
			s.moveDirection = s.nextMove
		}

		if s.collidesWithSelf() {
			s.sounds["zap"].Rewind()
			s.sounds["zap"].Play()
			s.reset()
		}

		if s.collidesWithApple() {
			s.sounds["squeek"].Rewind()
			s.sounds["squeek"].Play()
			s.apple.X = rand.Intn(xNumInScreen - 1)
			s.apple.Y = rand.Intn(yNumInScreen - 1)
			s.snakeBody = append(s.snakeBody, Position{
				X: s.snakeBody[len(s.snakeBody)-1].X,
				Y: s.snakeBody[len(s.snakeBody)-1].Y,
			})
			if len(s.snakeBody) > 10 && len(s.snakeBody) < 20 {
				s.level = 2
				s.moveTime = initialMoveSpeed - 5
			} else if len(s.snakeBody) > 20 {
				s.level = 3
				s.moveTime = initialMoveSpeed - 10
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
			if s.snakeBody[0].X < 0 {
				s.snakeBody[0].X = xNumInScreen - 1
			}
		case dirRight:
			s.snakeBody[0].X = (s.snakeBody[0].X + 1) % xNumInScreen
		case dirDown:
			s.snakeBody[0].Y = (s.snakeBody[0].Y + 1) % yNumInScreen
		case dirUp:
			s.snakeBody[0].Y--
			if s.snakeBody[0].Y < 0 {
				s.snakeBody[0].Y = yNumInScreen - 1
			}
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

func (s *Snake) SetGameSpeed(speed int) {
	s.moveTime = speed
}
