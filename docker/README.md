## SIMULATOR TEST

### REQUIREMENTS

---

- A Kubernetes cluster and a kubeconfig file with access to that cluster.
- A private registry to upload Docker images.
- Kubernetes cluster configured to access to images.
- An Aporeto control plane and app credentials with the permissions to create child namespaces and add enforcers.
- An Aporeto `enforcer` image accessible via Docker registry e.g. docker.io/aporeto/enforcerd:release-3.14.4
- Docker installed with user privileges to run, tag, upload images.

### Provided in Test Harness

- A `values.yaml` file for the chart parameters.
- The simulator image which can be found at `docker.io/aporeto/benchmark`

**NOTE:** Generally, the limiting factor for simulator density is the simulator's memory consumption. Internal testing on GKE has shown that a single simulator container will use **~45-55 MiB** at runtime. A dedicated GCP **n1-standard-4** or equivalent node should be able to handle up to **200** simulators.

### RUNNING THE TEST
---

1. Get the latest version of apoctl as per the instructions at: https://docs.aporeto.com/saas/start/apoctl/

2. Pull the latest image: `docker.io/aporeto/benchmark:v1.17.0`

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
# prefix to add to resource names. This allows to have multiple deployments in the same namespace.
prefix: ""
# number of pods
pods: 5
# number of simulators per pod
simulatorsPerPod: 5
# number of PUs per simulator
pusPerSimulator: 10
# the type of PUs (a gaia.ProcessingUnitTypeValue or "random")
puType: Docker
# list of external tags for each PU (must all start with "@"). If null, plan-gen defaults will be
# used.
puMeta: null
# number of flows per PU
flowsPerPU: 10
# initial simulator bring-up delay (seconds). This is additive for each pod, i.e. first pod will
# fire immediately, second after initDelay, third after 2 * initDelay etc.
initDelay: 0
# the max number of simulator restarts. After this is reached, the pod with the failing simulator
# will be deleted (when a simulator fails, it is restarted with an exponential backoff, capped at
# 5m).
restarts: 6
# Example values for PU lifecycle. These set of values will run each simulator for one hour.
puLife:
  puIter: "1"
  puInterval: 30 # value in seconds
  puCleanup: 1   # value in seconds
  flowIter: "12" # can be infinite
  flowInterval: 60 # value in seconds
# PU jitter
jitter:
  variance: 20    # value in percentage (%)
  puStart: 10     # value in seconds
  puReport: 1     # value in seconds
  flowReport: 500 # value in miliseconds
# enforcer image configuration
image:
  # here needs a full name of the image including registry and organization
  name: "docker.io/aporeto/enforcerd"
  tag: "release-3.14.4"
# plan generation image configuration
simulatorImage:
  # here needs a full name of the image including registry and organization
  name: "docker.io/aporeto/benchmark"
  tag: "v1.17.0"
# log level
log:
  level: info
  format: console
# By default 2 tags are added to enforcer (simulator) with the current charts.
# The first one has enforcerTagPrefix as prefix, as nsim=<prefix>-$i where i ranges
# through pods * simulatorsPerPod.
enforcerTagPrefix: simulator
# The second is the enforcerTag, it is fixed for all enforcers and should
# always be key=value.
enforcerTag: simbase=simulator
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
The test runs in batches of deploying `pods * simulators per pod` enforcers in each batch.
By specifying the enforcers, the test will deploy `TOTAL_ENFORCERS/BATCH_SIZE + 1` batches.

**NOTES:**
- The default size of pods per container is **30** and the simulators per pod **10**. These default values are a good option as per our previous tests and profiling. The values can be changed by specifying the `--pods` and `--simulators` respective flags.
- If a `values.yaml` file exists, it will be used to take default values for the charts. The pods, simulators and initial delay, will be overwritten by the values of the script (default 30, 10 and 5 respectively). A provided `values.yaml` file is inside the `files` directory, and it's the one used in the examples above.

Another important tweak is the `prefix`. By default the prefix is **simbatch**, and it points to the prefix of kubernetes namespaces that the test create its batches and also are used for the control plane namespaces under a base namespace. The base namespace can be given via the `--namespace` flag.

The test will create one namespace as `simbatch-random` and n sub-namespaces flat, as follows:
```
/base/namespace/simbatch-random/simbatch-random-1
/base/namespace/simbatch-random/simbatch-random-2
...
/base/namespace/simbatch-random/simbatch-random-n
```
The n sub-namespaces will be calculated as the total enforcers wanted, divide by
the namespace capacity (`--capacity`). After the namespace creation then a set
of mapping policies will be applied from `/base/namespace/simbatch-random`
towards `/base/namespace/simbatch-random/simbatch-random-i` for the enforcers.
Thus all the enforcers will be stored under the namespace
`/base/namespace/simbatch-random`, but will be mapped eventually under the
respective namespace to commit to the capacity per namespace.

Besides the aporeto control plane, also **m kubernetes namespaces** will be
created which are bound to the simulators batches to register:
```
simbatch-random-1
simbatch-random-2
...
simbatch-random-n
```
If one, needs to run another test on the same backend, in order to add more
simulators, can use different prefix for isolation purposes:

```shell
simulator.sh --enforcers 3000 \
  --namespace /aporeto/test \
  --prefix complementary-batches
```
In order to cleanup the test there is a `--cleanup` flag which can be used as shown in the example
below:
```shell
simulator.sh --cleanup \
  --namespace /your/base/namespace \
  --prefix simbatch
```
The above command will delete the namespaces that start with the prefix in both the Aporeto backend as well as the kubernetes namespaces. It deletes all the namespaces that match the regular expression `simbatch*`.

### OTHER NOTES
---

Depending on the cloud provider where k8s runs on there might be a few different ways of authenticating when using kubectl.

#### GOOGLE CLOUD

For GKE the gcloud-sdk tool is required which is already installed in the image.

#### IAM AUTHENTICATOR

For EKS clusters `aws-iam-authenticator` is required which is already installed in the image.
