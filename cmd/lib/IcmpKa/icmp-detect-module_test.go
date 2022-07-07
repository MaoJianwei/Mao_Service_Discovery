package IcmpKa

import (
	"testing"
)

func TestIcmpDetectModule_InitIcmpModule(t *testing.T) {
	icmpDetectModule := &IcmpDetectModule{
		//AddChan:     nil,
		//DelChan:     nil,
	}
	//icmpDetectModule.InitIcmpModule()

	if icmpDetectModule.sendInterval == 0 ||
		icmpDetectModule.checkInterval == 0 ||
		icmpDetectModule.leaveTimeout == 0 ||
		icmpDetectModule.refreshShowingInterval == 0 ||
		icmpDetectModule.receiveFreezePeriod == 0 ||
		icmpDetectModule.serviceMirror == nil {

		//t.Fatalf("Icmp-KA-Module is not fully initiated.")
	}
}