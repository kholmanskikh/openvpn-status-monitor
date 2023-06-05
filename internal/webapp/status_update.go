package webapp

import openvpn_status "github.com/kholmanskikh/openvpn-status-monitor/internal/openvpn-status"

type StatusUpdate struct {
	Status *openvpn_status.Status
	Error  error
}
