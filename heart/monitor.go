package heart

import (
	"errors"
	"time"

	log "github.com/coreos/fleet/Godeps/_workspace/src/github.com/golang/glog"
)

func NewMonitor(ttl time.Duration) *Monitor {
	return &Monitor{ttl, ttl / 2}
}

type Monitor struct {
	TTL  time.Duration
	ival time.Duration
}

// Monitor ensures a Heart is still beating until a channel is closed, returning
// an error if the heartbeats fail.
func (m *Monitor) Monitor(hrt Heart, stop chan bool) error {
	ticker := time.Tick(m.ival)
	for {
		select {
		case <-stop:
			log.V(1).Info("Monitor exiting due to stop signal")
			return nil
		case <-ticker:
			if _, err := m.check(hrt); err != nil {
				return err
			}
		}
	}
}

// check attempts to beat a Heart several times within a timeout, returning the
// log index at which the beat succeeded or an error
func (m *Monitor) check(hrt Heart) (idx uint64, err error) {
	// time out after a third of the machine presence TTL, attempting
	// the heartbeat up to four times
	timeout := m.TTL / 3
	interval := timeout / 4

	tchan := time.After(timeout)
	next := time.After(0)
	for idx == 0 {
		select {
		case <-tchan:
			err = errors.New("Monitor timed out before successful heartbeat")
			return
		case <-next:
			idx, err = hrt.Beat(m.TTL)
			if err != nil {
				log.V(1).Infof("Monitor heartbeat function returned err, retrying in %v: %v", interval, err)
			}

			next = time.After(interval)
		}
	}

	return
}
