<div align="center">
  <img src="./doc/nitriding-logo.svg" alt="Nitriding logo" width="250">
</div>

---

[![GoDoc](https://pkg.go.dev/badge/github.com/Amnesic-Systems/nitriding?utm_source=godoc)](https://pkg.go.dev/github.com/Amnesic-Systems/nitriding)

Nitriding is a Go tool kit (consisting of two services) that helps you run your application inside an
[AWS Nitro Enclave](https://aws.amazon.com/ec2/nitro/nitro-enclaves/).
Let's assume that you built a Web service in Rust.  You can now use nitriding to
move your Rust code into a Nitro Enclave, which provides two key security properties:

1. At runtime, Nitro Enclaves are effectively a sealed black box. Nobody can observe your application's state at runtime: not you, not Amnesic Systems, and not even AWS. This makes it possible to process sensitive data _without ever seeing the data_.
2. Optionally, using remote attestation, your users can verify (over the Internet) that you run the code you claim to run. This requires that your application is open source.

The diagram below illustrates how nitriding works.
Gray components are provided by AWS,
blue components are provided by nitriding,
the yellow component is provided by you,
and the brown component is your user – if you have users.
Nitriding helps you run your application (which is bundled as a Docker image)
inside a Nitro Enclave while abstracting away the pitfalls of working with enclaves.
In particular:

* Nitriding provides a [tap](https://docs.kernel.org/networking/tuntap.html) interface inside the enclave, enabling seamless networking for your application. Your application can listen for incoming connections and establish outgoing connections without having to worry about tunneling network traffic over the enclave's VSOCK interface.

* Nitriding's TCP proxy does not see your network traffic; it blindly forwards end-to-end encrypted packets. If your application speaks HTTPS, nitriding can act as a TLS-terminating HTTP reverse proxy. If your application speaks another protocol, you are responsible for the encryption layer.

* Nitriding exposes an HTTPS endpoint for remote attestation, allowing your users to verify over the Internet that you run the code you claim to run. You don't have to worry about the nuances of remote attestation.

* While nitriding is built in Go, it is application-agnostic: As long as you can bundle your application in a Docker image, you can run it using nitriding. You are free to use your favorite tech stack.

<div align="center">
  <img src="https://github.com/Amnesic-Systems/nitriding/assets/1316283/6309c401-4494-4add-a48b-f9403c6fb2c2.png" alt="Nitriding architecture" width="600">
</div>

## More documentation

* [How to use nitriding](doc/usage.md)
* [System architecture](doc/architecture.md)
* [HTTP API](doc/http-api.md)
* [Horizontal scaling](doc/key-synchronization.md)
* [Example application](example/)
* [Setup enclave EC2 host](doc/setup.md)

To learn more about nitriding's trust assumptions, architecture, and build
system, take a look at our [research paper](https://arxiv.org/abs/2206.04123).
