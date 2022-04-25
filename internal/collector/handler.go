package collector

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/openconfig/gnmi/proto/gnmi"
)

const (
	operationUpdate = "update"
	operationDelete = "delete"
	//
	dotReplChar   = "^"
	spaceReplChar = "~"
)

var regDot = regexp.MustCompile(`\.`)
var regSpace = regexp.MustCompile(`\s`)

func (c *targetCollector) handleSubscribeResponse(resp *gnmi.SubscribeResponse) error {
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
func (c *targetCollector) notificationToStateMsg(targetName string, n *gnmi.Notification) []*stateMsg {
	sb := new(strings.Builder)
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
			sm := &stateMsg{
				Subject:   sb.String(),
				Timestamp: n.GetTimestamp(),
				Operation: operationUpdate,
				Data:      []byte(upd.Val.GetStringVal()),
			}
			c.log.Debug("state message", "notification", n, "msg", sm)
			result = append(result, sm)
		}
	}
	for _, del := range n.GetDelete() {
		if pr := gNMIPathToSubject(del); pr != "" {
			sb.Reset()
			fmt.Fprintf(sb, "%s.%s", prefix, pr)
			sm := &stateMsg{
				Subject:   sb.String(),
				Timestamp: n.GetTimestamp(),
				Operation: operationDelete,
			}
			c.log.Debug("state message", "notification", n, "msg", sm)
			result = append(result, sm)
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
			for _, k := range kNames {
				sk := sanitizeKey(e.GetKey()[k])
				fmt.Fprintf(sb, ".{%s=%s}", k, sk)
			}
		}
	}
	return sb.String()
}

func sanitizeKey(k string) string {
	s := regDot.ReplaceAllString(k, dotReplChar)
	return regSpace.ReplaceAllString(s, spaceReplChar)
}

// TODO:
func subjectTogNMIPath(s string) string {
	return ""
}
