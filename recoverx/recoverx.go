package recoverx

import (
	"context"
	"fmt"
	"log"
	"runtime/debug"
)

func RecoverAndLog(label interface{}) {
	if r := recover(); r != nil {
		var labelStr string
		switch v := label.(type) {
		case string:
			labelStr = v
		case context.Context:
			labelStr = "Context: " + fmt.Sprintf("%v", v)
		default:
			labelStr = fmt.Sprintf("%T", v)
		}

		log.Printf("🔥 Panic recovered in %s: %v\n%s", labelStr, r, debug.Stack())
	}
}
