package bluetooth

import (
	"context"
	"time"

	"github.com/muka/go-bluetooth/api"
	"github.com/muka/go-bluetooth/bluez/profile/device"

	"local.com/internal/flow"
)

// AdapterScanner implements flow.BluetoothScanner using go-bluetooth.
type AdapterScanner struct{}

func NewAdapterScanner() *AdapterScanner { return &AdapterScanner{} }

func (s *AdapterScanner) Scan(ctx context.Context, timeout time.Duration) ([]flow.BTDevice, error) {
	a, err := api.GetDefaultAdapter()
	if err != nil {
		return nil, err
	}

	ch, cancel, err := api.Discover(a, nil)
	if err != nil {
		return nil, err
	}
	defer cancel()

	devices := []flow.BTDevice{}
	timer := time.NewTimer(timeout)
	defer timer.Stop()

	for {
		select {
		case <-ctx.Done():
			return devices, ctx.Err()
		case <-timer.C:
			return devices, nil
		case ev, ok := <-ch:
			if !ok {
				return devices, nil
			}
			// when a device is discovered, load its properties
			dev, err := device.NewDevice1(ev.Path)
			if err != nil {
				continue
			}
			props := dev.Properties
			name := props.Alias
			if name == "" {
				name = props.Name
			}
			devices = append(devices, flow.BTDevice{Address: props.Address, Name: name})
		}
	}
}
