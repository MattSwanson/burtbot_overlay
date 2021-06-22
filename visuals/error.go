package visuals

import (
	"fmt"
	"sync"
	"time"

	"github.com/MattSwanson/burtbot_overlay/sound"
	rl "github.com/MattSwanson/raylib-go/raylib"
)

const errorLifetime = 5

var img rl.Texture2D

type ErrorBox struct {
	scale     float32
	x         float32
	y         float32
	spawnTime time.Time
}
type ErrorManager struct {
	es      []*ErrorBox
	esLock  sync.Mutex
	Visible bool
}

func (em *ErrorManager) Draw() {
	if !em.Visible {
		return
	}
	for _, eb := range em.es {
		if eb == nil {
			fmt.Println("this one is already gone...")
			continue
		}
		rl.DrawTextureEx(img, rl.Vector2{X: eb.x, Y: eb.y}, 0.0, eb.scale, rl.White)
	}
}

func NewErrorManager() *ErrorManager {
	img = rl.LoadTexture("./images/hmm.png")
	es := []*ErrorBox{}
	return &ErrorManager{
		es:      es,
		esLock:  sync.Mutex{},
		Visible: true,
	}
}

func (em *ErrorManager) AddError(n int) {
	go func(n int) {
		for i := 0; i < n; i++ {
			em.esLock.Lock()
			e := &ErrorBox{
				x:         float32(len(em.es)) * 50,
				y:         float32(len(em.es)) * 50,
				scale:     3.0,
				spawnTime: time.Now(),
			}
			em.es = append(em.es, e)
			em.esLock.Unlock()
			sound.Play("sosumi")
			time.Sleep(time.Millisecond * 500)
		}
	}(n)
}

func (em *ErrorManager) Clear() {
	em.esLock.Lock()
	em.es = []*ErrorBox{}
	em.esLock.Unlock()
}

func (em *ErrorManager) Update(delta float64) {
	// for k, e := range em.es {
	// 	if e == nil {
	// 		fmt.Println("trying to update item that's already gone...")
	// 		continue
	// 	}
	// 	if time.Since(e.spawnTime).Seconds() >= errorLifetime {
	// 		// reslice
	// 		// e.active = false
	// 		em.esLock.Lock()
	// 		em.es = removeError(em.es, k)
	// 		em.esLock.Unlock()
	// 	}
	// }
}

func removeError(es []*ErrorBox, id int) []*ErrorBox {
	es[id], es[len(es)-1] = es[len(es)-1], es[id]
	return es[:len(es)-1]
}
