############# builder
FROM golang:1.18.5 AS builder

ARG EFFECTIVE_VERSION
WORKDIR /go/src/github.com/gardener/gardener-extension-runtime-gvisor
COPY . .
RUN make install install-binaries EFFECTIVE_VERSION=$EFFECTIVE_VERSION
############# gardener-extension-runtime-gvisor
FROM gcr.io/distroless/static-debian11:nonroot AS gardener-extension-runtime-gvisor
WORKDIR /

COPY charts /charts
COPY --from=builder /go/bin/gardener-extension-runtime-gvisor /gardener-extension-runtime-gvisor
ENTRYPOINT ["/gardener-extension-runtime-gvisor"]

############# gardener-extension-runtime-gvisor-installation for the installation daemonSet
FROM alpine:3.16.1 AS gardener-extension-runtime-gvisor-installation

COPY --from=builder /usr/local/bin/containerd-shim-runsc-v1 /var/content/containerd-shim-runsc-v1
COPY --from=builder /usr/local/bin/runsc /var/content/runsc
