#FROM golang:1.9
#
#WORKDIR /go/src/github.com/hyperpilotio/node-agent
#RUN go get github.com/Masterminds/glide
#
#COPY . .
#
#RUN make install_deps
#RUN make build-release



FROM ubuntu:xenial
RUN apt-get update && apt-get -y install curl
RUN mkdir -p /etc/node_agent

COPY ./bin/linux/node-agent .
COPY ./conf/use-task-test.json /etc/node_agent/tasks.json
CMD ["./node-agent"]