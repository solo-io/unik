# Deploying the Instance Listener

The Instance Listener is a special component of UniK that bootstraps UniK unikernel instances running on certain providers (currently [vSphere](providers/vsphere.md) and [Virtualbox](providers/virtualbox.md)).

The instance listener is a virtual appliance that is meant to run continuously on the provider you are using. If UniK detects configuration for one of the providers that requires use of the instance listener in your `daemon-config.yaml`, it will attempt to automatically deploy and boot the instance listener on that infrastructure when you launch the daemon.

There is no additional configuration necessary to deploy the instance listener. UniK will take the following steps at boot time to deploy the instance listener:
* Download the Instance Listener boot disk from **Amazon S3** (3.3GB)
* Create a vm on the target provider
* Attach the disk and boot the vm

The Instance Listener is an Ubuntu server provisioned with VMWare Tools (in the case of vSphere) or Virtualbox Guest Additions (in the case of Virtualbox) that runs a UDP broadcast server to communicate with and bootstrap UniK instances.

UniK Instances depend on the instance listener at for bootstrapping information when they boot. If your instances are not booting properly, check that the Instance Listener is alive and responding to requests on port `3000`. Don't be afraid to kill the Instance Listener and re-deploy it from scratch (note: this will require you you restart your instances in order to bootstrap them with the new Instance Listener).

The default login/password on the instance listener is `unikinstancelistener`:`unikinstancelistener`. If running the instance listener on a shared network, we recommend `ssh`ing to the machine and changing the default password.

We hope to deprecate the Instance Listener in favor of a lighter-weight solution in future releases of UniK.
