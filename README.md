# go-libp2p-vsock

> go-libp2p's VSOCK transport

Package `go-libp2p-vsock` is a libp2p [transport](https://pkg.go.dev/github.com/libp2p/go-libp2p/core/transport). It uses `virtio-vsock` enable communication channel to relays at Nitro Enclaves.

## Install

```sh
go get github.com/balena/go-libp2p-vsock
```

## Testing w/ AWS Nitro Enclaves

Prepare your AWS Nitro Enclave as documented in [AWS Nitro Enclaves User Guide](https://docs.aws.amazon.com/enclaves/latest/user/getting-started.html#launch-instance).

Then, create a `Dockerfile` at the root of this repository with:

```
FROM golang:1.19.5-alpine3.17 as builder
WORKDIR /build
COPY ./ .
RUN cd libp2p-vsock && go build

FROM alpine:3.17
COPY --from=builder /build/libp2p-vsock/libp2p-vsock /
CMD /libp2p-vsock -l /vsock/x/xtcp/5000
```

Then build the enclave:

```
nitro-cli build-enclave --docker-dir ./ --docker-uri libp2p-vsock:latest --output-file libp2p-vsock.eif
```

And run it, in debug mode:

```
nitro-cli run-enclave --eif-path libp2p-vsock.eif --cpu-count 1 --enclave-cid 6 --memory 256 --debug-mode
```

Open the debug console with:

```
nitro-cli console --enclave-name libp2p-vsock
```

At the end of the Kernel messages, you should see the following log:

```
2023/02/02 23:13:40 I am /vsock/6/xtcp/5000/p2p/Qmdpa...9tAN
2023/02/02 23:13:40 listening for connections
2023/02/02 23:13:40 Now run "./libp2p-vsock -l /vsock/x/xtcp/5001 -d /vsock/6/xtcp/5000/p2p/Qmdpa...9tAN" on a different terminal
```

Now execute the indicated command from the host:

```
./libp2p-vsock -l /vsock/x/xtcp/5001 -d /vsock/6/xtcp/5000/p2p/Qmdpa...9tAN
2023/02/02 23:14:51 I am /vsock/3/xtcp/5001/p2p/QmStHj...zH3R
2023/02/02 23:14:51 sender opening stream
2023/02/02 23:14:51 sender saying hello
2023/02/02 23:14:51 read reply: "Hello, world!\n"
```

## Contribute

Feel free to join in. All welcome. Open an [issue](https://github.com/balena/go-libp2p-vsock/issues)!

This repository falls under the libp2p [Code of Conduct](https://github.com/libp2p/community/blob/master/code-of-conduct.md).

### Want to hack on libp2p?

[![](https://cdn.rawgit.com/libp2p/community/master/img/contribute.gif)](https://github.com/libp2p/community/blob/master/CONTRIBUTE.md)

## License

[MIT](LICENSE) © 2023 Guilherme Versiani
