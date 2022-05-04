#! /bin/bash

TOTAL_ENFORCERS=250
NS_PREFIX="simulator"
CLEANUP=false
KUBECTL="kubectl"
CHARTS="$(ls -v enforcer-sim-*.tgz | tail -n 1)"

# These values refer to namespace/policies managment
PREPARE_BACKEND=true
POLICIES="policies"
NAMESPACE_CAPACITY=

COUNT_FLOWS=false
FLOW_COUNT_INTERVAL="5m"

# These values are needed on the script to identify the batch size
PODS=25
# rails support for namespace structure.
declare -A RAILSPODS=([public]=0 [private]=0 [protected]=0)

SIMULATORS=10
INIT_DELAY=5
EXTRA=0

APORETO_BASE_NAMESPACE=""
KUBECONFIG=""
APOCTL_CREDENTIALS="apoctl.json"

SECRET=""
ESTIMATE_ENFORCERS=false

NAMESPACE_DELAY=5              # the delay to wait after a namespace is created.
ENFORCER_CONVERGE_WAIT_TIME=60 # the time for waiting the enforcers to converge
# is approximattely ENFORCER_CONVERGE_WAIT_TIME * 5 seconds.
PODS_READY_TIMEOUT=300 # The time to wait before exit with failing status,
# until the pods get Read

PROG="$0"

USAGE=$(
  cat <<-EOM
This is an expiremental tool to manipulate a simulator test. Some prerequisites
are: apoctl, kubectl and helm in the \$PATH. For different tweaks other binaries
might be needed, such as jq.

Usage: $PROG [OPTIONS]

  OPTIONS:

    --enforcers ARG          The number of enforcers to deploy. Default: $TOTAL_ENFORCERS.
    --kubeconfig ARG         The kubeconfig file, that kubectl will use to provision the cluster. Default: $KUBECONFIG.
    --cleanup                If this set, it will delete all provisioned jobs to the client.
    -n, --namespace ARG      The aporeto base namespace, under which the test will run (Required).
    --prefix ARG             The prefix to use for the kubernetes and aporeto namespaces. Default: $NS_PREFIX.
    --pods ARG               The pods to create per batch. This overrides any configuration file. Default: $PODS.
    --simulators ARG         The simulators per pod to create. This overrides any configuration file. Default: $SIMULATORS.
    --delete-failed ARG      Delete all 'Failed' pods under 'ARG' namespace.
    --charts ARG             Path to the charts to use for helm templating. Default: $CHARTS
    --creds ARG              Path to aporeto application credentials. Default:
$APOCTL_CREDENTIALS.
    --secret ARG             This expects a path to the docker config authentication file which is logged in a private registry.
    --no-prepare             If set, will not create the namespaces/mapping policies.
    --capacity ARG           The namespace maximum capacity for enforcers registration. Default: batch size. This configuration is ignored if rails are defined.
    --extra                  Number of extra namespaces to configure with the test.Default: 0.
    --k8sns                  Kubernetes namespace for cleanup. This param is used only with --clenaup switch.
    --public                 Number of simulators in public rail.
    --private                Number of simulators in private rail.
    --protected              Number of simulators in protected rail.
    To activate rails model at least one or more counts on public, private & protected must be defined. Capacity and rails are mutually exclusive. If both are specified, only rails configuration is used for tests.
    -h, --help               Prints the usage.

EOM
)

usage() {
  printf "$USAGE\n\n"
  exit $1
}

while (($#)); do
  case "$1" in
  --enforcers)
    shift
    if [ ! -z "$1" ]; then
      TOTAL_ENFORCERS="$1"
      shift
    else
      usage 1
    fi
    ;;
  --kubeconfig)
    shift
    if [ ! -z "$1" ]; then
      KUBECONFIG="$1"
      shift
    else
      usage 1
    fi
    ;;
  --cleanup)
    shift
    CLEANUP=true
    ;;
  -n | --namespace)
    shift
    if [ ! -z "$1" ]; then
      APORETO_BASE_NAMESPACE="$1"
      shift
    else
      usage 1
    fi
    ;;
  --prefix)
    shift
    if [ ! -z "$1" ]; then
      NS_PREFIX="$1"
      shift
    else
      usage 1
    fi
    ;;
  --pods)
    shift
    if [ ! -z "$1" ]; then
      PODS="$1"
      shift
    else
      usage 1
    fi
    ;;
  --simulators)
    shift
    if [ ! -z "$1" ]; then
      SIMULATORS="$1"
      shift
    else
      usage 1
    fi
    ;;
  --k8sns)
    shift
    if [ ! -z "$1" ]; then
      # to be used only in cleanup
      nsK8s="$1"
      shift
    else
      usage 1
    fi
    ;;
  --charts)
    shift
    if [ ! -z "$1" ]; then
      CHARTS="$1" # Save the charts path here
      shift
    else
      usage 1
    fi
    ;;
  --creds)
    shift
    if [ ! -z "$1" ]; then
      APOCTL_CREDENTIALS="$1" # Save the path to apoctl credentials here
      shift
    else
      usage 1
    fi
    ;;
  --secret)
    shift
    if [ ! -z "$1" ]; then
      SECRET="$1"
      shift
    else
      usage 1
    fi
    ;;
  --no-prepare)
    shift
    PREPARE_BACKEND=false
    ;;
  --capacity)
    shift
    if [ ! -z "$1" ]; then
      NAMESPACE_CAPACITY=$1
      shift
    else
      usage 1
    fi
    ;;
  --extra)
    shift
    if [ ! -z "$1" ]; then
      EXTRA=$1
      shift
    else
      usage 1
    fi
    ;;
  --public)
    shift
    if [ ! -z "$1" ]; then
      RAILSPODS[public]=$1
      shift
    else
      usage 1
    fi
    ;;
  --private)
    shift
    if [ ! -z "$1" ]; then
      RAILSPODS[private]=$1
      shift
    else
      usage 1
    fi
    ;;
  --protected)
    shift
    if [ ! -z "$1" ]; then
      RAILSPODS[protected]=$1
      shift
    else
      usage 1
    fi
    ;;
  -h | --help)
    shift
    usage 0
    ;;
  *)
    usage 1
    ;;
  esac
done

###############################################################################
# Pre-run checks
is_installed() {
  which $1 >/dev/null || {
    echo "$1 needs to be installed"
    exit 2
  }
}

is_installed "apoctl"
is_installed "kubectl"
is_installed "helm"
is_installed "jq"
is_installed "$POLICIES"

HELM_TEMPLATE="helm template $CHARTS"
if test -f values.yaml; then
  HELM_TEMPLATE="$HELM_TEMPLATE -f values.yaml"
fi
HELM_TEMPLATE="$HELM_TEMPLATE \
  --set pods=$PODS,simulatorsPerPod=$SIMULATORS,initDelay=$INIT_DELAY"

if test -f $APOCTL_CREDENTIALS; then
  APOCTL="apoctl --creds $APOCTL_CREDENTIALS --api-skip-verify"
  POLICIES="$POLICIES --appcred $APOCTL_CREDENTIALS"
else
  echo "$APOCTL_CREDENTIALS dosen't exist"
  exit 3
fi

# If kubeconfig is given, add a flag for kubectl apply
if [ ! -z "$KUBECONFIG" ]; then
  KUBECTL="$KUBECTL --kubeconfig=$KUBECONFIG"
fi

if [ ! -z "$SECRET" ] && test ! -f "$SECRET"; then
  echo "$SECRET doesn't exist"
  exit 3
fi

if ! test -f "$CHARTS"; then
  echo "$CHARTS dosen't exist"
  exit 3
fi

echo
echo "-------------------------------------------"
echo
if [[ "$CLEANUP" = false ]]; then
  echo "Starting a new simulator scale test."
else
  echo "Cleaning up after scale test."
fi

RAILS=false
if [[ ${RAILSPODS[public]} -gt 0 ]] ||
  [[ ${RAILSPODS[private]} -gt 0 ]] ||
  [[ ${RAILSPODS[protected]} -gt 0 ]]; then
  RAILS=true
  echo 'Using rails (public/private/protected) namespace model for these tests.'
else
  BATCH_SIZE=$(($PODS * $SIMULATORS))

  # if capacity is not given as input, set explicitly to batch size.
  if [ -z "$NAMESPACE_CAPACITY" ]; then
    NAMESPACE_CAPACITY=$BATCH_SIZE
  fi
fi

# if max diff is not given as input, set explicitly to batch size.
if [ -z "$MAX_DIFF" ]; then
  MAX_DIFF=$BATCH_SIZE
fi

# DONE with pre-run checks
###############################################################################

# Takes as argument a namespace prefix. All namespaces with that prefix, will be deleted.
# ARGS: prefix (the prefix of the namespaces to clear)
cleanall() {

  pids=()

  $KUBECTL -n $nsK8s delete deployments --all --wait=true
  $KUBECTL -n $nsK8s delete configmaps --all --wait=true
  $KUBECTL -n $nsK8s delete secret --all --wait=true
  $KUBECTL delete namespace $nsK8s
  pids+=("$!")
  echo "Cleanup on Kuberenetes cluster completed !"

  for N in $($APOCTL api list namespace --namespace $APORETO_BASE_NAMESPACE | jq '.[].name' -r); do
    if [[ $N = $APORETO_BASE_NAMESPACE/$1* ]]; then
      echo "Deleting Aporeto namespace $N"
      $APOCTL api delete namespace "$N" --namespace $APORETO_BASE_NAMESPACE
    fi
  done

  for job in "${pids[@]}"; do
    CODE=0
    wait $job || CODE=$?
    if [[ ${CODE} != "0" ]]; then
      echo "cleanup failed on a batch"
    fi
  done

  echo "Cleanup on Aporeto control plane completed !"
  exit 0
}

# delete client side stuff
$CLEANUP && cleanall

# estimate the number of enforcers should be running based on pods currently
# under the namespaces with prefix PREFIX
estimate() {
  pods=0
  for N in $($KUBECTL get namespaces -o jsonpath={.items[*].metadata.name}); do
    if [[ $N = $1* ]]; then
      pods=$(($pods + $($KUBECTL get pods -n $N | grep "Running" | wc -l)))
    fi
  done
  echo "$(($pods * $SIMULATORS))"
  exit 0
}

$ESTIMATE_ENFORCERS && estimate $NS_PREFIX

# delete all pods with not running status.
# ARGS: namespace (the kubernetes namespace)
delete_failed() {
  NS="$1"
  failed_pods_name=$($KUBECTL get pods \
    --field-selector 'status.phase!=Running,status.phase!=Pending' \
    -n $NS -o jsonpath={.items[*].metadata.name})
  echo $failed_pods_name
  echo "$NS"
  for pod in $failed_pods_name; do
    $KUBECTL delete pod $pod -n $NS
  done
  return 0
}

# delete if DELETE_FAILED is set (with the namespace) and exit.
if test ! -z "$DELETE_FAILED"; then
  delete_failed $DELETE_FAILED
  exit 0
fi

# count enforcers using apoctl under $APORETO_BASE_NAMESPACE/namespace
# ARGS: namespace (the kubernetes namespace)
count_enforcers() {

  if test -z "$1"; then
    NAMESPACE="$APORETO_BASE_NAMESPACE"
  else
    NAMESPACE="$APORETO_BASE_NAMESPACE/$1"
  fi

  $APOCTL api count enforcers -r \
    -f 'unreachable == false and operationalStatus == Connected' \
    --namespace $NAMESPACE
}

if test $COUNT; then
  count_enforcers
  exit 0
fi

# waits for initDelay * pods time, and then polls every 5 seconds
# until pods get on 'Running' phase and containers are ready.
# After that, compares the connected enforcers, and compares them with the pods running.
wait_to_stabilize() {

  NS="$2"
  batches=$3

  if [ ! $INIT_DELAY -eq 0 ]; then
    echo "wait initDelay * pods = $(($INIT_DELAY * $PODS)) seconds"
    sleep $(($INIT_DELAY * $PODS))
  fi

  # wait for pods to get connected
  $KUBECTL wait --for condition=Ready -n $NS --all pods --timeout "$PODS_READY_TIMEOUT"s
  if [ ! "$?" -eq "0" ]; then
    echo "not all pods are ready ========="
  fi

  expected_enforcers=$(($PODS * $SIMULATORS))

  # count connected enforcers
  actual_enforcers=$(count_enforcers "$1")

  echo "connected and reachable enforcers are: $actual_enforcers"

  cnt=0
  while test "$actual_enforcers" -lt "$expected_enforcers"; do
    echo "waiting for enforcers to get connected on namespace $APORETO_BASE_NAMESPACE/$1"
    sleep 5
    cnt=$((cnt + 1))
    if test $cnt -eq $ENFORCER_CONVERGE_WAIT_TIME; then
      echo "Warning: enforcers are not converging with the expected ones..."
      echo "Has $actual_enforcers, wants $expected_enforcers"
      break
    fi
    # actual enforcers are the enforcers connected according to
    actual_enforcers=$(count_enforcers "$1")
    # the expected enforcers are the containers on all running pods
    expected_enforcers=$($KUBECTL get pods -n $NS \
      --field-selector 'status.phase==Running' \
      -o jsonpath={.items[*].spec.containers[*].name} | wc -w)
    echo "enforcers in $NS: $actual_enforcers"
  done

  expected_enforcers=$(($PODS * $SIMULATORS * $batches))
  actual_enforcers=$(count_enforcers)
  ! test "$actual_enforcers" -eq "$expected_enforcers" &&
    echo "connected enforcers are: $actual_enforcers, but expects: $expected_enforcers"
}

create_image_secret() {
  NS="$1"
  $KUBECTL create secret generic apo-secret \
    --from-file=.dockerconfigjson="$SECRET" \
    --type=kubernetes.io/dockerconfigjson \
    --namespace $NS
  $KUBECTL patch serviceaccount default \
    -p "{\"imagePullSecrets\": [{\"name\": \"apo-secret\"}]}" \
    --namespace $NS
}

# run $1 until exit status is 0
insist_run() {
  CMD="$1"

  while true; do
    $CMD
    test "$?" -eq "0" && break
    sleep 5
  done
}

# create a random string for the NS_PREFIX
random_string() {
  LEN=${1:-4}
  echo $RANDOM | md5sum | head -c $LEN
}

start_time=$(date)

# Run the scale test
if [[ "$RAILS" = true ]]; then

  enforcers=$((${RAILSPODS[public]} + ${RAILSPODS[private]} + ${RAILSPODS[protected]}))
  echo "Simulators per tenant = $((enforcers))"
  total_tenants=$(($TOTAL_ENFORCERS / $enforcers + $EXTRA))
  echo "Total tenants(+$EXTRA) = $((total_tenants))"
  tenants=0

  k8sNS="$NS_PREFIX-$(random_string 5)"
  echo "creating K8s namespace $k8sNS"
  $KUBECTL create namespace $k8sNS
  TAG_PREFIX=$(random_string 5)

  while [ $tenants -lt $total_tenants ]; do
    tenants=$(($tenants + 1))

    NAMESPACE="$NS_PREFIX-$tenants$(random_string 5)"
    echo
    echo "=== Batch($tenants/$total_tenants) on Aporeto namespace $APORETO_BASE_NAMESPACE/$NAMESPACE"
    echo
    echo "creating namespace $APORETO_BASE_NAMESPACE/$NAMESPACE"
    insist_run "$APOCTL api create namespace -k name $NAMESPACE --namespace $APORETO_BASE_NAMESPACE"

    # Loop over the child namespaces "public" "private", and "protected"
    for NS in "${!RAILSPODS[@]}"; do

      echo "creating namespace $APORETO_BASE_NAMESPACE/$NAMESPACE/$NS"
      insist_run "$APOCTL api create namespace -k name $NS --namespace $APORETO_BASE_NAMESPACE/$NAMESPACE"

      echo "creating enforcer credentials under $APORETO_BASE_NAMESPACE/$NAMESPACE/$NS"
      insist_run "$APOCTL appcred create enforcerd-$NAMESPACE-$NS \
      --namespace $APORETO_BASE_NAMESPACE/$NAMESPACE/$NS --role @auth:role=enforcer" >aporeto.creds

      echo "creating enforcer profile under $APORETO_BASE_NAMESPACE/$NAMESPACE/$NS"
      insist_run "$APOCTL api create enforcerprofiles -k name $NS --namespace $APORETO_BASE_NAMESPACE/$NAMESPACE/$NS"

      sleep 10

      # Create secret on K8s cluster
      $KUBECTL create secret generic enforcerd-$NAMESPACE-$NS --from-file=aporeto.creds --namespace $k8sNS
      test ! -z "$SECRET" && create_image_secret "$k8sNS"

      SIMPODS=10
      REPLICAS=$((${RAILSPODS[$NS]} / $SIMPODS))
      if [[ $((${RAILSPODS[$NS]} % $SIMPODS)) -ne 0 ]]; then
        REPLICAS=$(($REPLICAS + 1))
      fi

      echo "applying configmaps and jobs on $NAMESPACE"
      DEPNAME="$NAMESPACE-$NS"
      $HELM_TEMPLATE --set enforcerTagPrefix="$TAG_PREFIX$(($tenants - 1))" \
        --set k8sNS=\"$k8sNS\" \
        --set depName=\"$DEPNAME\" \
        --set k8sSecret=\"enforcerd-$NAMESPACE-$NS\" \
        --set pods=$REPLICAS \
        --set simulatorsPerPod=$SIMPODS \
        --set enforcerTag="simbase=$TAG_PREFIX-$tenants" |
        $KUBECTL apply -n $k8sNS -f -

      $KUBECTL rollout status deployment $DEPNAME --namespace $k8sNS
    done
  done

else
  $PREPARE_BACKEND && {
    set -e
    $POLICIES --simulators $TOTAL_ENFORCERS --capacity $NAMESPACE_CAPACITY \
      --namespace $APORETO_BASE_NAMESPACE/$NAMESPACE --batch $BATCH_SIZE \
      --multi-mapping --create-namespaces --prefix $TAG_PREFIX
    set +e
  }

  echo "creating enforcer credentials under $APORETO_BASE_NAMESPACE/$NAMESPACE"
  insist_run "$APOCTL appcred create enforcerd --namespace $APORETO_BASE_NAMESPACE/$NAMESPACE --role @auth:role=enforcer" >aporeto.creds
  sleep 5

  enforcers=0
  batches=0

  # Create namespace on K8s cluster
  $KUBECTL create namespace $NAMESPACE
  $KUBECTL create secret generic enforcerd --from-file=aporeto.creds --namespace $NAMESPACE
  test ! -z "$SECRET" && create_image_secret "$NAMESPACE"

  while [ $enforcers -lt $TOTAL_ENFORCERS ]; do
    batches=$(($batches + 1))
    echo "batch: $batches on $NAMESPACE"

    echo "applying configmaps and jobs on $NAMESPACE"
    DEPNAME="$NAMESPACE-pods-$(random_string 6)"
    $HELM_TEMPLATE --set enforcerTagPrefix="$TAG_PREFIX$(($batches - 1))" \
      --set k8sNS=\"$NAMESPACE\" \
      --set k8sSecret=\"enforcerd\" \
      --set depName=\"$DEPNAME\" \
      --set enforcerTag="simbase=$TAG_PREFIX-$batches" |
      $KUBECTL apply -n $NAMESPACE -f -
    echo "configmaps and jobs are applied"

    $KUBECTL rollout status deployment $DEPNAME --namespace $NAMESPACE

    enforcers=$(($enforcers + $BATCH_SIZE))
  done
fi

echo "Starting time: $start_time"
echo "End time: $(date)"
