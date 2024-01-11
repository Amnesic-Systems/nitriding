package internal

import (
	"fmt"
	"net"
	"os"
	"unsafe"

	"golang.org/x/sys/unix"

	"github.com/hf/nsm"
	"github.com/hf/nsm/request"
	"github.com/milosgajdos/tenus"
)

const (
	entropySeedDevice = "/dev/random"
	entropySeedSize   = 2048
	addrLo            = "127.0.0.1/8"
	ifaceLo           = "lo"
)

// configureLoIface assigns an IP address to the loopback interface.
func configureLoIface() error {
	l, err := tenus.NewLinkFrom(ifaceLo)
	if err != nil {
		return err
	}
	addr, network, err := net.ParseCIDR(addrLo)
	if err != nil {
		return err
	}
	if err = l.SetLinkIp(addr, network); err != nil {
		return err
	}
	return l.SetLinkUp()
}

// writeResolvconf creates our resolv.conf and adds a nameserver.
func writeResolvconf() error {
	// A Nitro Enclave's /etc/resolv.conf is a symlink to
	// /run/resolvconf/resolv.conf.  As of 2022-11-21, the /run/ directory
	// exists but not its resolvconf/ subdirectory.
	dir := "/run/resolvconf/"
	file := dir + "resolv.conf"

	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directories: %w", err)
	}

	// Our default gateway -- gvproxy -- also operates a DNS resolver.
	c := "nameserver 1.1.1.1\n"
	if err := os.WriteFile(file, []byte(c), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// maybeSeedEntropy obtains cryptographically secure random bytes from the
// Nitro Secure Module (NSM) and uses them to initialize the system's random
// number generator.  If we don't do that, our system is going to start with no
// entropy, which means that calls to /dev/(u)random will block.
func maybeSeedEntropy() {
	// Abort if we're not in an enclave.
	if !inEnclave {
		elog.Println("We are not inside an enclave.  Not seeding entropy pool.")
		return
	}

	s, err := nsm.OpenDefaultSession()
	if err != nil {
		elog.Fatal(err)
	}
	defer func() {
		_ = s.Close()
	}()

	fd, err := os.OpenFile(entropySeedDevice, os.O_WRONLY, os.ModePerm)
	if err != nil {
		elog.Fatal(err)
	}
	defer func() {
		if err = fd.Close(); err != nil {
			elog.Printf("Failed to close %q: %s", entropySeedDevice, err)
		}
	}()

	var written int
	for totalWritten := 0; totalWritten < entropySeedSize; {
		res, err := s.Send(&request.GetRandom{})
		if err != nil {
			elog.Fatalf("Failed to communicate with hypervisor: %s", err)
		}
		if res.GetRandom == nil {
			elog.Fatal("no GetRandom part in NSM's response")
		}
		if len(res.GetRandom.Random) == 0 {
			elog.Fatal("got no random bytes from NSM")
		}

		// Write NSM-provided random bytes to the system's entropy pool to seed
		// it.
		if written, err = fd.Write(res.GetRandom.Random); err != nil {
			elog.Fatal(err)
		}
		totalWritten += written

		// Tell the system to update its entropy count.
		if _, _, errno := unix.Syscall(
			unix.SYS_IOCTL,
			uintptr(fd.Fd()),
			uintptr(unix.RNDADDTOENTCNT),
			uintptr(unsafe.Pointer(&written)),
		); errno != 0 {
			elog.Printf("Failed to update system's entropy count: %s", errno)
		}
	}

	elog.Println("Initialized the system's entropy pool.")
}
