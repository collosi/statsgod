package client

import (
	"testing"
)

func BenchmarkClient(b *testing.B) {
	b.StopTimer()
	c, err := Dial(":16536", 10000000)
	if err != nil {
		b.Fatalf("%v", err)
	}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		c.Update("mykey", 123.4)
	}
	c.Close()
}
