# Plan Generation

This is a simple utility used for generating simulator plans.

At a high level, the plan defines the parameters used below (the source code of the simulator is of
course the authoritative reference for the plans):

```
for $lifecycle.pu-iterations:
    go:
        for $pus:
            wait jitter($jitter.pu-start)
            create PU
            start PU
            wait jitter($jitter.pu-report)
        for $flows:
            wait jitter($jitter.flow-report)
            append flow to flows
        for $lifecycle.flow-iterations:
            wait $lifecycle.flow-interval (or 10s if invalid)
            report all flows
        if $lifecycle.pu-cleanup != "never":
            wait $lifecycle.pu-cleanup
            for $pus:
                stop PU
                wait jitter($jitter.pu-report)
                destroy PU

    wait $lifecycle.pu-interval (or 10s if invalid)
```

One must give a yaml configuration file as represented in the `config.example.yaml`.

**NOTE:** `plan.example.yaml` is an example of the plans generated.

Build:

```bash
make
```

Edit a configuration file:

```bash
cp config.example.yaml config.yaml
```

Run:

```bash
./plan-gen --config config.yaml --output plan.yaml
```

Buil Docker image:
```
make docker
```
