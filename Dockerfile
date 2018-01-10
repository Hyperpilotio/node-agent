FROM golang:1.9

WORKDIR /go/src/github.com/hyperpilotio/node-agent
RUN go get github.com/Masterminds/glide

COPY . .

RUN make install_deps
RUN make build-in-docker



FROM ubuntu:xenial
RUN apt-get update && apt-get -y install curl
RUN mkdir -p /etc/node_agent

COPY --from=0 /go/src/github.com/hyperpilotio/node-agent/bin/linux/node-agent .
COPY --from=0 /go/src/github.com/hyperpilotio/node-agent/conf/disk-task-test.json /etc/node_agent
CMD ["./node-agent"]