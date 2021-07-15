package cube

import "testing"

func BenchmarkRotateFrontCW(b *testing.B) {
	for i := 0; i < b.N; i++ {
		rotateFrontCW()
	}
}

func BenchmarkHandleCommand(b *testing.B) {
	running = true
	for i := 0; i < b.N; i++ {
		HandleCommand([]string{"move", "Y'"})
	}
}
