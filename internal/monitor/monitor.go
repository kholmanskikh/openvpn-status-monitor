package monitor

import (
	openvpn_status "github.com/kholmanskikh/openvpn-status-monitor/internal/openvpn-status"
	"github.com/kholmanskikh/openvpn-status-monitor/internal/webapp"
	"log"
	"os"
	"time"
)

var logger *log.Logger

func init() {
	logger = log.New(os.Stderr, "Monitor: ", log.Ldate|log.Ltime|log.Lmsgprefix)
}

func Monitor(filePath string, interval time.Duration, updateChannel chan<- *webapp.StatusUpdate) {
	logger.Printf("monitoring '%s' for updates", filePath)

	for {
		status, err := openvpn_status.ReadFromFile(filePath)
		updateChannel <- &webapp.StatusUpdate{Status: status, Error: err}
		time.Sleep(interval)
	}
}
