# vSphere UniK Provider

UniK supports running rumprun unikernels on vSphere.
The vSphere stub of your `daemon-config.yaml` file should look something like the following:
```yaml
providers:
  #...
  vsphere:
  - name: vsphere-1
    vsphere_user: user
    vsphere_password: password
    vsphere_url: url
    datastore: datastore1
    default_instance_memory: 512
```

Running on vSphere requires the host network to support UDP broadcast (see [instance listerner](../instance_listener.md)). Instances that launch on vSphere without access to UDP broadcast will fail to bootstrap.

UniK stores a JSON representation of the state in the local `$HOME/.unik/vsphere/state.json`

UniK stores files for running virtual machines in the following folders in the configured datastore:
* Images (boot vmdks, copied when an instance is launched): `[datastore_name] unik/vsphere/images`
* Instances (contains vSphere folder for each instance, plus the copy of the original boot image): `[datastore_name] unik/vsphere/instances`
* Volumes (mountable volumes which will persist after Instances are removed): `[datastore_name] unik/vsphere/volumes`

If UniK gets into a bad state (i.e. you manually remove a file or vSphere VM), you should manually edit the `$HOME/.unik/vsphere/state.json` file to remove the instance that no longer exists. UniK will eventually become self-correcting to deal with disruptions in the state.
