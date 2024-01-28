package main

import (
	"bufio"
	"flag"
	"io"
	"log"
	"math"
	"net/url"
	"os"
	"os/exec"
	"strings"

	"github.com/Amnesic-Systems/nitriding/internal"
)

var (
	elog = log.New(os.Stderr, "nitriding: ", log.Ldate|log.Ltime|log.LUTC|log.Lshortfile)
)

func main() {
	var fqdn, fqdnLeader, appURL, appWebSrv, appCmd, prometheusNamespace, mockCertFp string
	var extPubPort, extPrivPort, intPort, hostProxyPort, prometheusPort uint
	var useACME, waitForApp, useProfiling, useVsockForExtPort, disableKeepAlives, debug bool
	var err error

	flag.StringVar(&fqdn, "fqdn", "",
		"FQDN of the enclave application (e.g., \"example.com\").")
	flag.StringVar(&fqdnLeader, "fqdn-leader", "",
		"FQDN of the leader enclave (e.g., \"leader.example.com\").  Setting this enables key synchronization.")
	flag.StringVar(&appURL, "app-url", "",
		"Code repository of the enclave application (e.g., \"github.com/foo/bar\").")
	flag.StringVar(&appWebSrv, "appwebsrv", "",
		"Enclave-internal HTTP server of the enclave application (e.g., \"http://127.0.0.1:8081\").")
	flag.StringVar(&appCmd, "app-cmd", "",
		"Launch enclave application via the given command.")
	flag.StringVar(&prometheusNamespace, "prometheus-namespace", "",
		"Prometheus namespace for exported metrics.")
	flag.UintVar(&extPubPort, "ext-pub-port", 443,
		"Nitriding's external, public HTTPS port.  Must match port forwarding rules on EC2 host.")
	flag.UintVar(&extPrivPort, "ext-priv-port", 444,
		"Nitriding's external, non-public HTTPS port.  Must match port forwarding rules on the EC2 host.")
	flag.BoolVar(&disableKeepAlives, "disable-keep-alives", false,
		"Disables keep-alive connections for the HTTPS service.")
	flag.BoolVar(&useVsockForExtPort, "vsock-ext", false,
		"Listen on VSOCK interface for HTTPS port.")
	flag.UintVar(&intPort, "int-port", 8080,
		"Nitriding's enclave-internal HTTP port.  Only used by the enclave application.")
	flag.UintVar(&hostProxyPort, "host-proxy-port", 1024,
		"Port of proxy application running on EC2 host.")
	flag.UintVar(&prometheusPort, "prometheus-port", 0,
		"Port to expose Prometheus metrics at.")
	flag.BoolVar(&useProfiling, "profile", false,
		"Enable pprof profiling.  Only useful for debugging and must not be used in production.")
	flag.BoolVar(&useACME, "acme", false,
		"Use Let's Encrypt's ACME to fetch HTTPS certificate.")
	flag.BoolVar(&waitForApp, "wait-for-app", false,
		"Start Internet-facing Web server only after application signals its readiness.")
	flag.BoolVar(&debug, "debug", false,
		"Print extra debug messages and use dummy attester for testing outside enclaves.")
	flag.StringVar(&mockCertFp, "mock-cert-fp", "",
		"Mock certificate fingerprint to use in attestation documents (hexadecimal)")
	flag.Parse()

	if fqdn == "" {
		elog.Fatalf("-fqdn must be set.")
	}
	if extPubPort < 1 || extPubPort > math.MaxUint16 {
		elog.Fatalf("-extport must be in interval [1, %d]", math.MaxUint16)
	}
	if extPrivPort < 1 || extPrivPort > math.MaxUint16 {
		elog.Fatalf("-extPrivPort must be in interval [1, %d]", math.MaxUint16)
	}
	if intPort < 1 || intPort > math.MaxUint16 {
		elog.Fatalf("-intport must be in interval [1, %d]", math.MaxUint16)
	}
	if hostProxyPort < 1 || hostProxyPort > math.MaxUint32 {
		elog.Fatalf("-host-proxy-port must be in interval [1, %d]", math.MaxUint32)
	}
	if prometheusPort > math.MaxUint16 {
		elog.Fatalf("-prometheus-port must be in interval [1, %d]", math.MaxUint16)
	}
	if prometheusPort != 0 && prometheusNamespace == "" {
		elog.Fatalf("-prometheus-namespace must be set when Prometheus is used.")
	}

	c := &internal.Config{
		FQDN:                fqdn,
		FQDNLeader:          fqdnLeader,
		ExtPubPort:          uint16(extPubPort),
		ExtPrivPort:         uint16(extPrivPort),
		IntPort:             uint16(intPort),
		UseVsockForExtPort:  useVsockForExtPort,
		DisableKeepAlives:   disableKeepAlives,
		PrometheusPort:      uint16(prometheusPort),
		PrometheusNamespace: prometheusNamespace,
		HostProxyPort:       uint32(hostProxyPort),
		UseACME:             useACME,
		WaitForApp:          waitForApp,
		UseProfiling:        useProfiling,
		MockCertFp:          mockCertFp,
		Debug:               debug,
	}
	if appURL != "" {
		u, err := url.Parse(appURL)
		if err != nil {
			elog.Fatalf("Failed to parse application URL: %v", err)
		}
		c.AppURL = u
	}
	if appWebSrv != "" {
		u, err := url.Parse(appWebSrv)
		if err != nil {
			elog.Fatalf("Failed to parse URL of Web server: %v", err)
		}
		c.AppWebSrv = u
	}
	if debug {
		elog.Println("WARNING: Using debug mode, which must not be enabled in production!")
	}

	enclave, err := internal.NewEnclave(c)
	if err != nil {
		elog.Fatalf("Failed to create enclave: %v", err)
	}

	if err := enclave.Start(); err != nil {
		elog.Fatalf("Enclave terminated: %v", err)
	}

	// Nitriding supports two ways of starting the enclave application:
	//
	// 1) Nitriding spawns the enclave application itself, and waits for it
	//    to terminate.
	//
	// 2) The enclave application is started by a shell script (which also
	//    starts nitriding).  In this case, we simply block forever.
	if appCmd != "" {
		runAppCommand(appCmd, func(s string) { elog.Printf("> %s", s) })
	} else {
		// Block forever.
		<-make(chan struct{})
	}
	elog.Println("Exiting nitriding.")
}

// runAppCommand (i) runs the given command, (ii) waits until the command
// finished execution, and (iii) in the meanwhile prints the command's stdout
// and stderr.
func runAppCommand(appCmd string, f func(string)) {
	var (
		err            error
		stdout, stderr io.ReadCloser
	)
	elog.Printf("Invoking command: %s", appCmd)
	args := strings.Split(appCmd, " ")
	cmd := exec.Command(args[0], args[1:]...)

	if stderr, err = cmd.StderrPipe(); err != nil {
		elog.Fatalf("Error obtaining stderr pipe: %v", err)
	}
	if stdout, err = cmd.StdoutPipe(); err != nil {
		elog.Fatalf("Error obtaining stdout pipe: %v", err)
	}

	if err := cmd.Start(); err != nil {
		elog.Fatalf("Error starting application: %v", err)
	}

	s := bufio.NewScanner(io.MultiReader(stdout, stderr))
	for s.Scan() {
		f(s.Text())
	}
	if err := s.Err(); err != nil {
		elog.Printf("Error reading from application: %v", err)
	}

	if err := cmd.Wait(); err != nil {
		elog.Fatalf("Enclave application exited with non-0 exit code: %v", err)
	}
	elog.Println("Enclave application exited.")
}
