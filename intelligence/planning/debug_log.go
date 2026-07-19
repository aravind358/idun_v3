package planning

import (
	"fmt"
	"os"
)

var devLogEnabled = os.Getenv("IDUN_DEV_LOG") != "0"

func devLog(subsystem, message string, detail ...string) {
	if !devLogEnabled {
		return
	}
	if len(detail) > 0 {
		fmt.Printf("[%s]\n%s:\n%s\n\n", subsystem, message, detail[0])
	} else {
		fmt.Printf("[%s]\n%s\n\n", subsystem, message)
	}
}
