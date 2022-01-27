############# builder
FROM golang:1.17.5 AS builder

WORKDIR /go/src/github.com/gardener/gardener-extension-runtime-gvisor
COPY . .
RUN make install install-binaries
############# gardener-extension-runtime-gvisor
FROM alpine:3.15.0 AS gardener-extension-runtime-gvisor

COPY charts /charts
COPY --from=builder /go/bin/gardener-extension-runtime-gvisor /gardener-extension-runtime-gvisor
ENTRYPOINT ["/gardener-extension-runtime-gvisor"]

############# gardener-extension-runtime-gvisor-installation for the installation daemonSet
FROM alpine:3.15.0 AS gardener-extension-runtime-gvisor-installation

COPY --from=builder /usr/local/bin/containerd-shim-runsc-v1 /var/content/containerd-shim-runsc-v1
COPY --from=builder /usr/local/bin/runsc /var/content/runsc
