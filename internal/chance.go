package internal

import (
	"fmt"
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().Unix())
}

// there's a 5 in 12 chance i'm going to graduate
// there's a Odds(5, 12) chance i'm going to graduate
func Odds(x int, in int) bool {
	randomNumber := rand.Intn(in)
	fmt.Println(randomNumber)
	return x > randomNumber
}
