package consumer

import (
	"cdcworker/internal/model"
	"encoding/json"
	"log"
	"time"

	"github.com/lib/pq"
)

func StartListener(connStr string, handler func(model.OrderEvent)) error {
	listener := pq.NewListener(connStr, 10*time.Second, time.Minute, nil)
	err := listener.Listen("order_events")
	if err != nil {
		return err
	}
	log.Println("Listening on order_events")

	for n := range listener.Notify {
		if n == nil {
			continue
		}

		var event model.OrderEvent

		if err := json.Unmarshal(
			[]byte(n.Extra),
			&event,
		); err != nil {

			log.Println(err)
			continue
		}

		handler(event)
	}
	return err
}
