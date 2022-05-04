## SIMULATOR TEST

### REQUIREMENTS

---

- A Kubernetes cluster and a kubeconfig file with access to that cluster.
- A private registry to upload Docker images.
- Kubernetes cluster configured to access to images.
- An Aporeto control plane and app credentials with the permissions to create child namespaces and add enforcers.
- An Aporeto `enforcer` image accessible via Docker registry e.g. docker.io/aporeto/enforcerd:release-5.2.1
- Docker installed with user privileges to run, tag, upload images.

### Provided in Test Harness

- A `values.yaml` file for the chart parameters.
- The simulator image which can be found at `docker.io/aporeto/benchmark`

**NOTE:** Generally, the limiting factor for simulator density is the simulator's memory consumption. Internal testing on GKE has shown that a single simulator container will use **~45-55 MiB** at runtime. A dedicated GCP **n1-standard-4** or equivalent node should be able to handle up to **200** simulators.

### RUNNING THE TEST
---

1. Get the latest version of apoctl

2. Pull the latest image. e.g. `docker.io/aporeto/benchmark:v1.22.1`

3. Create the required files:

  For the appcreds and based on your backend you can use the following command:
  ```shell
  export APOCTL_API=https://your-backend-api.com
  export APOCTL_NAMESPACE=/your/preferred/namespace
  ```
  Make sure you have a valid token:
  ```shell
  apoctl auth verify
  ```
  Create the appcred:
  ```shell
  apoctl appcred create "simulators" --role @auth:role=namespace.administrator > apoctl.json
  ```
4. Edit the values.yaml, and change any value needed.
```yaml
# number of pods
pods: 1

# number of simulators per pod
simulatorsPerPod: 10

# number of PUs per simulator
pusPerSimulator: 10

# the type of PUs (a gaia.ProcessingUnitTypeValue or "random")
puType: random

# list of external tags for each PU (must all start with "@"). If null, plan-gen defaults will be used.
puMeta: null

# number of flows per PU
flowsPerPU: 200

# Example values for PU lifecycle.
puLife:
  puIter: "1"
  puInterval: 30 # value in seconds
  puCleanup: 90 # value in seconds
  flowIter: "infinite" # number of flow iterations
  flowInterval: 60 # value in seconds
  dnsReportRate: 1 # dns reports per PU per minute

# PU jitter
jitter:
  variance: 20 # value in percentage (%)
  puStart: 10 # value in seconds
  puReport: 1 # value in seconds
  flowReport: 500 # value in miliseconds

# enforcer image configuration
image:
  name: gcr.io/aporetodev/enforcerd
  # tag: master-staged
  # tag: "pr-2640"
  # tag: "release-6.12.0-staged"
  tag: "release-5.2.0-staged"

# plan generation image configuration
simulatorImage:
  name: gcr.io/aporetodev/simulator
  # tag: master-staged
  tag: "v0.0.0-dev"
  #tag: "pr-688"

# log level
log:
  level: debug
  format: console
  toConsole: 1
  disableWrite: false

# Enforcer configuration flags.
enforcerOpts:
  policiesSync: "20m"
  certRenewal: "96h"
  tagSync: "1m"
  handleOfflineAPI: false


# Enforcer jitter configuration
enforcerJitters:
  puSync: "20%"
  puFailureRetry: "20%"
  puStatusUpdateRetry: "20%"
  policiesSync: "20%"
  apiReconnect: "20%"
  flowReportDispatch: "20%"
  certRenewal: "20%"
  tagSync: "20%"

# By default 2 tags are added to enforcer (simulator) with the current charts.
# The first one has enforcerTagPrefix as prefix, as nsim=<prefix>-$i where i ranges
# through pods * simulatorsPerPod.
enforcerTagPrefix: simulator
# The second is the enforcerTag, it is fixed for all enforcers and should
# always be key=value.
enforcerTag: simbase=simulator

# Do not edit these parameters, exclusively used by automation. Value set here will be ignored.
depName: ""
k8sNS: ""
k8sSecret: ""

```

9. Create a container from the image and mount the required files:
```shell
export SIMULATOR_PATH=/root/simulator
docker run -it --rm \
    -v $(pwd)/apoctl.json:${SIMULATOR_PATH}/apoctl.json \
    -v $(pwd)/values.yaml:${SIMULATOR_PATH}/values.yaml \
    -v $(pwd)/kubeconfig.yaml:/root/.kube/config \
    simulator:latest
```
10. Once inside the container you can view the available options of the simulator test:
```shell
simulator.sh -h
```
or start a test:
```shell
simulator.sh --enforcers 3000 --namespace /your/base/namespace
```

### NOTES
---

Depending on the cloud provider where k8s runs on there might be a few different ways of authenticating when using kubectl.

#### GOOGLE CLOUD

For GKE the gcloud-sdk tool is required which is already installed in the image.

#### IAM AUTHENTICATOR

For EKS clusters `aws-iam-authenticator` is required which is already installed in the image.
