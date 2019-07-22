package api

import (
	"encoding/json"
	"strings"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/labstack/echo"
	"github.com/nats-io/nats.go"
	"github.com/nsyszr/lcm/pkg/api/resource"
	log "github.com/sirupsen/logrus"
)

func (h *Handler) realtimeEventsHandler() echo.HandlerFunc {
	return func(c echo.Context) error {
		conn, _, _, err := ws.UpgradeHTTP(c.Request(), c.Response())
		if err != nil {
			log.Error("api: failed to upgrade to websocket: ", err)
			return nil
		}
		// go func() {
		defer conn.Close()

		for {

			if _, err := h.nc.Subscribe("iotcore.devicecontrol.v1.*.events.*", func(msg *nats.Msg) {

				// Get namespace and topic from NATS subject
				strippedSubject := strings.TrimPrefix(msg.Subject, "iotcore.devicecontrol.v1.")
				s := strings.Split(strippedSubject, ".")
				namespace := s[0]
				topic := s[2]

				// Parse the message and send it
				var data interface{}
				if err := json.Unmarshal(msg.Data, &data); err == nil {
					event := resource.NewRealtimeEvent(namespace, topic, data)
					out, _ := json.Marshal(event)
					err = wsutil.WriteServerMessage(conn, ws.OpText, out)
					if err != nil {
						log.Error("api: failed to send realtime event: ", err)
					}
				}

			}); err != nil {
				log.Fatal(err)
			}

			select {}

		}
		// }()

		return nil
	}
}
