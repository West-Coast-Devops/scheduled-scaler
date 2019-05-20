# Helm Chart for scheduled Scaler

## Introduction

This chart bootstraps a scheduled-scaler controller and allows the deployment of `ScheduledScaler` resources to your cluster.

For more information read [scheduled-scaler overview](http://k8s.restdev.com/p/scheduled-scaler.html)

## Prerequisites

* Kubernetes Version: 1.7+
* Kubernetes Cluster Settings:
    * "Legacy authorization": "Enabled"

## Installing the Chart

```bash
$ helm install . --name scheduled-scaler
```

## Uninstalling the Chart

```bash
$ helm delete --purge scheduled-scaler
```

## Configuration

| Parameter                               | Description                                                         | Default                              |
|-----------------------------------------|---------------------------------------------------------------------|--------------------------------------|
| `image.repository`                      | Scheduled scaler container image                                    | `k8srestdev/scaling`                 |
| `image.tag`                             | Scheduled scaler container image tag                                | `0.0.2`                              |
| `image.pullPolicy`                      | Scheduled scaler container image pull policy                        | `Always`                             |
| `replicaCount`                          | Number of scheduled-scaler replicas to create (only 1 is supported) | `1`                                  |
| `sslCerts.hostPath`                     | TLS certs for secure connections                                    | `/etc/ssl/certs`                     |
| `rbac.create`                           | install required RBAC service account, roles and rolebindings       | `true`                               |
| `resources`                             | Resource configuration for Scheduled scaler controller pod          | `{}`                                 |
| `nodeSelector`                          | Node labels for Scheduled scaler controller pod assignment          | `{}`                                 |
| `tolerations`                           | Tolerations for Scheduled scaler controller pod assignment          | `[]`                                 |
| `affinity`                              | Affinity Rules for Scheduled scaler controller pod assignment       | `[]`                                 |

## RBAC

By default the chart will install the recommended RBAC roles and rolebindings.

To determine if your cluster supports this running the following:

```bash
$ kubectl api-versions | grep rbac
```

You also need to have the following parameter on the api server. See the following document for how to enable RBAC

```bash
--authorization-mode=RBAC
```

If the output contains "beta" or both "alpha" and "beta" you can may install rbac by default, if not, you may turn RBAC off as described below.

### RBAC role/rolebinding creation

RBAC resources are enabled by default. To disable RBAC do the following:

```bash
$ helm install . --name scheduled-scaler --set rbac.create=false
```
