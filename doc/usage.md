## Usage

The following steps are necessary to run your application inside an enclave,
using nitriding.

1. Make sure that your enclave application supports [reproducible
   builds](https://reproducible-builds.org); otherwise, users won't be able to
   verify your enclave image.  Both Rust and Go support reproducible builds
   although some effort may be necessary to get there.
   [Nitriding's Makefile](../Makefile) shows how one can build a Go program
   reproducibly.

2. Set up
   [nitriding-proxy](https://github.com/Amnesic-Systems/nitriding-proxy)
   on the EC2 host.  Run it as follows:
   ```
   sudo ./nitriding-proxy
   ```

3. Build the nitriding executable by running `make nitriding`.
   (Then, run `./nitriding -help` to see a list of command line options.)
   For reproducible Docker images, we recommend
   [kaniko](https://github.com/GoogleContainerTools/kaniko)
   or
   [ko](https://github.com/ko-build/ko) (for Go applications only).
   Take a look at [this
   Makefile](https://github.com/Amnesic-Systems/example-enclave-applications/blob/baceb32edb053581a4619be94c79028409ee3c20/iperf3-enclave/Makefile#L14-L21)
   to see an application of kaniko.

3. Bundle nitriding and your enclave application together
   in a Dockerfile.  The nitriding stand-alone executable must be invoked
   first, followed by your application.  There are two ways to go about this.
   First, you can create a shell script that first starts nitriding in the
   background, followed by the enclave application.  [Here's an
   example](https://github.com/Amnesic-Systems/example-enclave-applications/blob/baceb32edb053581a4619be94c79028409ee3c20/iperf3-enclave/start.sh).  Second, you can tell nitriding to start your
   enclave application for you:
   ```
   nitriding -app-cmd "my-enclave-app -s foo"
   ```
   This instructs nitriding to invoke the command `my-enclave-app -s foo`.
   Nitriding keeps running as long as my-enclave-app is running.

4. There's one more thing, but only if you invoked nitriding with the flag
   `-wait-for-app`: Once your application is done bootstrapping, it must let
   nitriding know, so it can start the Internet-facing Web server that handles
   remote attestation and other tasks.  To do so, the application must issue an
   HTTP GET request to `http://127.0.0.1:8080/enclave/ready`.  The handler
   ignores URL parameters and responds with a status code 200 if the request
   succeeded.  Note that the port in this example, 8080, is controlled by
   nitriding's `-int-port` command line flag.  Ignore this paragraph if you did
   not use `-wait-for-app`.

Finally, take a look at
[this simple application](/example)
for an example on how one can use nitriding.
