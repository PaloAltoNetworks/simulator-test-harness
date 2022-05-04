# Benchmark Suite

The benchmark-suite contains libraries to aid test creation as well as a set of
tests.

Basic directories:

- `libs`: A set of packages to aid benchmark test creation.
- `test`: A set of benchmark tests.
- `utils`: A set of utilties related to benchmarking.

## DOCKER EXAMPLES

Tests can run in a Docker container too. Currently, the tests supported are:

- Control plane tests
  - API authorization policy tests.
  - Network access policy tests.
- Datapath tests
  - Throughput tests.
- Simulator scale test

### CONTROL PLANE TESTS

1. Build the tests and the Docker container:
  ```bash
  make docker
  ```
2. Run the desired test, providing its corresponding environment variables. For
   example, to run the API authorization policy test:
  ```bash
  MYCREDS=my-app-creds.json
  TEST_PATH=/root/benchmark-suite/tests
  docker run --rm \
    -e API_URL=https://api.example.com \
    -e APPCRED=$MYCREDS \
    -e ROOT_NS=/example/apipol-test \
    -e TEST_NS=apipol-test-01 \
    -e CLEANUP=yes \
    -e VALIDATE=yes \
    -e MAX_DEPTH=16 \
    -e POL_TOTAL=1000 \
    -v $(pwd)/$MYCREDS:/$TEST_PATH/$appcreds \
    benchmark-suite:latest apipols --mode apipols --log-level debug
  ```
  The above test will run an API authorization policy test, under the
  `/example/apipol-test/apipol-test-01` namespace (depth 16), creating 1000
  policies against the specified API (`api.example.com`).

### THROUGHPUT TESTS

1. Build the tests and the Docker container:
  ```bash
  make docker
  ```
2. Create all the necessary file dependencies and edit configuration values.
   Make sure the specified paths match the docker mount points in the YAML
   files:
  ```bash
  cd tests/throughput
  cp backend_skel.yml backend.yml
  cp config_skel.yml config.yml
  mkdir test-1-results
  # vim config.yml ...
  #   ...
  #   public_key_path: "key.pub"
  #   private_key_path: "key"
  #   ...
  #   service_account_path: "service-account.json"
  #   ...
  #   gcs_backend:
  #     backend:
  #       credentials: "service-account.json"
  #   ...
  #   enforcers:
  #     appcred_path: "enforcerd.creds"
  #   ...
  #   fortio:
  #     ...
  #     result_download_path: "test-1-results"
  #     collector_local_path: "fortio-collector"
  ###
  # vim backend.yml
  #   ...
  #   appcred: "appcred.json"

  export TEST_PATH=/root/benchmark-suite/tests  # This should not change
  export UTIL_PATH=/root/benchmark-suite/utils  # This should not change
  docker run --rm \
    -v $(pwd)/appcred.json:${TEST_PATH}/appcred.json \
    -v $(pwd)/config.yml:${TEST_PATH}/config.yml \
    -v $(pwd)/backend.yml:${TEST_PATH}/backend.yml \
    -v $(pwd)/key:${TEST_PATH}/key \
    -v $(pwd)/key.pub:${TEST_PATH}/key.pub \
    -v $(pwd)/service-account.json:${TEST_PATH}/service-account.json \
    -v $(pwd)/enforcerd.creds:${TEST_PATH}/enforcerd.creds \
    -v $(pwd)/fortio-collector:${TEST_PATH}/fortio-collector \
    -v $(pwd)/test-1-results:${TEST_PATH}/test-1-results \
    benchmark-suite:latest throughput --log-level debug
```

### SIMULATOR TEST

1. Build the Docker container:
  ```bash
  make simulator-docker
  ```

2. Get a valid `apoctl.json` file:
  ```bash
  export APOCTL_API=https://api.preprod.aporeto.us
  export APOCTL_NAMESPACE=/aporeto/username
  eval $(apoctl auth google -e)
  apoctl appcred create "simulators" --role @auth:role=namespace.administrator > apoctl.json
  ```

3. Create all the necessary file dependencies and edit configuration values.
   Make sure the specified paths match the docker mount points in the YAML
   files:
  ```bash
  export SIMULATOR_PATH=/root/simulator # This should not change
  docker run -it --rm \
    -v $(pwd)/apoctl.json:${SIMULATOR_PATH}/apoctl.json \
    -v $(pwd)/values.yaml:${SIMULATOR_PATH}/values.yaml \
    -v $(pwd)/kubeconfig.yaml:${SIMULATOR_PATH}/kubeconfig.yaml \
    -v $(pwd)/backend.yml:${SIMULATOR_PATH}/backend.yml \
    simulator:latest
  ```

4. For more information refer to the utility file [README.md](utils/simulator/README.md)
