package filters

import (
	"math/rand"
	"testing"
	"time"
	"beta/tools"
)

func TestMiddleware(t *testing.T) {
	go func() {
		for true {
			UpdateProfExtraData(map[string]interface{}{
				"online_count": rand.Int(),
			})
		}
	}()
	event := new(tools.ProfEvent)
	go func() {
		for true {
			GetProfExtraData(event)
		}
	}()

	time.Sleep(10 * time.Minute)
}
