package drop

import (
    "time"
)

const (
    dropHeight float64 = -5000
    dropVelocity float64 = 25
)

type Core struct {
    lastUpdate time.Time
}

func (c *Core) Update() error {
    // Update state here
}

func (c *Core) Draw() {
    // Draw stuff here
}

func (c *Core) HandleMessage(args []string) {
    // Handle messages sent from the bot to this game
}
