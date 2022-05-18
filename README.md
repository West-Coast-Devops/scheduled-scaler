# Scheduled Scaler
<!-- ALL-CONTRIBUTORS-BADGE:START - Do not remove or modify this section -->
[![All Contributors](https://img.shields.io/badge/all_contributors-1-orange.svg?style=flat-square)](#contributors-)
<!-- ALL-CONTRIBUTORS-BADGE:END -->
[![Docker Image Version (latest semver)](https://img.shields.io/docker/v/k8srestdev/scaling?style=for-the-badge)](https://hub.docker.com/repository/docker/k8srestdev/scaling) [![Travis (.com) branch](https://img.shields.io/travis/com/West-Coast-Devops/scheduled-scaler/master?style=for-the-badge)](https://travis-ci.com/github/West-Coast-Devops/scheduled-scaler)

In order to use the ScheduledScaler you will need to install the CRD and deploy the Scaling Controller into your Kubernetes cluster.

## Requirements

* Kubernetes Version: 1.7+
* Kubernetes Cluster Settings:
  * "Legacy authorization": "Enabled"

## Tested Environments

* Google Kubernetes Engine
  * Kubernetes Version: 1.9.3-gke.0, 1.7.15
  * Docker Version: 1.12.5
  * Golang Version: 1.9.4

## Getting Started

**Clone this repo**
```
mkdir -p $GOPATH/src/k8s.restdev.com && \
git clone https://github.com/k8s-restdev/scheduled-scaler.git $GOPATH/src/k8s.restdev.com/operators && \
cd $GOPATH/src/k8s.restdev.com/operators
```  
    
**Install using Helm Chart**
```bash
helm install scheduled-scaler artifacts/kubes/scaling/chart
```

> **Note**: This uses the image stored at https://hub.docker.com/r/k8srestdev/scaling by default.
   
[See chart README](artifacts/kubes/scaling/chart) for detailed configuration options 

**Installation without Helm (and compiling binary yourself):**

1. Install the CRD
```
kubectl create -f ./artifacts/kubes/scaling/crd.yml
```
2. Once you have the repo installed on your local dev you can test, build, push and deploy using `make`

> **Note**: If you are just looking for a prebuilt image you can find the latest build [here](https://hub.docker.com/r/k8srestdev/scaling/).
> Just add that image tag to the deployment yml in the artificats dir and apply to your `kube-system` namespace to get up and running without doing a fresh build :D


### Using Make
The `Makefile` provides the following steps:
1. test - Run go unit tests
2. build - Build the go bin file and docker image locally
3. push - Push the built docker image to gcr (or another repository of your choice)
4. deploy - Deploy the updated image to your Kubernetes cluster

Each of these steps can be run in a single pass or can be used individually.

**Examples**

- Do all the things (kubectl)
```
# This example will test, build, push and deploy using kubectl's currently configured cluster
make OPERATOR=scaling PROJECT_ID=my_project_id
```

- Do all the things (kubernodes)
```
# This example will test, build, push and deploy using kubernodes
make OPERATOR=scaling PROJECT_ID=my_project_id DEPLOYBIN=kn KN_PROJECT_ID=my_kubernodes_project_id
```
> **Note:** You only need to add `KN_PROJECT_ID` if it differs from `PROJECT_ID` 

- Just build the image
```
make build OPERATOR=scaling PROJECT_ID=my_project_id
``` 

- Just push any image
```
make push IMAGE=myrepo/myimage:mytag
```

- Just deploy any image (kubectl)
```
make deploy OPERATOR=scaling IMAGE=myrepo/myimage:mytag
```

- Just deploy any image (kubernodes)
```
make deploy OPERATOR=scaling IMAGE=myrepo/myimage:mytag DEPLOYBIN=kn KN_PROJECT_ID=my_kubernodes_project_id
``` 

Now that you have all the resources required in your cluster you can begin creating ScheduledScalers.

## Scheduled Scaler Spec

> **Note:** This controller uses the following [Cron Expression Format](https://godoc.org/github.com/robfig/cron#hdr-CRON_Expression_Format)

### HPA

```yaml
apiVersion: "scaling.k8s.restdev.com/v1alpha1"
kind: ScheduledScaler
metadata:
  name: my-scheduled-scaler-1
spec:
  timeZone: America/Los_Angeles
  target:
    kind: HorizontalPodAutoscaler
    name: my-hpa
    apiVersion: autoscaling/v1
  steps:
  #run at 5:30am PST
  - runat: '0 30 5 * * *'
    mode: range
    minReplicas: 1
    maxReplicas: 5
```

### Instance Group

```yaml
apiVersion: "scaling.k8s.restdev.com/v1alpha1"
kind: ScheduledScaler
metadata:
  name: my-scheduled-scaler-2
spec:
  timeZone: America/Los_Angeles
  target:
    kind: InstanceGroup
    name: my-instance-group-name
    apiVersion: compute/v1
  steps:
  #run at 5:30am PST
  - runat: '0 30 5 * * *'
    mode: fixed
    replicas: 3
```

As you'll see above, you can target either instance groups (if you are on GKE) or hpa, but all the other options are the same.

## Options

| Option | Description | Required |
|--|--|--|
| spec.timeZone | Timezone to run crons in | False |
| spec.target.kind | Type of target (InstanceGroup/HorizontalPodAutoscaler) | True
| spec.target.name | Name of the target resource | True
| spec.target.apiVersion | API Version of the target | True
| spec.steps | List of steps | True
| spec.steps[].runat | Cronjob time string (gocron) | True
| spec.steps[].mode | Type of scaling to run (fixed/range) | True
| spec.steps[].replicas | Defined if mode is 'fixed' | False
| spec.steps[].minReplicas | Defined if mode is 'range' | False
| spec.steps[].maxReplicas | Defined if mode is 'range' | False

For more details on how this add-on can be used please follow the link below:
[Learn More...](http://k8s.restdev.com/p/scheduled-scaler.html)

## Contributors ✨

Thanks goes to these wonderful people ([emoji key](https://allcontributors.org/docs/en/emoji-key)):

<!-- ALL-CONTRIBUTORS-LIST:START - Do not remove or modify this section -->
<!-- prettier-ignore-start -->
<!-- markdownlint-disable -->
<table>
  <tr>
    <td align="center"><a href="https://github.com/scr-oath"><img src="https://avatars.githubusercontent.com/u/41922797?v=4?s=100" width="100px;" alt=""/><br /><sub><b>Sheridan C Rawlins</b></sub></a><br /><a href="#maintenance-scr-oath" title="Maintenance">🚧</a> <a href="https://github.com/West-Coast-Devops/scheduled-scaler/commits?author=scr-oath" title="Tests">⚠️</a> <a href="https://github.com/West-Coast-Devops/scheduled-scaler/commits?author=scr-oath" title="Code">💻</a></td>
  </tr>
</table>

<!-- markdownlint-restore -->
<!-- prettier-ignore-end -->

<!-- ALL-CONTRIBUTORS-LIST:END -->

This project follows the [all-contributors](https://github.com/all-contributors/all-contributors) specification. Contributions of any kind welcome!