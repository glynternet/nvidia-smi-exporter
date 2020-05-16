FROM golang:1.10 as build
WORKDIR /go/src/nvidia-smi-exporter
ADD . /go/src/nvidia-smi-exporter
RUN go install .

FROM gcr.io/distroless/base
COPY --from=build /go/bin/nvidia-smi-exporter /
EXPOSE 9101:9101
ENTRYPOINT ["/nvidia-smi-exporter"]
