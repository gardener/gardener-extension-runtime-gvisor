############# builder
FROM golang:1.21.2 AS builder

ARG EFFECTIVE_VERSION
WORKDIR /go/src/github.com/gardener/gardener-extension-runtime-gvisor
COPY . .
RUN make install EFFECTIVE_VERSION=$EFFECTIVE_VERSION

############# binaries-installer
FROM alpine:3.18.3 AS binaries-installer

COPY hack/ hack/
COPY GVISOR_VERSION ./
RUN hack/install-binaries.sh $(cat GVISOR_VERSION)

############# gardener-extension-runtime-gvisor
FROM gcr.io/distroless/static-debian11:nonroot AS gardener-extension-runtime-gvisor
WORKDIR /

COPY charts /charts
COPY --from=builder /go/bin/gardener-extension-runtime-gvisor /gardener-extension-runtime-gvisor
ENTRYPOINT ["/gardener-extension-runtime-gvisor"]

############# gardener-extension-runtime-gvisor-installation for the installation daemonSet
FROM alpine:3.18.3 AS gardener-extension-runtime-gvisor-installation

COPY --from=binaries-installer /usr/local/bin/containerd-shim-runsc-v1 /var/content/containerd-shim-runsc-v1
COPY --from=binaries-installer /usr/local/bin/runsc /var/content/runsc
