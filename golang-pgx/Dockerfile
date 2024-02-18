FROM golang:1.22-bullseye as builder

COPY . /workdir
WORKDIR /workdir

ENV CGO_ENABLED=0

RUN go build -ldflags "-s -w" -trimpath -o app cmd/api/main.go

# https://iximiuz.com/en/posts/containers-distroless-images/
FROM gcr.io/distroless/base-debian11:nonroot

COPY --from=builder /workdir/app /bin/app

# USER 65534
USER nonroot

ENTRYPOINT ["/bin/app"]
