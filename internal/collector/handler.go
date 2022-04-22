package collector

import (
	"fmt"
	"sort"
	"strings"

	"github.com/openconfig/gnmi/proto/gnmi"
)

const (
	operationUpdate = "update"
	operationDelete = "delete"
)

func (c *stateCollector) handleSubscription(resp *gnmi.SubscribeResponse) error {
	targetName := c.GetTarget().Config.Name

	log := c.log.WithValues("Target", targetName)
	//log.Debug("handle target update from device")

	switch resp.GetResponse().(type) {
	case *gnmi.SubscribeResponse_Update:
		log.Debug("handle target update from device", "Prefix", resp.GetUpdate().GetPrefix())

		for _, msg := range c.notificationToStateMsg(targetName, resp.GetUpdate()) {
			c.updateCh <- msg
		}

	case *gnmi.SubscribeResponse_SyncResponse:
		log.Debug("SyncResponse")
	}

	return nil
}

// TODO: sanitize subject before returning
func (c *stateCollector) notificationToStateMsg(targetName string, n *gnmi.Notification) []*stateMsg {
	sb := new(strings.Builder)
	// fmt.Fprintf(sb, "%s.%s.%s", streamName, c.namespace, targetName)
	fmt.Fprintf(sb, "%s.%s", streamName, targetName)
	if pr := gNMIPathToSubject(n.GetPrefix()); pr != "" {
		fmt.Fprintf(sb, ".%s", pr)
	}
	prefix := sb.String()
	result := make([]*stateMsg, 0, len(n.GetUpdate())+len(n.GetDelete()))
	for _, upd := range n.GetUpdate() {
		if pr := gNMIPathToSubject(upd.GetPath()); pr != "" {
			sb.Reset()
			fmt.Fprintf(sb, "%s.%s", prefix, pr)
			result = append(result, &stateMsg{
				Subject:   sb.String(),
				Timestamp: n.GetTimestamp(),
				Operation: operationUpdate,
				Data:      []byte(upd.Val.GetStringVal()),
			})
		}
	}
	for _, del := range n.GetDelete() {
		if pr := gNMIPathToSubject(del); pr != "" {
			sb.Reset()
			fmt.Fprintf(sb, "%s.%s", prefix, pr)
			result = append(result, &stateMsg{
				Subject:   sb.String(),
				Timestamp: n.GetTimestamp(),
				Operation: operationDelete,
			})
		}
	}
	return result
}

func gNMIPathToSubject(p *gnmi.Path) string {
	if p == nil {
		return ""
	}
	sb := new(strings.Builder)
	if p.GetOrigin() != "" {
		fmt.Fprintf(sb, "%s.", p.GetOrigin())
	}
	for i, e := range p.GetElem() {
		if i > 0 {
			sb.WriteString(".")
		}
		sb.WriteString(e.Name)
		if len(e.Key) > 0 {
			// sort keys by name
			kNames := make([]string, 0, len(e.Key))
			for k := range e.Key {
				kNames = append(kNames, k)
			}
			sort.Strings(kNames)
			fmt.Fprintf(sb, ".{")
			for _, k := range kNames {
				fmt.Fprintf(sb, "%s=%s", k, e.GetKey()[k])
			}
			fmt.Fprintf(sb, "}")
		}
	}
	return sb.String()
}

func getValue(v *gnmi.TypedValue) (string, []byte) {
	// TODO: for handling json/json_ietf encodings
	return "", nil
}
