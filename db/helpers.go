package db

import (
	"fmt"
	"time"
)

func GenerateTicketUUID(orderID int, turn int) string {
	return fmt.Sprintf("%dx%dx%d", orderID, time.Now().UnixNano(), turn)
}
