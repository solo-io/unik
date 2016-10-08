# UKVM Provider
UniK supports running mirage unikernels through Solo5/UKVM.
In order to run on Solo5/UKVM, you must have `KVM` available on your host.

To run UniK instances with UKVM, add a ukvm stub to your `daemon-config.yaml`:

```yaml
providers:
  #...
  ukvm:
    - name: ukvm-name
      tap_device: tap100
```

To run:

```
unik build --name uk1 --path ./Work/mirage-skeleton-dev/stackv4/  --base mirage --language ocaml --provider ukvm
```

Limitations of UKVM provider:
* Supports only mirage/ocaml
* Need to have KVM enabled
* Prepare a tap device to enable networking.
