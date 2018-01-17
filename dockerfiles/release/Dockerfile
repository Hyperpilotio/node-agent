FROM ubuntu:16.04

RUN mkdir -p /etc/node_agent
RUN apt-get update > /dev/null && \
    apt-get install -y curl gcc python-dev python-pip \
    jq python python-dev build-essential \
    linux-tools-common linux-tools-generic sysstat && \
    rm -rf /var/lib/apt/lists/*

COPY run.sh /usr/local/bin/run.sh
COPY ./bin/linux/node-agent node-agent
COPY nodeAgent_init.py nodeAgent_init.py

COPY requirements.txt requirements.txt
RUN pip install -r requirements.txt && mkdir -p /usr/host

EXPOSE 8181

ENTRYPOINT ["/usr/local/bin/run.sh"]