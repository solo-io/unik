# Deploying the Instance Listener

The Instance Listener is a special component of UniK that bootstraps UniK unikernel instances running on certain providers (currently [vSphere](providers/vsphere.md) and [Virtualbox](providers/virtualbox.md)).

The instance listener is a virtual appliance that is meant to run continuously on the provider you are using. If UniK detects configuration for one of the providers that requires use of the instance listener in your `daemon-config.yaml`, it will attempt to automatically deploy and boot the instance listener on that infrastructure when you launch the daemon.

There is no additional configuration necessary to deploy the instance listener. UniK **automatically compiles and deploys the instance listener as a unikernel**.

The code for the instance listener

UniK Instances depend on the instance listener at for bootstrapping information when they boot. If your instances are not booting properly, check that the Instance Listener is alive and responding to requests on port `3000`. If the instance listener fails to respond, you can restart it. Or, simply restart the **daemon**, and UniK will automatically re-deploy the instance listener.
