############# builder
FROM golang:1.13.4 AS builder

WORKDIR /go/src/github.com/gardener/gardener-extension-runtime-gvisor
COPY . .
RUN make install-requirements && make VERIFY=true all

############# gardener-extension-runtime-gvisor
FROM alpine:3.11.3 AS gardener-extension-runtime-gvisor

COPY --from=builder /go/bin/gardener-extension-runtime-gvisor /gardener-extension-runtime-gvisor
ENTRYPOINT ["/gardener-extension-runtime-gvisor"]
