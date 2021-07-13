package cube

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	rl "github.com/MattSwanson/raylib-go/raylib"
)

const (
	cubeSize = 3 // X x X
	drawSize = 180
)

var running bool
var c *cube
var randoCancelFunc context.CancelFunc

func init() {
	resetCube()
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
//		 [][][]
//		 [][][]
//		 [][][]
//
//[][][] [0][1][2] [][][] [][][]
//[][][] [3][4][5] [][][] [][][]
//[][][] [6][7][8] [][][] [][][]
//
//       [][][]
//		 [][][]
//       [][][]

func HandleCommand(args []string) {
	switch args[0] {
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
		switch args[1] {
		case "R":
			rotateRightCW()
		case "R'":
			rotateRightCW()
			rotateRightCW()
			rotateRightCW()
		case "r":
			rotateRightCW()
			rotateMCW()
		case "r'":
			rotateRightCW()
			rotateRightCW()
			rotateRightCW()
			rotateMCW()
			rotateMCW()
			rotateMCW()
		case "L":
			rotateLeftCW()
		case "L'":
			rotateLeftCW()
			rotateLeftCW()
			rotateLeftCW()
		case "l":
			rotateLeftCW()
			rotateMCW()
			rotateMCW()
			rotateMCW()
		case "l'":
			rotateLeftCW()
			rotateLeftCW()
			rotateLeftCW()
			rotateMCW()
		case "U":
			rotateTopCW()
		case "U'":
			rotateTopCW()
			rotateTopCW()
			rotateTopCW()
		case "u":
			rotateTopCW()
			rotateECW()
			rotateECW()
			rotateECW()
		case "u'":
			rotateTopCW()
			rotateTopCW()
			rotateTopCW()
			rotateECW()
		case "D":
			rotateBottomCW()
		case "D'":
			rotateBottomCW()
			rotateBottomCW()
			rotateBottomCW()
		case "d":
			rotateBottomCW()
			rotateECW()
		case "d'":
			rotateBottomCW()
			rotateBottomCW()
			rotateBottomCW()
			rotateECW()
			rotateECW()
			rotateECW()
		case "F":
			rotateFrontCW()
		case "F'":
			rotateFrontCW()
			rotateFrontCW()
			rotateFrontCW()
		case "f":
			rotateFrontCW()
			rotateSCW()
		case "f'":
			rotateFrontCW()
			rotateFrontCW()
			rotateFrontCW()
			rotateSCW()
			rotateSCW()
			rotateSCW()
		case "M":
			rotateMCW()
		case "M'":
			rotateMCW()
			rotateMCW()
			rotateMCW()
		case "B":
			rotateBackCW()
		case "B'":
			rotateBackCW()
			rotateBackCW()
			rotateBackCW()
		case "b":
			rotateBackCW()
			rotateSCW()
			rotateSCW()
			rotateSCW()
		case "b'":
			rotateBackCW()
			rotateBackCW()
			rotateBackCW()
			rotateSCW()
		case "X":
			rotateXCW()
		case "X'":
			rotateXCW()
			rotateXCW()
			rotateXCW()
		case "Y":
			rotateYCW()
		case "Y'":
			rotateYCW()
			rotateYCW()
			rotateYCW()
		case "Z":
			rotateZCW()
		case "Z'":
			rotateZCW()
			rotateZCW()
			rotateZCW()
		case "E":
			rotateECW()
		case "E'":
			rotateECW()
		case "S":
			rotateSCW()
		case "S'":
			rotateSCW()
		}
		//rotateRightCW()

	}
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
}

func rotateFaceCW(face []byte) {
	face[0], face[2], face[8], face[6] = face[6], face[0], face[2], face[8]
	face[1], face[5], face[7], face[3] = face[3], face[1], face[5], face[7]
}

func rotateFrontCW() {
	rotateFaceCW(c.front)
	c.top[6], c.top[7], c.top[8], c.right[0], c.right[3], c.right[6], c.bottom[0], c.bottom[1], c.bottom[2],
		c.left[2], c.left[5], c.left[8] = c.left[8], c.left[5], c.left[2], c.top[6], c.top[7], c.top[8],
		c.right[6], c.right[3], c.right[0], c.bottom[0], c.bottom[1], c.bottom[2]
}

func rotateTopCW() {
	rotateFaceCW(c.top)
	c.back[0], c.back[1], c.back[2], c.right[0], c.right[1], c.right[2], c.front[0], c.front[1], c.front[2],
		c.left[0], c.left[1], c.left[2] = c.left[0], c.left[1], c.left[2], c.back[0], c.back[1], c.back[2],
		c.right[0], c.right[1], c.right[2], c.front[0], c.front[1], c.front[2]
}

func rotateRightCW() {
	rotateFaceCW(c.right)
	c.top[2], c.top[5], c.top[8], c.back[0], c.back[3], c.back[6], c.bottom[2], c.bottom[5], c.bottom[8],
		c.front[2], c.front[5], c.front[8] = c.front[2], c.front[5], c.front[8], c.top[8], c.top[5], c.top[2],
		c.back[6], c.back[3], c.back[0], c.bottom[2], c.bottom[5], c.bottom[8]
}

func rotateLeftCW() {
	rotateFaceCW(c.left)
	c.top[0], c.top[3], c.top[6], c.front[0], c.front[3], c.front[6], c.bottom[0], c.bottom[3], c.bottom[6],
		c.back[2], c.back[5], c.back[8] = c.back[8], c.back[5], c.back[2], c.top[0], c.top[3], c.top[6],
		c.front[0], c.front[3], c.front[6], c.bottom[6], c.bottom[3], c.bottom[0]
}

func rotateBottomCW() {
	rotateFaceCW(c.bottom)
	// front -> right -> back -> left -> front ...
	c.front[6], c.front[7], c.front[8], c.right[6], c.right[7], c.right[8], c.back[6], c.back[7], c.back[8],
		c.left[6], c.left[7], c.left[8] = c.left[6], c.left[7], c.left[8], c.front[6], c.front[7], c.front[8],
		c.right[6], c.right[7], c.right[8], c.back[6], c.back[7], c.back[8]
}

func rotateBackCW() {
	rotateFaceCW(c.back)
	// top -> left -> bottom -> right -> top...
	c.top[0], c.top[1], c.top[2], c.left[0], c.left[3], c.left[6], c.bottom[6], c.bottom[7], c.bottom[8],
		c.right[2], c.right[5], c.right[8] = c.right[2], c.right[5], c.right[8], c.top[2], c.top[1], c.top[0],
		c.left[0], c.left[3], c.left[6], c.bottom[8], c.bottom[7], c.bottom[6]
}

func rotateYCW() {
	// right -> front -> left -> back -> right...
	rotateTopCW()
	rotateECW()
	rotateECW()
	rotateECW()
	rotateBottomCW()
	rotateBottomCW()
	rotateBottomCW()

}

func rotateXCW() {
	// top -> back -> bottom -> front -> top...
	rotateLeftCW()
	rotateLeftCW()
	rotateLeftCW()
	rotateRightCW()
	rotateMCW()
}

func rotateZCW() {
	// top -> right -> bottom -> left -> top...
	rotateFrontCW()
	rotateSCW()
	rotateBackCW()
	rotateBackCW()
	rotateBackCW()
}

func rotateMCW() {
	// top -> back -> bottom -> front -> top...
	c.top[1], c.top[4], c.top[7], c.back[1], c.back[4], c.back[7], c.bottom[1], c.bottom[4], c.bottom[7],
		c.front[1], c.front[4], c.front[7] = c.front[1], c.front[4], c.front[7], c.top[7], c.top[4], c.top[1],
		c.back[7], c.back[4], c.back[1], c.bottom[1], c.bottom[4], c.bottom[7]
}

func rotateECW() {
	// this is clockwise relative to the bottom
	// front -> right -> back -> left -> front...
	c.front[3], c.front[4], c.front[5], c.right[3], c.right[4], c.right[5], c.back[3], c.back[4], c.back[5],
		c.left[3], c.left[4], c.left[5] = c.left[3], c.left[4], c.left[5], c.front[3], c.front[4], c.front[5],
		c.right[3], c.right[4], c.right[5], c.back[3], c.back[4], c.back[5]
}

func rotateSCW() {
	// top -> right -> bottom -> left -> top...
	c.top[3], c.top[4], c.top[5], c.right[1], c.right[4], c.right[7], c.bottom[3], c.bottom[4], c.bottom[5],
		c.left[1], c.left[4], c.left[7] = c.left[7], c.left[4], c.left[1], c.top[3], c.top[4], c.top[5],
		c.right[7], c.right[4], c.right[1], c.bottom[3], c.bottom[4], c.bottom[5]
}

func start() {
	if running {
		return
	}
	running = true
	fmt.Println("starting the cube... how exciting")
}

func stop() {
	if !running {
		return
	}
	running = false
	fmt.Println("stopping the cube... how sad")
}

func Draw() {
	if !running {
		return
	}

	for k, v := range c.top {
		//rl.DrawRectangle(cubeSize*drawSize+drawSize*int32(k%cubeSize), drawSize*int32(k/cubeSize), drawSize, drawSize, getColor(v))

		// we know the bottom 3 need to butt up agains the top of the front face
		rl.DrawTriangleStrip([]rl.Vector2{
			{X: float32(cubeSize*drawSize+drawSize*int32(k%cubeSize)) + drawSize/2.0 + float32(int32(cubeSize-1-k/cubeSize)*drawSize/2.0), Y: float32(drawSize*int32(k/cubeSize)) + drawSize/2.0 + float32(int32(cubeSize-1-k/cubeSize)*drawSize/2.0)},          // TL
			{X: float32(cubeSize*drawSize+drawSize*int32(k%cubeSize)) + float32(int32(cubeSize-1-k/cubeSize)*drawSize/2.0), Y: float32(drawSize*int32(k/cubeSize)+drawSize) + float32(int32(cubeSize-1-k/cubeSize)*drawSize/2.0)},                               // BL
			{X: float32(cubeSize*drawSize+drawSize*int32(k%cubeSize)+drawSize) + drawSize/2.0 + float32(int32(cubeSize-1-k/cubeSize)*drawSize/2.0), Y: float32(drawSize*int32(k/cubeSize)) + drawSize/2.0 + float32(int32(cubeSize-1-k/cubeSize)*drawSize/2.0)}, // TR
			{X: float32(cubeSize*drawSize+drawSize*int32(k%cubeSize)+drawSize) + float32(int32(cubeSize-1-k/cubeSize)*drawSize/2.0), Y: float32(drawSize*int32(k/cubeSize)+drawSize) + float32(int32(cubeSize-1-k/cubeSize)*drawSize/2.0)},                      // BR
		}, getColor(v))
	}
	// for k, v := range c.bottom {
	// 	rl.DrawRectangle(cubeSize*drawSize+drawSize*int32(k%cubeSize), 2*cubeSize*drawSize+drawSize*int32(k/cubeSize), drawSize, drawSize, getColor(v))
	// }
	for k, v := range c.front {
		rl.DrawRectangle(cubeSize*drawSize+drawSize*int32(k%cubeSize), cubeSize*drawSize+drawSize*int32(k/cubeSize), drawSize, drawSize, getColor(v))
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
			{X: float32(2*cubeSize*drawSize + drawSize*int32(k%cubeSize)/2.0), Y: float32(cubeSize*drawSize+drawSize*int32(k/cubeSize)) - float32(int32(k%cubeSize)*drawSize/2.0)},                               // TL
			{X: float32(2*cubeSize*drawSize + drawSize*int32(k%cubeSize)/2.0), Y: float32(cubeSize*drawSize+drawSize*int32(k/cubeSize)+drawSize) - float32(int32(k%cubeSize)*drawSize/2.0)},                      // BL
			{X: float32(2*cubeSize*drawSize + drawSize*int32(k%cubeSize)/2.0 + drawSize/2.0), Y: float32(cubeSize*drawSize+drawSize*int32(k/cubeSize)) - drawSize/2.0 - float32(int32(k%cubeSize)*drawSize/2.0)}, // TR
			{X: float32(2*cubeSize*drawSize + drawSize*int32(k%cubeSize)/2.0 + drawSize/2.0), Y: float32(cubeSize*drawSize+drawSize*int32(k/cubeSize)) + drawSize/2.0 - float32(int32(k%cubeSize)*drawSize/2.0)}, // BR
		}, getColor(v))
	}
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
		fmt.Println("canceling")
		randoCancelFunc()
		randoCancelFunc = nil
		return
	}
	var c context.Context
	c, randoCancelFunc = context.WithCancel(context.Background())
	go func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				r := rand.Intn(10)
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
					rotateXCW()
				case 6:
					rotateYCW()
				case 7:
					rotateZCW()
				case 8:
					rotateBackCW()
				case 9:
					rotateBottomCW()
				}
				time.Sleep(25 * time.Millisecond)
			}
		}
	}(c)
	go func() {
		time.Sleep(5 * time.Second)
		randoCancelFunc()
		randoCancelFunc = nil
	}()
}
