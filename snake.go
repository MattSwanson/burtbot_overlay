package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/MattSwanson/burtbot_overlay/sound"
	rl "github.com/MattSwanson/raylib-go/raylib"
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
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

func newSnake() *Snake {
	s := &Snake{
		apple:     Position{X: gridSize, Y: gridSize},
		moveTime:  initialMoveSpeed,
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

func (s *Snake) Update(currentInput int) error {
	if rl.IsKeyPressed(rl.KeyLeft) || currentInput == rl.KeyLeft {
		s.nextMove = dirLeft
	} else if rl.IsKeyPressed(rl.KeyRight) || currentInput == rl.KeyRight {
		s.nextMove = dirRight
	} else if rl.IsKeyPressed(rl.KeyDown) || currentInput == rl.KeyDown {
		s.nextMove = dirDown
	} else if rl.IsKeyPressed(rl.KeyUp) || currentInput == rl.KeyUp {
		s.nextMove = dirUp
	} else if rl.IsKeyPressed(rl.KeyEscape) {
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
			sound.Play("zap")
			s.reset()
		}

		if s.collidesWithApple() {
			sound.Play("squeek")
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

func (s *Snake) Draw() {
	for _, v := range s.snakeBody {
		// ebitenutil.DrawRect(screen, float64(v.X*gridSize), float64(v.Y*gridSize), gridSize, gridSize, color.RGBA{0x00, 0xff, 0x00, 0xff})
		rl.DrawRectangle(int32(v.X*gridSize), int32(v.Y*gridSize), gridSize, gridSize, rl.Color{R: 0x00, G: 0xff, B: 0x00, A: 0xff})
	}
	// ebitenutil.DrawRect(screen, float64(s.apple.X*gridSize), float64(s.apple.Y*gridSize), gridSize, gridSize, color.RGBA{0xFF, 0x00, 0x00, 0xff})
	rl.DrawRectangle(int32(s.apple.X*gridSize), int32(s.apple.Y*gridSize), gridSize, gridSize, rl.Color{R: 0xff, G: 0x00, B: 0x00, A: 0xff})
	if s.moveDirection == dirNone {
		instStr := "w, a, s or d to start"
		b := rl.MeasureTextEx(rl.GetFontDefault(), instStr, 96, 0)
		textX := screenWidth/2 - b.X/2
		textY := screenHeight/2 - b.Y/2
		//text.Draw(screen, instStr, myFont, textX, textY, color.RGBA{0xFF, 0xFF, 0xFF, 0xFF})
		rl.DrawText(instStr, int32(textX), int32(textY), 96, rl.White)
		rl.DrawText(instStr, int32(textX+1), int32(textY+1), 96, rl.Green)
		//text.Draw(screen, instStr, myFont, textX+1, textY+1, color.RGBA{0, 0xFF, 0, 0xFF})
	} else {
		s := fmt.Sprintf("Level: %d Score: %d Best Score: %d", s.level, s.score, s.bestScore)
		//text.Draw(screen, s, myFont, 25, 25, color.RGBA{0xFF, 0xFF, 0xFF, 0xFF})
		rl.DrawText(s, 25, 25, 96, rl.White)
		rl.DrawText(s, 26, 26, 96, rl.Green)
		//text.Draw(screen, s, myFont, 25, 25, color.RGBA{0, 0xFF, 0, 0xFF})
	}
}

func (s *Snake) SetGameSpeed(speed int) {
	if speed <= 0 {
		return
	}
	s.moveTime = speed
}
