# Simualtor Test Harness
## Build

1. Build the Docker container:
  ```bash
  make simulator
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
