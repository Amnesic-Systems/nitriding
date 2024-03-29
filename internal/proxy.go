package internal

import (
	"fmt"
	"sync"
	"time"

	proxy "github.com/Amnesic-Systems/nitriding-proxy"
	"github.com/mdlayher/vsock"
)

// runNetworking sets up our networking environment.  If anything fails, we try
// again after a brief wait period.
func runNetworking(c *Config, stop chan struct{}) {
	var err error
	for {
		if err = setupNetworking(c, stop); err == nil {
			return
		}
		elog.Printf("Error setting up networking: %v", err)
		time.Sleep(time.Second)
	}
}

// setupNetworking sets up the enclave's networking environment.  In
// particular, this function:
//
//  1. Creates a tun device.
//  2. Set up networking links.
//  3. Establish a connection with the proxy running on the host.
//  4. Spawn goroutines to forward traffic between the tun device and the proxy
//     running on the host.
func setupNetworking(c *Config, stop chan struct{}) error {
	// proxyCID determines the CID (analogous to an IP address) of the parent
	// EC2 instance.  According to the AWS docs, it is always 3:
	// https://docs.aws.amazon.com/enclaves/latest/user/nitro-enclave-concepts.html
	const proxyCID = 3
	var wg sync.WaitGroup

	// Establish TCP-over-VSOCK connection with nitriding-proxy.
	conn, err := vsock.Dial(proxyCID, proxy.DefaultPort, nil)
	if err != nil {
		return fmt.Errorf("failed to connect to nitriding-proxy: %w", err)
	}
	defer conn.Close()
	elog.Println("Established TCP connection with nitriding-proxy.")

	// Create and configure the tun device.
	tun, err := proxy.SetupTunAsEnclave()
	if err != nil {
		return fmt.Errorf("failed to set up tun device: %w", err)
	}
	defer tun.Close()
	elog.Println("Set up tun device.")

	// Configure our DNS resolver.
	if err = writeResolvconf(); err != nil {
		return fmt.Errorf("failed to create resolv.conf: %w", err)
	}
	elog.Println("Configured DNS resolver.")

	// Spawn goroutines that forward traffic.
	errCh := make(chan error, 1)
	wg.Add(2)
	defer wg.Wait()
	go proxy.VsockToTun(conn, tun, errCh, &wg)
	go proxy.TunToVsock(tun, conn, errCh, &wg)
	elog.Println("Started goroutines to forward traffic.")

	select {
	case err := <-errCh:
		return err
	case <-stop:
		elog.Printf("Shutting down networking.")
		return nil
	}
}
