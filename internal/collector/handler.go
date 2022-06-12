package collector

import (
	"fmt"
	"strings"

	"github.com/openconfig/gnmi/proto/gnmi"
	"github.com/yndd/pubsub"
	statesubject "github.com/yndd/state/pkg/subject"
)

func (c *targetCollector) handleSubscribeResponse(resp *gnmi.SubscribeResponse) error {
	targetName := c.GetTarget().Config.Name

	log := c.log.WithValues("Target", targetName)
	//log.Debug("handle target update from device")

	switch resp.GetResponse().(type) {
	case *gnmi.SubscribeResponse_Update:
		log.Debug("handle target update from device", "Prefix", resp.GetUpdate().GetPrefix())

		for _, msg := range c.notificationToPubSubMsg(targetName, resp.GetUpdate()) {
			c.updateCh <- msg
		}

	case *gnmi.SubscribeResponse_SyncResponse:
		log.Debug("SyncResponse")
	}

	return nil
}

// TODO: sanitize subject before returning
func (c *targetCollector) notificationToPubSubMsg(targetName string, n *gnmi.Notification) []*pubsub.Msg {
	sb := new(strings.Builder)
	fmt.Fprintf(sb, "%s.%s", streamName, targetName)
	if pr := statesubject.GNMIPathToSubject(n.GetPrefix()); pr != "" {
		fmt.Fprintf(sb, ".%s", pr)
	}
	prefix := sb.String()
	result := make([]*pubsub.Msg, 0, len(n.GetUpdate())+len(n.GetDelete()))
	for _, upd := range n.GetUpdate() {
		if pr := statesubject.GNMIPathToSubject(upd.GetPath()); pr != "" {
			sb.Reset()
			fmt.Fprintf(sb, "%s.%s", prefix, pr)
			sm := &pubsub.Msg{
				Subject:   sb.String(),
				Timestamp: n.GetTimestamp(),
				Operation: pubsub.Operation_OPERATION_UPDATE,
				Data:      []byte(upd.Val.GetStringVal()),
				Tags: map[string]string{
					"target": targetName,
				},
			}
			c.log.Debug("state message", "notification", n, "msg", sm)
			result = append(result, sm)
		}
	}
	for _, del := range n.GetDelete() {
		if pr := statesubject.GNMIPathToSubject(del); pr != "" {
			sb.Reset()
			fmt.Fprintf(sb, "%s.%s", prefix, pr)
			sm := &pubsub.Msg{
				Subject:   sb.String(),
				Timestamp: n.GetTimestamp(),
				Operation: pubsub.Operation_OPERATION_DELETE,
				Tags: map[string]string{
					"target": targetName,
				},
			}
			c.log.Debug("state message", "notification", n, "msg", sm)
			result = append(result, sm)
		}
	}
	return result
}
