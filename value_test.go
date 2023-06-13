package metrics

import (
	"fmt"
	"runtime"
	"testing"

	"github.com/uber-go/tally"
	"go.uber.org/net/metrics/tallypush"
)

func BenchmarkValueVector(b *testing.B) {
	b.Run("getOrCreate", func(b *testing.B) {
		vect := newCounterVector(metadata{varTagNames: []string{"key"}})
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			_, err := vect.getOrCreate([]string{"key", fmt.Sprint("val", i)})
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("get", func(b *testing.B) {
		vect := newCounterVector(metadata{varTagNames: []string{"key"}})
		_, err := vect.getOrCreate([]string{"key", "val0"})
		if err != nil {
			b.Fatal(err)
		}
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			runtime.KeepAlive(vect.MustGet("key", "val0"))
		}
	})

	const _loopLimit = 10_000
	b.Run(fmt.Sprint("loop", _loopLimit), func(b *testing.B) {
		name := ""
		vect := newCounterVector(metadata{Name: &name, varTagNames: []string{"key"}})
		for i := 0; i < _loopLimit; i++ {
			_, err := vect.getOrCreate([]string{"key", fmt.Sprint("val", i)})
			if err != nil {
				b.Fatal(err)
			}
		}
		pusher := tallypush.New(tally.NoopScope)
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			vect.push(pusher)
		}
	})
}
