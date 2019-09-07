FROM ubuntu:xenial
RUN apt-get update && apt-get -y install curl netcat jq
RUN mkdir -p /etc/node_agent

COPY ./bin/linux/node-agent .
COPY ./conf/use-task-test.json /etc/node_agent/tasks.json
COPY ./conf/agent_config.json /etc/node_agent/agent_config.json
CMD ["./node-agent"]