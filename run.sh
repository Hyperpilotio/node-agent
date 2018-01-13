#!/bin/sh

# To install perf tool on Ubuntu 16.04
echo "Install perf tool ..."
apt-get update >/dev/null && apt-get install -y linux-tools-$(uname -r)

echo "args: $@"
[ -f /etc/node_agent/tasks.json ] || (curl -sfL "$1" -o /etc/node_agent/tasks.json)

# Start Snap_init.py in foreground job
echo "Starting python in foreground"
python nodeAgent_init.py
exit_status=$?
echo "exit status: $exit_status"

# Check exit status
if [ $exit_status -ne "0" ]; then
	exit 1 # Unable to load plugin OR create tasks after trialsche. Let k8s to restart this pod
fi

./node-agent
