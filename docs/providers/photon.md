# Photon Controller Provider
UniK supports running VMware-compatibel unikernels on ESXi hosts using the Photon Controller API.

To run UniK instances with Photon, add a Photon stub to your `daemon-config.yaml`:

```yaml
providers:
  #...
  photon:
  - name: my-photon
    photon_url: http://172.16.78.200
    project_id: 3ff7e05b-c16b-4440-8ebb-f5dad7c833de
```

`photon_url` is the url of the photon controller

`project_id` is the id of the project you'd like to use for UniK to create flavors, provision storage, and create VMs / templates.
