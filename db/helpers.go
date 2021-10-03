package db

import (
	"fmt"
	"time"
)

func GenerateTicketUUID() string {
	return fmt.Sprintf("PO%d", time.Now().Unix())
}
