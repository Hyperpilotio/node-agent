#!/usr/bin/env python

# node-agent runs in host-network, and influxdb runs in pod-network.
# node-agent cannot communicate to influxdb using Service.
# This script will find Pod IP of each Pod and fill in /etc/node_agent/tasks.json.

import errno
import os
import sys
from jinja2 import Template
from kubernetes import client, config


class Accessor(object):
    def env(self, env_name):
        return os.environ[env_name]

    def deployment_id(self):
        """Call kubernetes api container to retrieve the deployment id"""
        try:
            config.load_incluster_config()
            nodes = client.CoreV1Api().list_node(watch=False)
            if len(nodes.items) > 0:
                return nodes.items[0].metadata.labels.get("hyperpilot/deployment", "")
        except config.ConfigException:
            print("Failed to load configuration. This container cannot run outside k8s.")
            sys.exit(errno.EPERM)

    def k8s_service(self, service_name, namespace='default'):
        """Call kubernetes api service to get service cluster ip"""
        try:
            config.load_incluster_config()
            pod_service = client.CoreV1Api().read_namespaced_service(service_name, namespace)
            cluster_ip = pod_service.spec.cluster_ip
            port = pod_service.spec.ports[0].port
            url = "http://%s:%s" % (cluster_ip, port)
            print("Replacing k8s service %s to url %s" % (service_name, url))
            return url
        except config.ConfigException:
            print("Failed to load configuration. This container cannot run outside k8s.")
            sys.exit(errno.EPERM)

    def pod_ip_label_selector(self, label_selector, namespace='default'):
        """Call kubernetes api service to get pod ip"""
        try:
            print("pod_ip_label_selector init label_selector: %s, namespace: %s" % (
                label_selector, namespace))
            config.load_incluster_config()
            result = client.CoreV1Api().list_namespaced_pod(
                namespace, label_selector=label_selector)
            pod_name = result.items[0].metadata.name
            pod = client.CoreV1Api().read_namespaced_pod(pod_name, namespace)
            pod_ip = pod.status.pod_ip
            print("pod_name: %s, label_selector: %s, namespace: %s, pod_ip: %s" % (
                pod_name, label_selector, namespace, pod_ip))
            return pod_ip
        except config.ConfigException:
            print("Failed to load configuration. This container cannot run outside k8s.")
            sys.exit(errno.EPERM)

    def pod_ip_env_name(self, pod_env_name, namespace='default'):
        """Call kubernetes api service to get pod ip"""
        try:
            config.load_incluster_config()
            label_selector = "app=%s" % os.environ[pod_env_name]
            return self.pod_ip_label_selector(label_selector)
        except config.ConfigException:
            print("Failed to load configuration. This container cannot run outside k8s.")
            sys.exit(errno.EPERM)


def main():
    config_path = "/etc/node_agent/tasks.json"
    with open(config_path, "r") as f:
        # Double curly braces appears in json too often,
        # so use <%= VAR => expression here instead
        template = Template(f.read(),
                            variable_start_string="<%=",
                            variable_end_string="=>")
        with open(config_path, "w") as f:
            template_values = {
                'a': Accessor(),
            }
            f.write(template.render(template_values))

    # Success
    sys.exit(0)


if __name__ == '__main__':
    main()
