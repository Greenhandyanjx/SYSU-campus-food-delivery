package utils

import (
	"fmt"
	"math/rand"
	"time"
)

func GeneratePickupCode() string {
	rand.Seed(time.Now().UnixNano())
	letters := "ABCDEFGHJKMNPQRSTUVWXYZ"
	return fmt.Sprintf("%c%03d",
		letters[rand.Intn(len(letters))],
		rand.Intn(1000),
	)
}
