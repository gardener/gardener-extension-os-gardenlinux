############# builder
FROM golang:1.16.2 AS builder

WORKDIR /go/src/github.com/gardener/gardener-extension-os-gardenlinux
COPY . .
RUN make install-requirements && make generate && make install

############# gardener-extension-os-gardenlinux
FROM alpine:3.13.5 AS gardener-extension-os-gardenlinux

COPY --from=builder /go/bin/gardener-extension-os-gardenlinux /gardener-extension-os-gardenlinux
ENTRYPOINT ["/gardener-extension-os-gardenlinux"]
