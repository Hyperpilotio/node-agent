FROM ubuntu:xenial
RUN apt-get update && apt-get -y install curl
RUN mkdir -p /etc/node_agent

COPY ./bin/linux/node-agent .
COPY ./conf/use-task-test.json /etc/node_agent/tasks.json
CMD ["./node-agent"]