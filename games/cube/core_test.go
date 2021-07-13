package cube

import "testing"

func BenchmarkRotateFrontCW(b *testing.B) {
	for i := 0; i < b.N; i++ {
		rotateFrontCW()
	}
}
