package cube

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"sync"
	"time"

	"github.com/MattSwanson/burtbot_overlay/sound"
	"github.com/MattSwanson/burtbot_overlay/speech"
	rl "github.com/MattSwanson/raylib-go/raylib"
)

const (
	cubeSize                   = 3 // X x X
	lineSize     float32       = 3.0
	drawOffsetX                = 150
	drawOffsetY                = 1100
	shuffleDelay               = 1      // ms between moves
	shuffleTime  time.Duration = 100000 // seconds
)

var running bool
var c *cube
var randoCancelFunc context.CancelFunc
var cubeLock sync.Mutex = sync.Mutex{}
var hasShuffled bool
var moveCount uint64
var currentScore int
var highScore int
var drawSize float32 = 20

func init() {
	j, err := os.ReadFile("cube.json")
	if err != nil {
		log.Println("couldn't load cube save")
		resetCube()
		return
	}
	data := struct {
		Front      []byte
		Back       []byte
		Left       []byte
		Right      []byte
		Top        []byte
		Bottom     []byte
		TotalMoves uint64
		HighScore  int
	}{}
	err = json.Unmarshal(j, &data)
	if err != nil {
		log.Println("couldn't parse cube save")
		resetCube()
	}
	c = &cube{
		front:  data.Front,
		back:   data.Back,
		top:    data.Top,
		bottom: data.Bottom,
		left:   data.Left,
		right:  data.Right,
	}
	moveCount = data.TotalMoves
	highScore = data.HighScore
	hasShuffled = false

}

type cube struct {
	front  []byte
	back   []byte
	top    []byte
	bottom []byte
	left   []byte
	right  []byte
}

//
//[2][5][8] [0][1][2]
//[1][4][7] [3][4][5]
//[0][3][6] [6][7][8]
//
//[0][1][2] [0][1][2] [0][1][2] [0][1][2]
//[3][4][5] [3][4][5] [3][4][5] [3][4][5]
//[6][7][8] [6][7][8] [6][7][8] [6][7][8]
//
//[6][3][0] [0][1][2]
//[7][4][1] [3][4][5]
//[8][5][2] [6][7][8]

func HandleCommand(args []string) {
	switch args[0] {
	case "movecount":
		speech.Speak(fmt.Sprintf("BurtBot has made %d moves on the cube", moveCount), false, false)
	case "start":
		start()
	case "stop":
		stop()
	case "reset":
		resetCube()
	case "shuffle":
		shuffle()
	case "move":
		if !running || len(args) < 2 {
			return
		}
		cubeLock.Lock()
		switch args[1] {
		case "R":
			rotateRightCW()
		case "R'":
			rotateRightCCW()
		case "r":
			rotateRightCW()
			rotateMCW()
		case "r'":
			rotateRightCCW()
			rotateMCCW()
		case "L":
			rotateLeftCW()
		case "L'":
			rotateLeftCCW()
		case "l":
			rotateLeftCW()
			rotateMCCW()
		case "l'":
			rotateLeftCCW()
			rotateMCW()
		case "U":
			rotateTopCW()
		case "U'":
			rotateTopCCW()
		case "u":
			rotateTopCW()
			rotateECCW()
		case "u'":
			rotateTopCCW()
			rotateECW()
		case "D":
			rotateBottomCW()
		case "D'":
			rotateBottomCCW()
		case "d":
			rotateBottomCW()
			rotateECW()
		case "d'":
			rotateBottomCCW()
			rotateECCW()
		case "F":
			rotateFrontCW()
		case "F'":
			rotateFrontCCW()
		case "f":
			rotateFrontCW()
			rotateSCW()
		case "f'":
			rotateFrontCCW()
			rotateSCCW()
		case "M":
			rotateMCW()
		case "M'":
			rotateMCCW()
		case "B":
			rotateBackCW()
		case "B'":
			rotateBackCCW()
		case "b":
			rotateBackCW()
			rotateSCCW()
		case "b'":
			rotateBackCCW()
			rotateSCW()
		case "X":
			rotateXCW()
		case "X'":
			rotateXCCW()
		case "Y":
			rotateYCW()
		case "Y'":
			rotateYCCW()
		case "Z":
			rotateZCW()
		case "Z'":
			rotateZCCW()
		case "E":
			rotateECW()
		case "E'":
			rotateECW()
		case "S":
			rotateSCW()
		case "S'":
			rotateSCW()
		}
		// check for completion
		moveCount++
		if checkCube() {
			fmt.Println("oh joy")
		}
		cubeLock.Unlock()
	}
}

func GetHighScore() int {
	return highScore
}

func GetTotalCubeMoves() uint64 {
	return moveCount
}

func resetCube() {
	c = &cube{
		front:  []byte{'G', 'G', 'G', 'G', 'G', 'G', 'G', 'G', 'G'},
		back:   []byte{'B', 'B', 'B', 'B', 'B', 'B', 'B', 'B', 'B'},
		top:    []byte{'W', 'W', 'W', 'W', 'W', 'W', 'W', 'W', 'W'},
		bottom: []byte{'Y', 'Y', 'Y', 'Y', 'Y', 'Y', 'Y', 'Y', 'Y'},
		left:   []byte{'O', 'O', 'O', 'O', 'O', 'O', 'O', 'O', 'O'},
		right:  []byte{'R', 'R', 'R', 'R', 'R', 'R', 'R', 'R', 'R'},
	}
	moveCount = 0
	highScore = 0
	hasShuffled = false
}

// TODO Can we make a version of this which returns a level of completion?
// I'd like to slow down the shuffling mess when it's like one move away??
func checkCube() bool {
	complete := true
	for i := 0; i < len(c.front); i++ {
		if c.front[i] != c.front[4] || c.back[i] != c.back[4] || c.top[i] != c.top[4] ||
			c.bottom[i] != c.bottom[4] || c.left[i] != c.left[4] || c.right[i] != c.right[4] {
			return false
		}
	}
	hasShuffled = false
	return complete
}

// Simply count the matches to each center cube face[4]
// A solved cube would be 48
// One move away would be 36 - actually... this could be 2 moves away too...
//   - unless we track how close certain parts are to their target face,
//     if that makes any sense
//
// A score between 36 and 48 could actually require more moves
//
//	and some may be un obtainable
//
// 24 could be two moves and also 26
// There's probably a better way to do this, but I just want to slow the shuffle
// When it's one move away
func scoreCube() int {
	var score int
	for i := 0; i < len(c.front); i++ {
		if i == 4 {
			continue
		}
		if c.front[i] == c.front[4] {
			score++
		}
		if c.back[i] == c.back[4] {
			score++
		}
		if c.left[i] == c.left[4] {
			score++
		}
		if c.right[i] == c.right[4] {
			score++
		}
		if c.top[i] == c.top[4] {
			score++
		}
		if c.bottom[i] == c.bottom[4] {
			score++
		}
	}
	return score
}

func rotateFaceCW(face []byte) {
	face[0], face[2], face[8], face[6] = face[6], face[0], face[2], face[8]
	face[1], face[5], face[7], face[3] = face[3], face[1], face[5], face[7]
}

func rotateFaceCCW(face []byte) {
	face[0], face[2], face[8], face[6] = face[2], face[8], face[6], face[0]
	face[1], face[5], face[7], face[3] = face[5], face[7], face[3], face[1]
}

func rotateFrontCW() {
	rotateFaceCW(c.front)
	c.top[6], c.top[7], c.top[8], c.right[0], c.right[3], c.right[6], c.bottom[0], c.bottom[1], c.bottom[2],
		c.left[2], c.left[5], c.left[8] = c.left[8], c.left[5], c.left[2], c.top[6], c.top[7], c.top[8],
		c.right[6], c.right[3], c.right[0], c.bottom[0], c.bottom[1], c.bottom[2]
}

func rotateFrontCCW() {
	rotateFaceCCW(c.front)
	c.top[6], c.top[7], c.top[8], c.right[0], c.right[3], c.right[6], c.bottom[0], c.bottom[1], c.bottom[2],
		c.left[2], c.left[5], c.left[8] = c.right[0], c.right[3], c.right[6], c.bottom[2], c.bottom[1], c.bottom[0],
		c.left[2], c.left[5], c.left[8], c.top[8], c.top[7], c.top[6]
}

func rotateTopCW() {
	rotateFaceCW(c.top)
	c.back[0], c.back[1], c.back[2], c.right[0], c.right[1], c.right[2], c.front[0], c.front[1], c.front[2],
		c.left[0], c.left[1], c.left[2] = c.left[0], c.left[1], c.left[2], c.back[0], c.back[1], c.back[2],
		c.right[0], c.right[1], c.right[2], c.front[0], c.front[1], c.front[2]
}

func rotateTopCCW() {
	rotateFaceCCW(c.top)
	c.back[0], c.back[1], c.back[2], c.right[0], c.right[1], c.right[2], c.front[0], c.front[1], c.front[2], c.left[0], c.left[1], c.left[2] =
		c.right[0], c.right[1], c.right[2], c.front[0], c.front[1], c.front[2], c.left[0], c.left[1], c.left[2], c.back[0], c.back[1], c.back[2]
}

func rotateRightCW() {
	rotateFaceCW(c.right)
	c.top[2], c.top[5], c.top[8], c.back[0], c.back[3], c.back[6], c.bottom[2], c.bottom[5], c.bottom[8],
		c.front[2], c.front[5], c.front[8] = c.front[2], c.front[5], c.front[8], c.top[8], c.top[5], c.top[2],
		c.back[6], c.back[3], c.back[0], c.bottom[2], c.bottom[5], c.bottom[8]
}

func rotateRightCCW() {
	rotateFaceCCW(c.right)
	c.top[2], c.top[5], c.top[8], c.back[0], c.back[3], c.back[6], c.bottom[2], c.bottom[5], c.bottom[8], c.front[2], c.front[5], c.front[8] =
		c.back[6], c.back[3], c.back[0], c.bottom[8], c.bottom[5], c.bottom[2], c.front[2], c.front[5], c.front[8], c.top[2], c.top[5], c.top[8]
}

func rotateLeftCW() {
	rotateFaceCW(c.left)
	c.top[0], c.top[3], c.top[6], c.front[0], c.front[3], c.front[6], c.bottom[0], c.bottom[3], c.bottom[6],
		c.back[2], c.back[5], c.back[8] = c.back[8], c.back[5], c.back[2], c.top[0], c.top[3], c.top[6],
		c.front[0], c.front[3], c.front[6], c.bottom[6], c.bottom[3], c.bottom[0]
}

func rotateLeftCCW() {
	rotateFaceCCW(c.left)
	c.top[0], c.top[3], c.top[6], c.front[0], c.front[3], c.front[6], c.bottom[0], c.bottom[3], c.bottom[6], c.back[2], c.back[5], c.back[8] =
		c.front[0], c.front[3], c.front[6], c.bottom[0], c.bottom[3], c.bottom[6], c.back[8], c.back[5], c.back[2], c.top[6], c.top[3], c.top[0]
}

func rotateBottomCW() {
	rotateFaceCW(c.bottom)
	// front -> right -> back -> left -> front ...
	c.front[6], c.front[7], c.front[8], c.right[6], c.right[7], c.right[8], c.back[6], c.back[7], c.back[8],
		c.left[6], c.left[7], c.left[8] = c.left[6], c.left[7], c.left[8], c.front[6], c.front[7], c.front[8],
		c.right[6], c.right[7], c.right[8], c.back[6], c.back[7], c.back[8]
}

func rotateBottomCCW() {
	rotateFaceCCW(c.bottom)
	// front -> left -> back -> right -> front ...
	c.front[6], c.front[7], c.front[8], c.right[6], c.right[7], c.right[8], c.back[6], c.back[7], c.back[8], c.left[6], c.left[7], c.left[8] =
		c.right[6], c.right[7], c.right[8], c.back[6], c.back[7], c.back[8], c.left[6], c.left[7], c.left[8], c.front[6], c.front[7], c.front[8]
}

func rotateBackCW() {
	rotateFaceCW(c.back)
	// top -> left -> bottom -> right -> top...
	c.top[0], c.top[1], c.top[2], c.left[0], c.left[3], c.left[6], c.bottom[6], c.bottom[7], c.bottom[8],
		c.right[2], c.right[5], c.right[8] = c.right[2], c.right[5], c.right[8], c.top[2], c.top[1], c.top[0],
		c.left[0], c.left[3], c.left[6], c.bottom[8], c.bottom[7], c.bottom[6]
}

func rotateBackCCW() {
	rotateFaceCCW(c.back)
	// top -> right -> bottom -> left -> top...
	c.top[0], c.top[1], c.top[2], c.left[0], c.left[3], c.left[6], c.bottom[6], c.bottom[7], c.bottom[8], c.right[2], c.right[5], c.right[8] =
		c.left[6], c.left[3], c.left[0], c.bottom[6], c.bottom[7], c.bottom[8], c.right[8], c.right[5], c.right[2], c.top[0], c.top[1], c.top[2]
}

func rotateYCW() {
	// right -> front -> left -> back -> right...
	rotateTopCW()
	rotateECCW()
	rotateBottomCCW()
}

func rotateYCCW() {
	// left -> front -> right -> back -> left...
	rotateTopCCW()
	rotateECW()
	rotateBottomCW()
}

func rotateXCW() {
	// top -> back -> bottom -> front -> top...
	rotateLeftCCW()
	rotateRightCW()
	rotateMCW()
}

func rotateXCCW() {
	// top -> front -> bottom -> back -> top...
	rotateLeftCW()
	rotateRightCCW()
	rotateMCCW()
}

func rotateZCW() {
	// top -> right -> bottom -> left -> top...
	rotateFrontCW()
	rotateSCW()
	rotateBackCCW()
}

func rotateZCCW() {
	rotateFrontCCW()
	rotateSCCW()
	rotateBackCW()
}

func rotateMCW() {
	// top -> back -> bottom -> front -> top...
	c.top[1], c.top[4], c.top[7], c.back[1], c.back[4], c.back[7], c.bottom[1], c.bottom[4], c.bottom[7],
		c.front[1], c.front[4], c.front[7] = c.front[1], c.front[4], c.front[7], c.top[7], c.top[4], c.top[1],
		c.back[7], c.back[4], c.back[1], c.bottom[1], c.bottom[4], c.bottom[7]
}

func rotateMCCW() {
	// top -> front -> bottom -> back -> top...
	c.top[1], c.top[4], c.top[7], c.back[1], c.back[4], c.back[7], c.bottom[1], c.bottom[4], c.bottom[7], c.front[1], c.front[4], c.front[7] =
		c.back[7], c.back[4], c.back[1], c.bottom[7], c.bottom[4], c.bottom[1], c.front[1], c.front[4], c.front[7], c.top[1], c.top[4], c.top[7]
}

func rotateECW() {
	// this is clockwise relative to the bottom
	// front -> right -> back -> left -> front...
	c.front[3], c.front[4], c.front[5], c.right[3], c.right[4], c.right[5], c.back[3], c.back[4], c.back[5],
		c.left[3], c.left[4], c.left[5] = c.left[3], c.left[4], c.left[5], c.front[3], c.front[4], c.front[5],
		c.right[3], c.right[4], c.right[5], c.back[3], c.back[4], c.back[5]
}

func rotateECCW() {
	// this is clockwise relative to the bottom
	// front -> left -> back -> right -> front...
	c.front[3], c.front[4], c.front[5], c.right[3], c.right[4], c.right[5], c.back[3], c.back[4], c.back[5], c.left[3], c.left[4], c.left[5] =
		c.right[3], c.right[4], c.right[5], c.back[3], c.back[4], c.back[5], c.left[3], c.left[4], c.left[5], c.front[3], c.front[4], c.front[5]
}

func rotateSCW() {
	// top -> right -> bottom -> left -> top...
	c.top[3], c.top[4], c.top[5], c.right[1], c.right[4], c.right[7], c.bottom[3], c.bottom[4], c.bottom[5],
		c.left[1], c.left[4], c.left[7] = c.left[7], c.left[4], c.left[1], c.top[3], c.top[4], c.top[5],
		c.right[7], c.right[4], c.right[1], c.bottom[3], c.bottom[4], c.bottom[5]
}

func rotateSCCW() {
	// top -> left -> bottom -> right -> top...
	c.top[3], c.top[4], c.top[5], c.right[1], c.right[4], c.right[7], c.bottom[3], c.bottom[4], c.bottom[5], c.left[1], c.left[4], c.left[7] =
		c.right[1], c.right[4], c.right[7], c.bottom[3], c.bottom[4], c.bottom[5], c.left[1], c.left[4], c.left[7], c.top[3], c.top[4], c.top[5]
}

func start() {
	if running {
		return
	}
	running = true
}

func stop() {
	if !running {
		return
	}
	if randoCancelFunc != nil {
		randoCancelFunc()
	}
	randoCancelFunc = nil
	running = false
}

func Draw() {
	if !running {
		return
	}

	for k, v := range c.top {
		drawPosBaseY := drawOffsetY - drawSize*cubeSize
		cubeOffset := float32(cubeSize - 1 - k/cubeSize)
		rl.DrawTriangleStrip([]rl.Vector2{
			{X: drawOffsetX + drawSize*float32(k%cubeSize) + drawSize/2.0 + cubeOffset*drawSize/2.0,
				Y: drawPosBaseY + drawSize*float32(k/cubeSize) + drawSize/2.0 + cubeOffset*drawSize/2.0}, // TL

			{X: drawOffsetX + drawSize*float32(k%cubeSize) + cubeOffset*drawSize/2.0,
				Y: drawPosBaseY + drawSize*float32(k/cubeSize) + drawSize + (cubeOffset * drawSize / 2.0)}, // BL

			{X: (drawOffsetX + drawSize*float32(k%cubeSize) + drawSize) + drawSize/2.0 + float32(cubeOffset*drawSize/2.0),
				Y: drawPosBaseY + float32(drawSize*float32(k/cubeSize)) + drawSize/2.0 + float32(cubeOffset*drawSize/2.0)}, // TR

			{X: float32(drawOffsetX+drawSize*float32(k%cubeSize)+drawSize) + float32(cubeOffset*drawSize/2.0),
				Y: drawPosBaseY + float32(drawSize*float32(k/cubeSize)+drawSize) + float32(cubeOffset*drawSize/2.0)}, // BR
		}, getColor(v))
	}
	// for k, v := range c.bottom {
	// 	rl.DrawRectangle(cubeSize*drawSize+drawSize*int32(k%cubeSize), 2*cubeSize*drawSize+drawSize*int32(k/cubeSize), drawSize, drawSize, getColor(v))
	// }
	for k, v := range c.front {
		rl.DrawRectangle(int32(drawOffsetX+drawSize*float32(k%cubeSize)), int32(drawOffsetY+drawSize*float32(k/cubeSize)), int32(drawSize), int32(drawSize), getColor(v))
	}
	// for k, v := range c.back {
	// 	rl.DrawRectangle(3*cubeSize*drawSize+drawSize*int32(k%cubeSize), cubeSize*drawSize+drawSize*int32(k/cubeSize), drawSize, drawSize, getColor(v))
	// }
	// for k, v := range c.left {
	// 	rl.DrawRectangle(drawSize*int32(k%cubeSize), cubeSize*drawSize+drawSize*int32(k/cubeSize), drawSize, drawSize, getColor(v))
	// }
	for k, v := range c.right {
		//rl.DrawRectangle(2*cubeSize*drawSize+drawSize*int32(k%cubeSize), cubeSize*drawSize+drawSize*int32(k/cubeSize), drawSize, drawSize, getColor(v))
		rl.DrawTriangleStrip([]rl.Vector2{
			{X: float32(drawOffsetX + cubeSize*drawSize + drawSize*float32(k%cubeSize)/2.0), Y: drawOffsetY - cubeSize*drawSize + float32(cubeSize*drawSize+drawSize*float32(k/cubeSize)) - float32(float32(k%cubeSize)*drawSize/2.0)},                               // TL
			{X: float32(drawOffsetX + cubeSize*drawSize + drawSize*float32(k%cubeSize)/2.0), Y: drawOffsetY - cubeSize*drawSize + float32(cubeSize*drawSize+drawSize*float32(k/cubeSize)+drawSize) - float32(float32(k%cubeSize)*drawSize/2.0)},                      // BL
			{X: float32(drawOffsetX + cubeSize*drawSize + drawSize*float32(k%cubeSize)/2.0 + drawSize/2.0), Y: drawOffsetY - cubeSize*drawSize + float32(cubeSize*drawSize+drawSize*float32(k/cubeSize)) - drawSize/2.0 - float32(float32(k%cubeSize)*drawSize/2.0)}, // TR
			{X: float32(drawOffsetX + cubeSize*drawSize + drawSize*float32(k%cubeSize)/2.0 + drawSize/2.0), Y: drawOffsetY - cubeSize*drawSize + float32(cubeSize*drawSize+drawSize*float32(k/cubeSize)) + drawSize/2.0 - float32(float32(k%cubeSize)*drawSize/2.0)}, // BR
		}, getColor(v))
	}

	// Front horiz
	rl.DrawLineEx(rl.Vector2{X: drawOffsetX, Y: drawOffsetY}, rl.Vector2{X: drawOffsetX + cubeSize*drawSize, Y: drawOffsetY}, lineSize, rl.Black)
	rl.DrawLineEx(rl.Vector2{X: drawOffsetX, Y: drawOffsetY + drawSize}, rl.Vector2{X: drawOffsetX + cubeSize*drawSize, Y: drawOffsetY + drawSize}, lineSize, rl.Black)
	rl.DrawLineEx(rl.Vector2{X: drawOffsetX, Y: drawOffsetY + 2*drawSize}, rl.Vector2{X: drawOffsetX + cubeSize*drawSize, Y: drawOffsetY + 2*drawSize}, lineSize, rl.Black)
	rl.DrawLineEx(rl.Vector2{X: drawOffsetX, Y: drawOffsetY + 3*drawSize}, rl.Vector2{X: drawOffsetX + cubeSize*drawSize, Y: drawOffsetY + 3*drawSize}, lineSize, rl.Black)

	// Front Vert
	rl.DrawLineEx(rl.Vector2{X: drawOffsetX, Y: drawOffsetY}, rl.Vector2{X: drawOffsetX, Y: drawOffsetY + 3*drawSize}, lineSize, rl.Black)
	rl.DrawLineEx(rl.Vector2{X: drawOffsetX + drawSize, Y: drawOffsetY}, rl.Vector2{X: drawOffsetX + drawSize, Y: drawOffsetY + 3*drawSize}, lineSize, rl.Black)
	rl.DrawLineEx(rl.Vector2{X: drawOffsetX + 2*drawSize, Y: drawOffsetY}, rl.Vector2{X: drawOffsetX + 2*drawSize, Y: drawOffsetY + 3*drawSize}, lineSize, rl.Black)
	rl.DrawLineEx(rl.Vector2{X: drawOffsetX + 3*drawSize, Y: drawOffsetY}, rl.Vector2{X: drawOffsetX + 3*drawSize, Y: drawOffsetY + 3*drawSize}, lineSize, rl.Black)

	// Side horiz
	rl.DrawLineEx(rl.Vector2{X: drawOffsetX + cubeSize*drawSize, Y: drawOffsetY + drawSize}, rl.Vector2{X: drawOffsetX + 9*drawSize/2, Y: drawOffsetY - drawSize/2}, lineSize, rl.Black)
	rl.DrawLineEx(rl.Vector2{X: drawOffsetX + cubeSize*drawSize, Y: drawOffsetY + 2*drawSize}, rl.Vector2{X: drawOffsetX + 9*drawSize/2, Y: drawOffsetY + drawSize/2}, lineSize, rl.Black)
	rl.DrawLineEx(rl.Vector2{X: drawOffsetX + cubeSize*drawSize, Y: drawOffsetY + 3*drawSize}, rl.Vector2{X: drawOffsetX + 9*drawSize/2, Y: drawOffsetY + 3*drawSize/2}, lineSize, rl.Black)

	// Side Vert
	rl.DrawLineEx(rl.Vector2{X: drawOffsetX + 7*drawSize/2, Y: drawOffsetY - drawSize/2}, rl.Vector2{X: drawOffsetX + 7*drawSize/2, Y: drawOffsetY + 5*drawSize/2}, lineSize, rl.Black)
	rl.DrawLineEx(rl.Vector2{X: drawOffsetX + 4*drawSize, Y: drawOffsetY - drawSize}, rl.Vector2{X: drawOffsetX + 4*drawSize, Y: drawOffsetY + 2*drawSize}, lineSize, rl.Black)
	rl.DrawLineEx(rl.Vector2{X: drawOffsetX + 9*drawSize/2, Y: drawOffsetY - 3*drawSize/2}, rl.Vector2{X: drawOffsetX + 9*drawSize/2, Y: drawOffsetY + 3*drawSize/2}, lineSize, rl.Black)

	// Top Vert
	rl.DrawLineEx(rl.Vector2{X: drawOffsetX, Y: drawOffsetY}, rl.Vector2{X: drawOffsetX + 3*drawSize/2, Y: drawOffsetY - 3*drawSize/2}, lineSize, rl.Black)
	rl.DrawLineEx(rl.Vector2{X: drawOffsetX + drawSize, Y: drawOffsetY}, rl.Vector2{X: drawOffsetX + 5*drawSize/2, Y: drawOffsetY - 3*drawSize/2}, lineSize, rl.Black)
	rl.DrawLineEx(rl.Vector2{X: drawOffsetX + 2*drawSize, Y: drawOffsetY}, rl.Vector2{X: drawOffsetX + 7*drawSize/2, Y: drawOffsetY - 3*drawSize/2}, lineSize, rl.Black)
	rl.DrawLineEx(rl.Vector2{X: drawOffsetX + 3*drawSize, Y: drawOffsetY}, rl.Vector2{X: drawOffsetX + 9*drawSize/2, Y: drawOffsetY - 3*drawSize/2}, lineSize, rl.Black)

	// Top Horiz
	rl.DrawLineEx(rl.Vector2{X: drawOffsetX + 3*drawSize/2, Y: drawOffsetY - 3*drawSize/2}, rl.Vector2{X: drawOffsetX + cubeSize*drawSize + 3*drawSize/2, Y: drawOffsetY - 3*drawSize/2}, lineSize, rl.Black)
	rl.DrawLineEx(rl.Vector2{X: drawOffsetX + drawSize, Y: drawOffsetY - drawSize}, rl.Vector2{X: drawOffsetX + cubeSize*drawSize + drawSize, Y: drawOffsetY - drawSize}, lineSize, rl.Black)
	rl.DrawLineEx(rl.Vector2{X: drawOffsetX + drawSize/2, Y: drawOffsetY - drawSize/2}, rl.Vector2{X: drawOffsetX + cubeSize*drawSize + drawSize/2, Y: drawOffsetY - drawSize/2}, lineSize, rl.Black)

	rl.DrawText(fmt.Sprintf("Moves: %d", moveCount), drawOffsetX, drawOffsetY-50, 18, rl.Orange)
}

func getColor(b byte) rl.Color {
	var color rl.Color
	switch b {
	case 'G':
		color = rl.Green
	case 'W':
		color = rl.White
	case 'R':
		color = rl.Red
	case 'O':
		color = rl.Orange
	case 'B':
		color = rl.Blue
	case 'Y':
		color = rl.Yellow
	}
	return color
}

func shuffle() {
	if randoCancelFunc != nil {
		return
	}
	var c context.Context
	c, randoCancelFunc = context.WithCancel(context.Background())
	cubeLock.Lock()
	go func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				cubeLock.Unlock()
				return
			default:
				drawSize = 20
				r := rand.Intn(20)
				switch r {
				case 0:
					rotateFrontCW()
				case 1:
					rotateTopCW()
				case 2:
					rotateRightCW()
				case 3:
					rotateLeftCW()
				case 4:
					rotateMCW()
				case 5:
					rotateBackCW()
				case 6:
					rotateBottomCW()
				case 7:
					rotateFrontCCW()
				case 8:
					rotateTopCCW()
				case 9:
					rotateRightCCW()
				case 10:
					rotateLeftCCW()
				case 11:
					rotateMCCW()
				case 12:
					rotateBackCCW()
				case 13:
					rotateBottomCCW()
				case 14:
					rotateECW()
				case 15:
					rotateECCW()
				case 16:
					rotateMCW()
				case 17:
					rotateMCCW()
				case 18:
					rotateSCW()
				case 19:
					rotateSCCW()
				}
				moveCount++
				currentScore = scoreCube()
				if currentScore > highScore {
					highScore = currentScore
				}
				var timeToWait time.Duration = 0
				if currentScore == 48 {
					drawSize = 80
					fmt.Println("OMG IT DEIFN THSK WHOW")
					hasShuffled = true
					cubeLock.Unlock()
					return
				}
				if currentScore == 36 {
					drawSize = 80
					sound.Play("sosumi")
					timeToWait = 10000000000
					time.Sleep(timeToWait * time.Nanosecond)
				}
				//time.Sleep(timeToWait * time.Nanosecond)
			}
		}
	}(c)

	go func() {
		time.Sleep(shuffleTime * time.Second)
		randoCancelFunc()
		randoCancelFunc = nil
		hasShuffled = true
	}()
}

func SaveCube() {
	if randoCancelFunc != nil {
		randoCancelFunc()
	}
	data := struct {
		Front      []byte
		Back       []byte
		Left       []byte
		Right      []byte
		Top        []byte
		Bottom     []byte
		TotalMoves uint64
		HighScore  int
	}{
		Front:      c.front,
		Back:       c.back,
		Left:       c.left,
		Right:      c.right,
		Top:        c.top,
		Bottom:     c.bottom,
		TotalMoves: moveCount,
		HighScore:  highScore,
	}
	json, _ := json.Marshal(data)
	if err := os.WriteFile("cube.json", json, 0644); err != nil {
		log.Println(err.Error())
	}
}
