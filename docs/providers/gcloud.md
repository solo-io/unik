# Google Compute UniK Provider

UniK supports running Golang rumprun unikernels on Google Compute.
The Google Compute stub of your `daemon-config.yaml` file should look something like the following:
```yaml
providers:
  #...
  gcloud:
    - name: my-gcloud
      project_id: google-project-id-1234
      zone: us-east1-d
```

UniK requires that your Google Cloud credentials are set via [`gcloud auth`](https://cloud.google.com/sdk/gcloud/reference/auth/).

UniK stores Google Compute data in the following paths:
* JSON representation of the state: `$HOME/.unik/gcloud/state.json`

* UniK boot volumes are stored as Images
* UniK data volumes are not currently supported for Google Cloud (PRs on this are welcome).
* UniK instances are `g1-small` EC2 Instances
