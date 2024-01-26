# Usage

This document explains how you can use nitriding to run your application inside an AWS Nitro Enclave.
Feeling impatient?
Take a look at [this enclave application](/example) for a simple example of how one can use nitriding.

## Optional: Ensure reproducible builds

Make sure that both your enclave application *and* Dockerfile
[build reproducibly](https://reproducible-builds.org);
otherwise,
users won't be able to verify your enclave image.

> [!NOTE]
> Reproducible builds are only necessary if you have users who will verify your enclave application.
> If you have no users, you don't need reproducible builds.

Docker itself cannot build images reproducibly because timestamps
and other artifacts result in slightly different images on every compilation.
To build reproducible Docker images, we recommend
[kaniko](https://github.com/GoogleContainerTools/kaniko)
or
[ko](https://github.com/ko-build/ko) (for Go applications only).

Both Rust and Go support reproducible builds although some effort may be necessary to get there.
[Nitriding's Makefile](../Makefile)
shows how one can build a Go program reproducibly.

> [!TIP]
> For Alice and Bob to compile identical Docker images from a given Dockerfile,
> they need to use a reproducible build system like kaniko or ko in the same version,
> use compilers in the same version,
> and compile the source code reproducibly.

## Set up nitriding-proxy

Your EC2 host must be running
[nitriding-proxy](https://github.com/Amnesic-Systems/nitriding-proxy),
which forwards network traffic to and from nitriding.
Without nitriding-proxy, nitriding is unable to talk to the outside world.
Clone the repository and start the service by running:

```
make run
```

If your enclave application

## Create your Dockerfile

An enclave image file is little more than a Docker image.
This Docker image must contain your application and nitriding.
You can use the following Dockerfile excerpt to incorporate nitriding:

```dockerfile
FROM golang:1.21 as builder
# Unset CGO_ENABLED to build a statically-linked binary. This is necessary because
# the golang image uses libc while alpine uses libmusl.
RUN CGO_ENABLED=0 go install github.com/Amnesic-Systems/nitriding@latest
RUN cp $GOPATH/bin/nitriding /
# Use whatever base image you like.
FROM alpine
COPY --from=builder /nitriding /bin/

# The rest of your Dockerfile goes here.
```

The nitriding executable must be invoked first,
followed by your application.
There are two ways to go about this in your Dockerfile:

1. Instead of invoking nitriding or your application directly,
   you can invoke a shell script that first starts nitriding in the
   background, followed by the enclave application.
   [Here's an example](https://github.com/Amnesic-Systems/example-enclave-applications/blob/baceb32edb053581a4619be94c79028409ee3c20/iperf3-enclave/start.sh).

3. You can tell nitriding to start your enclave application for you:
   
   ```
   nitriding -app-cmd "my-enclave-app -s foo"
   ```

   This instructs nitriding to invoke the command `my-enclave-app -s foo`.
   Nitriding keeps running as long as my-enclave-app is running.

## Optional: Signal readiness to nitriding

If you invoked nitriding with the command line flag `-wait-for-app`,
nitriding refrains from setting up its Internet-facing networking
until your application signals its readiness.
To do so,
your application must issue an HTTP GET request to `http://127.0.0.1:8080/enclave/ready`.
This endpoint ignores URL parameters and responds with a status code 200 if the request succeeded.
Note that the port in this example,
8080,
is controlled by nitriding's `-int-port` command line flag.
There's usually no need to change this port but you can,
if you need to.

> [!NOTE]
> This step is only necessary if you invoked nitriding with the flag `-wait-for-app`.
