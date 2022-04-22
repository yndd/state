package collector

import (
	"github.com/nats-io/nats.go"
	"github.com/openconfig/gnmi/proto/gnmi"
)

func (c *stateCollector) handleSubscription(resp *gnmi.SubscribeResponse) error {
	targetName := c.GetTarget().Config.Name

	log := c.log.WithValues("Target", targetName)
	//log.Debug("handle target update from device")

	nc, err := nats.Connect(natsServer)
	if err != nil {
		return err
	}
	defer nc.Close()

	js, err := nc.JetStream()
	if err != nil {
		return err
	}

	switch resp.GetResponse().(type) {
	case *gnmi.SubscribeResponse_Update:
		//log.Debug("handle target update from device", "Prefix", resp.GetUpdate().GetPrefix())
		for _, u := range resp.GetUpdate().GetUpdate() {
			publishGnmiUpdate(js, targetName, u)
		}

	case *gnmi.SubscribeResponse_SyncResponse:
		log.Debug("SyncResponse")
	}

	return nil
}
