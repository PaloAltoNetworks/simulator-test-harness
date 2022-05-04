# Simulator

This is a utility to enhance the monitoring process on simulator invocation.
The core of this utility are the `helm` charts, which are described in the next
sections.

## PREREQUISITES

In order for the test to run needs:

- A Kubernetes cluster which will be used as a client to install the simulators. This cluster will now be referred as `test harness`. Simulator invocation test expects a kubeconfig file, which points to the client-cluster. The kubeconfig file should be mounted on the `$HOME/.kube/config` for the default behavior. Details on test harness used for testing these
changes are shared below.

- `apoctl` [installed](https://junon.sandbox.aporeto.us/saas/start/apoctl/) in
the `$PATH`.
- `kubectl` [installed](https://kubernetes.io/docs/tasks/tools/install-kubectl/)
in the `$PATH`.
- `helm` [installed](https://github.com/helm/helm/releases) in the `$PATH`.
- `jq` [installed](https://stedolan.github.io/jq) in the `$PATH`.
- Charts to use for enforcer installation (see under `charts/enforcer-sim`).
The default value that the test will seek is `enforcer-sim-<release>.tgz` under
current directory.
- `apoctl.json` (or any other path passed on `--creds` flag) with the application
credentials with namespace administrator privileges:

  ```bash
  # sample example to get application credentials for preprod
  export APOCTL_API=https://api.preprod.aporeto.us
  export APOCTL_NAMESPACE=/aporetodev/username
  eval $(apoctl auth google -e)
  apoctl appcred create "simulators" --role @auth:role=namespace.administrator \
    > apoctl.json
  ```

If a `values.yaml` file exists, it will be used to take default values for the
charts. The pods, simulators and initial delay, will be overwritten by the
values of the script (default 30, 10 and 5 respectively). Also the
`enforcerTagPrefix` will be overwritten by the test from a random value. Here is
an example of a valid `values.yaml` file:

```yaml
# number of pods
pods: 25

# number of simulators per pod
simulatorsPerPod: 10

# number of PUs per simulator
pusPerSimulator: 20

# the type of PUs (a gaia.ProcessingUnitTypeValue or "random")
puType: random

# list of external tags for each PU (must all start with "@"). If null, plan-gen defaults will be
# used.
puMeta: null

# number of flows per PU
flowsPerPU: 50

# Example values for PU lifecycle. These set of values will run each simulator for one hour.
puLife:
  puIter: "1"
  puInterval: 30 # value in seconds
  puCleanup: 1 # value in seconds
  flowIter: "12"
  flowInterval: 60 # value in seconds

# PU jitter
jitter:
  variance: 20 # value in percentage (%)
  puStart: 10 # value in seconds
  puReport: 1 # value in seconds
  flowReport: 500 # value in miliseconds

# enforcer image configuration
image:
  # here needs a full name of the image including registry and organization
  name: "docker.io/aporeto/enforcerd"
  tag: "release-5.0.14"

# plan generation image configuration (the plan gen is incoprorated on the simulator image)
simulatorImage:
  # here needs a full name of the image including registry and organization
  name: "docker.io/aporeto/benchmark"
  tag: "v1.17.0"

# log level
log:
  level: debug
  format: console
  toConsole: 1
  disableWrite: false

# Enforcer jitter configuration
enforcerJitters:
  puSync: "20%"
  puFailureRetry: "20%"
  puStatusUpdateRetry: "20%"
  policiesSync: "20%"
  apiReconnect: "20%"
  flowReportDispatch: "20%"

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
```

### DESCRIPTION

To see available options run:

```shell
simulator.sh -h
```

The test runs in batches of deploying `pods * simulators per pod` enforcer in
each batch. By specifying the total enforcer number you want, the test will
deploy TOTAL_ENFORCERS/BATCH_SIZE + 1 batches.

Each batch is configured as deployment into single test harness namespace. Namespace 
will be created in format `simulator-<6 character random string>`.

A simple example to run:

```shell
# BATCH_SIZE=pods*simulators
simulator.sh --pods 30 --simulators 10 --enforcers 3000 --namespace /base/namespace
```

The default sizes of pods are 30 and the simulators per pod 10. By default the prefix is **simulator**,
and it points to the prefix of kubernetes namespaces that the test create its batches and also are used for the control plane namespaces under a base namespace. The base namespace can be given via the `--namespace` flag.

The test will create one namespace as `simulator-<random>` and n sub-namespaces flat, as follows:

```shell
/base/namespace/simulator-random/simulator-random-1
/base/namespace/simulator-random/simulator-random-2
...
/base/namespace/simulator-random/simulator-random-n
```

The n sub-namespaces will be calculated as the total enforcers, divided by the namespace capacity (`--capacity`). After the namespace creation then a set of mapping policies will be applied from `/base/namespace/simulator-random` towards `/base/namespace/simulator-random/simulator-random-i` for the enforcers.

Thus all the enforcers will be stored under the namespace `/base/namespace/simulator-random`, but will be mapped eventually under the respective namespace to commit to the capacity per namespace.

**NOTE:** If the capacity and batch size are not equal, then will be created
mapping policies as the number of total enforcers.

If one, needs to run another test on the same backend, in order to add more simulators, can use different prefix for isolation purposes:

```shell
simulator.sh --enforcers 3000 \
  --namespace /aporeto/test \
  --prefix complementary-batches
```

### Cleanup

There is a cleanup switch which deletes the aporeto namespace
`/base-namespace/prefix`, as well as the clients premises, on which will delete
all namespaces found with the prefix passed:

```bash
simulator.sh --cleanup --namespace /base/namespace --prefix simulator
```

So the above will delete the aporeto namespace `base/namespace/simulator` and
the cluster namespaces that match the regular expression `simulator*`.

## Scale Test Charts

The simulator script wraps the procedures to deploy the scale tests charts,
which are described as follows.

### Build

To build the charts you can run:

```shell
helm package charts/enforcer-sim
```

### Deploying

You must first create an appcred, the same way you would do to install
an enforcer on Kubernetes:

```shell
apoctl appcred create enforcerd \
  --type k8s \
  --role "@auth:role=enforcer" \
  --namespace /target/namespace \
  | kubectl apply -f -
```

alternatively, if you have a credentials file (**must** be named `aporeto.creds`) locally:

```shell
kubectl create secret generic enforcerd --from-file=aporeto.creds
```

Then, to run the default test:

```shell
helm install nsim enforcer-sim-<release>.tgz
```

**NOTE**: GKE at least, applies by default a limitrange to the `default` namespace for `100m CPU`
per container. You might want to delete (or alter) that to achieve higher simulator density.

### Options

- `image.name` (**required**): enforcer image path (e.i. server/org/imagename)
- `image.tag` (**required**): enforcer image tag
- `simulatorImage.name` (**required**): plan generation image path (e.i. server/org/imagename)
- `simulatorImage.tag` (**required**): plan generation image tag
- `prefix`: prepended to resource names, allowing for multiple deployments in the same namespace
- `pods`: the number of pods to deploy
- `simulatorsPerPod`: the number of simulators to run per pod
- `pusPerSimulator`: the number of PUs per enforcer to simulate
- `puType`: the type of PUs
- `puMeta`: the external tags to add to PUs
- `flowsPerPU`: the number of flows per PU to simulate
- `initDelay`: the delay (in seconds) between bringing up consecutive simulators in a pod
- `restarts`: max simulator restarts before deleting a pod
- `puLife`: the PU lifecycle parameters (see the default values for details)
- `jitter`: the jitter parameters (see the default values for details)
- `log.level`: log level for simulated enforcer
- `log.format`: log format for simulated enforcer
- `enforcerTagPrefix`: prefix of the enforcer tag assigned to all enforcers
- `enforcerTag`: tag to assign to all enforcers

**NOTE**: The total number of simulators will be `pods * simulatorsPerPod`. How to split the
simulators into pods is up to you. Bear in mind that more pods make managing the test a bit easier
(fewer k8s objects) and could prove a bit more efficient (pod overhead and limits on pods / node),
but gives away some flexibility and robustness (e.g. is a simulator reaches the restarts limit, the
whole pod goes down with it). A suggestion is:

- `simulatorsPerPod: 10` for bulk testing e.g. the first few thousand simulators in a large test.
- As low as `simulatorsPerPod: 2` or even `simulatorsPerPod: 2` in the more unstable, later hases of
a test (when simulator timeouts and thus restarts are more likely).

Generally, the limiting factor for simulator density is the simulator's memory consumption. Testing
has shown that a single simulator container will use ~45-55 MiB at runtime. A single `n1-standard-4`
node should be able to handle up to 200 simulators (assuming that the node is not running anything
else).

## Simulator test harness

We have tested the simulator test harness with following configuration. 

Kubernetes cluster configured with two node pools
- default - reserved for system pods, nodes labeled as `pod: system` on `e2-standard-8` images
- workload - to be used for pods, nodes labeled as `pod: workload` on `e2-standard-32` images

Simulator pods will be deployed with toplogy spread constraints on zone and node of maximum skew of 1. This
will allow to have pods in a balanced load cluster.

In our experiments we observed each pod containing 10 simulators to consumbe ~450Mi of memory and ~100m of CPU cycles.
This guidance allowed us pick node sizes & images based on test requirements.

## Simulator plans

The plans for the simulators are generated at runtime by the
[`plan-gen`](https://github.com/aporeto-inc/benchmark-suite/tree/master/utils/plan-gen)
utility of the benchmark suite. This runs in an init container per pod, producing the plans for
that pod's simulators. Refer to that for further details.
