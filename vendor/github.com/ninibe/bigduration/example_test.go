package bigduration_test

import (
	"fmt"

	"github.com/ninibe/bigduration"
)

func ExampleParseBigDuration() {
	bd, _ := bigduration.ParseBigDuration("1year2month1day5h10s")
	duration := bd.Duration() // convert to time.Duration
	fmt.Println(duration)
	// Output: 10229h0m10s
}
