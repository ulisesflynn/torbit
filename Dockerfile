FROM golang:1.10 AS builder

# Download and install the latest release of dep
ADD https://github.com/golang/dep/releases/download/v0.4.1/dep-linux-amd64 /usr/bin/dep
RUN chmod +x /usr/bin/dep

# Copy the code from the host and compile it
WORKDIR $GOPATH/src/github.com/ulisesflynn/torbit
COPY Gopkg.toml Gopkg.lock ./
RUN dep ensure --vendor-only
COPY . ./
COPY config.toml /
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix nocgo -o /chat .

FROM scratch
EXPOSE 2000:2000
EXPOSE 8080:8080
COPY --from=builder /chat ./
COPY --from=builder /config.toml ./
COPY --from=builder /tmp /tmp
ENTRYPOINT ["./chat"]