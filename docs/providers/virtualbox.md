# Virtualbox UniK Provider

UniK supports running OSv and rumprun unikernels on Virtualbox.
The virtualbox stub of your `daemon-config.yaml` file should look something like the following:
```yaml
providers:
  #...
  virtualbox:
    - name: my-vbox
      adapter_name: "en0: Wi-Fi (AirPort)"
      adapter_type: bridged
```
or:
```yaml
providers:
  #...
  virtualbox:
    - name: my-vbox
      adapter_name: "vboxnet0"
      adapter_type: host_only
```
Depending on whether you prefer running instances on [HostOnly](https://www.virtualbox.org/manual/ch06.html#network_hostonly) network or [Bridged](https://www.virtualbox.org/manual/ch06.html#network_bridged) mode.

We recommend running with HostOnly networking, as it is guaranteed to support *UDP broadcast*, which is a necessary prequisite for bootstrapping UniK instances (see [instance listerner](../instance_listener.md)). UniK will attach a NAT adapter as a second interface to enable Virtualbox instances to reach the internet.

UniK stores Virtualbox data in the following paths:
* JSON representation of the state: `$HOME/.unik/virtualbox/state.json`
* Images (boot vmdks, copied when an instance is launched): `$HOME/.unik/virtualbox/images/`
* Instances (contains Virtualbox folder for each instance, plus the copy of the original boot image): `$HOME/.unik/virtualbox/instances/`
* Volumes (mountable volumes which will persist after Instances are removed): `$HOME/.unik/virtualbox/volumes/`

If UniK gets into a bad state (i.e. you manually remove a file or Virtualbox VM), you should manually edit the `$HOME/.unik/virtualbox/state.json` file to remove the instance that no longer exists. UniK will eventually become self-correcting to deal with disruptions in the state.
