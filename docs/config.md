# Configuring UniK

1. [Daemon Config](#Daemon Config)
    * [Virtualbox](#Virtualbox)
    * [AWS](#AWS)
    * [vSphere](#vSphere)
3. [CLI Config](#CLI Config)

## Daemon Config
Starting the UniK daemon with `unik daemon` requires a `yaml` file with configuration for each desired [provider](providers/README.md).

By default, `unik daemon` reads from a configuration file located at `$HOME/.unik/daemon-config.yaml`. We recommend placing your config file there. However, you can specify a different config file with `unik daemon --f <path-to-file>`.

UniK requires valid `yaml` matching the following [example](docs/example-daemon-config.yaml):
```
providers:
  aws:
    - name: aws-1
      region: us-west-1
      zone: us-west-1a
  vsphere:
    - name: vsphere-1
      vsphere_user: user
      vsphere_password: password
      vsphere_url: url
      datastore: datastore1
      default_instance_memory: 512
  virtualbox:
    - name: vsphere-1
      adapter_name: vboxnet0
      adapter_type: host_only
version: 0.0.0
```

Include the provider stub for any provider you wish to use. For example, to use only Virtualbox, your `daemon-config.yaml` should look like this.

```
providers:
  virtualbox:
    - name: vsphere-1
      adapter_type: host_only
      adapter_name: vboxnet0
```
If using both AWS and vSphere, the file should look like so:
```
providers:
  aws:
    - name: aws-1
      region: us-west-1
      zone: us-west-1a
  vsphere:
    - name: vsphere-1
      vsphere_user: user
      vsphere_password: password
      vsphere_url: url
      datastore: datastore1
      default_instance_memory: 512
```

### Providers

#### Virtualbox
To run on virtualbox, you will need to tell UniK what type of network card to attach to instances. Available options are `host_only` for [Host-Only Networking](https://www.virtualbox.org/manual/ch06.html#network_hostonly), or `bridged` for [Bridged Networking](https://www.virtualbox.org/manual/ch06.html#network_bridged). UniK must also know the name of the network adapter to use. These are the only properties that virtualbox provider requires. (`name` field is not used currently).

In the Virtualbox stub:
```
  virtualbox:
    - name: any-name-you-want
      adapter_type: ADAPTER_TYPE
      adapter_name: ADAPTER_NAME
```
* Set `ADAPTER_TYPE` to `host_only` or `bridged`
* Set `ADAPTER_NAME` to the name of the adapter. If using `host_only`, you may need to [create a HostOnly network in Virtualbox](http://askubuntu.com/questions/293816/in-virtualbox-how-do-i-set-up-host-only-virtual-machines-that-can-access-the-in).

#### AWS
AWS provider in UniK assumes use of default AWS credential chain. This means either [setting AWS access key id and secret key in your environment](http://docs.aws.amazon.com/aws-sdk-php/v2/guide/credentials.html#environment-credentials), or using the default [AWS configuration file](http://docs.aws.amazon.com/cli/latest/topic/config-vars.html).

Region and zone should be speified like so in the AWS stub:
```
  aws:
    - name: any-name-you-want
      region: us-west-1
      zone: us-west-1a
```

#### vSphere
vSphere provider requires vSphere username, password, url (in the format `http://host_url` or `https://host_url`), as well as the name of the datastore to use for storage.

Default memory is optional and can be used to set the amount of memory allocated to each instance. If it is not set, the default amount used is 512mb per instance.

```
  vsphere:
    - name: any-name-you-want
      vsphere_user: user
      vsphere_password: password
      vsphere_url: url
      datastore: datastore1
      default_instance_memory: 512
```

## CLI Config
After the daemon is running, you can target it through the CLI. To target the daemon, run `unik target --host <host_url>` where `host_url` is the url of the host running the daemon. If running the host on your local machine, you can just use `unik target --host localhost`
