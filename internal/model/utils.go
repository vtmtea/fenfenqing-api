package model

import (
	"fmt"
	"math/rand"
	"time"
)

// GenerateRoomID 生成 6 位房间号
func GenerateRoomID() string {
	rand.Seed(time.Now().UnixNano())
	return fmt.Sprintf("%06d", rand.Intn(1000000))
}
