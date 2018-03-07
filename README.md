# Scheduled Scaler

In order to use the ScheduledScaler you will need to install the CRD and deploy the Scaling Controller into your Kubernetes cluster.

### Getting Started
1. Install the CRD
```
kubectl create -f ./artifacts/kubes/scaling/crd.yml
```
2. Build the docker Image
	a. Install godeps (`godep restore`)
	b. Build the image (`./make scaling [PROJECT]`)
3. Deploy the image (`./deploygke [IMAGE] scaling [PROJECT_NAME]`)
	a. *Note: The deploygke script using kubernodes, you may manually deploy using the file in ./artifacts/kubes/scaling/deployment.yml if you prefer* 

Now that you have all the resources required in your cluster you can begin creating ScheduledScalers.

### Scheduled Scaler Spec
**HPA**
```
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
**Instance Group**
```
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
As you'll see above you can target either instance groups or hpa, but all the other options are the same.

### Options
| Option | Description | Required |
|--|--|--|
| spec.timeZone | Timezone to run crons in | False |
| spec.target.type | Type of target (ig/hpa) | True
| spec.target.name | Name of the target resource | True
| spec.steps | List of steps | True
| spec.steps[].runat | Cronjob time string (gocron) | True
| spec.steps[].mode | Type of scaling to run (fixed/range) | True
| spec.steps[].replicas | Defined if mode is 'fixed' | False
| spec.steps[].minReplicas | Defined if mode is 'range' | False
| spec.steps[].maxReplicas | Defined if mode is 'range' | False


For more details on how this add-on can be used please follow the link below:
[Learn More...](http://k8s.restdev.com/p/scheduled-scaler.html)
