# Command-Line Interface

The UniK cli wraps calls to UniK's [REST API](api.md) to make using UniK easy.

* Managing Unik
  * [`unik daemon`](cli.md#running-the-daemon)
  * [`unik target`](cli.md#targeting-the-unik-daemon)
  * [`unik providers`](cli.md#list-available-providers)
  * [`unik compilers`](cli.md#list-available-compilers)
* Images
  * [`unik build`](cli.md#building-an-image)
  * [`unik images`](cli.md#list-available-images)
  * [`unik describe-image`](cli.md#get-json-representation-of-a-specifig-image)
  * [`unik delete-image`](cli.md#delete-an-image)
* Instances
  * [`unik run`](cli.md#run-an-instance)
  * [`unik instances`](cli.md#list-available-instances)
  * [`unik describe-instance`](cli.md#get-json-representation-of-a-specifig-instance)
  * [`unik delete-instance`](cli.md#delete-an-instance)
  * [`unik stop`](cli.md#power-off-an-instance)
  * [`unik start`](cli.md#power-on-an-instance)
  * [`unik logs`](cli.md#retrieve-or-follow-instance-logs)
* Volumes
  * [`unik create-volume`](cli.md#create-a-volume)
  * [`unik volumes`](cli.md#list-volumes)
  * [`unik attach-volume`](cli.md#attach-a-volume)
  * [`unik detach-volume`](cli.md#detach-a-volume)
  * [`unik delete-volume`](cli.md#delete-a-volume)

#### Running the daemon
The cli is used to start the UniK daemon. To start the daemon:
```
unik daemon
```
It is recommended to start the daemon as a background process with `&` as it is designed to be long-running.

`unik daemon` makes use of the following flags:
  * `--debug`           (bool, optional) more verbose logging for the daemon
  * `--f string`       (string, optional) path to [daemon config file](configure.md) (default is $HOME/.unik/daemon-config.yaml)
  * `--logfile string`   (string, optional) output logs to file (in addition to stdout)
  * `--port int`         (int, optional) listening port for daemon (default 3000)
  * `--trace`            (bool, optional) add stack trace to daemon logs

Example usage:
```
unik daemon --f ./my-config.yaml --port 12345 --debug --trace --logfile logs.txt
```
  * will start the daemon using config file at my-config.yaml
  * running on port 12345
  * debug mode activated
  * trace mode activated
  * outputting logs to logs.txt

---

#### Targeting the UniK daemon
Run
```
unik target --host localhost
```
If running UniK on your local machine. Otherwise
```
unik target --host host_address [--port port]
```
Will set the target to a remote UniK host. Use of the `port` flag is optional and only necessary if the  `daemon` is running on a non-default (3000) port.

Note:
The target for client commands (commands other than `unik daemon`) can be overridden with the `--host` flag (to use a target other than the default).

---

#### List available Providers
```
unik providers
```
Returns a list of providers available to the targeted unik backend.

---

#### List available Compilers
```
unik compilers
```
Returns a list of compilers available to the targeted unik backend.

---

#### Building an image
Compiles source files into a runnable unikernel image.

Images must be compiled for a specific provider, specified with the `--provider` flag
To see a list of available providers, run `unik providers`

A unikernel compiler that is compatible with the provider must be specified with the `--compiler` flag
To see a list of available compilers, run `unik compilers`

If you wish to attach volumes to instances of an image, the image must be compiled in advance
with a list of the expected mount points. e.g. for an application that reads from a '/data' folder,
the unikernel should be compiled with the flag `--mount /data`

Runtime arguments to be passed to your unikernel must also be specified at compile time.
You can specify arguments as a single string passed to the `--args` flag

Image names must be unique. If an image exists with the same name, you can force overwriting with the
--force flag

Example usage:

```
unik build --name myUnikernel --path ./myApp/src --compiler rump-go-xen --provider aws --mountpoint /foo --mountpoint /bar --args 'arg1 arg2 arg3' --force
```
  * will create a Go unikernel named myUnikernel using the sources found in ./myApp/src,
  * compiled using rumprun for the xen hypervisor, targeting AWS infrastructure,
  * expecting a volume to be mounted at /foo at runtime,
  * expecting another volume to be mounted at /bar at runtime,
  * passing 'arg1 arg2 arg3' as arguments to the application when it is run,
  * and deleting any previous existing instances and image for the name myUnikernel before compiling

Another example (using only the required parameters):
```
unik build -name anotherUnikernel -path ./anotherApp/src -compiler rump-vmware -provider vsphere
```

Usage:
  unik build [flags]

Flags:
  *  `--args string`        (string,optional) to be passed to the unikernel at runtime
  *  `--compiler string`    (string,required) name of the unikernel compiler to use
  *  `--force`              (bool, optional) force overwriting a previously existing image with this name
  *  `--mountpoint value`   (string,repeated) specify up to 8 mount points for volumes (default [])
  *  `--name string`        (string,required) name to give the unikernel. must be unique
  *  `--path string`        (string,required) path to root application sources folder
  *  `--provider string`    (string,required) name of the target infrastructure to compile for
  * `--no-cleanup`          (bool, optional) tell UniK not to clean up any artifacts from the build process if building fails. for debugging purposes.

---

#### List available images
```
unik images
```
Lists all available unikernel images across providers. Includes important information for running and managing instances, including the required mount points for the image.

---

#### Get JSON representation of a specifig image:
```
unik describe-image --image IMAGE_NAME
```

---

#### Delete an image
```
unik delete-image --image IMAGE_NAME
```
Use `--force` to force deleting unikernel and associated instances if any instances of this image are currently running.

---

#### Run an instance
```
unik run --instanceName INSTANCE_NAME --imageName IMAGE_TO_USE
```

Deploys a running instance from a unik-compiled unikernel disk image.
The instance will be deployed on the provider the image was compiled for.
e.g. if the image was compiled for virtualbox, unik will attempt to deploy
the image on the configured virtualbox environment.

'unik run' requires a unik-managed volume (see 'unik volumes' and 'unik create volume')
to be attached and mounted to each mount point specified at image compilation time.
This means that if the image was compiled with two mount points, /data1 and /data2,
'unik run' requires 2 available volumes to be attached to the instance at runtime, which
must be specified with the flags --vol SOME_VOLUME_NAME:/data1 --vol ANOTHER_VOLUME_NAME:/data2
If no mount points are required for the image, volumes cannot be attached.

environment variables can be set at runtime through the use of the -env flag.

Example usage:

```
unik run --instanceName newInstance --imageName myImage --vol myVol:/mount1 --vol yourVol:/mount2 --env foo=bar --env another=one --memory 1234
```
  * will create and run an instance of myImage on the provider environment myImage is compiled for
  * instance will be named newInstance
  * instance will attempt to mount unik-managed volume myVol to /mount1
  * instance will attempt to mount unik-managed volume yourVol to /mount2
  * instance will boot with env variable `foo` set to `bar`
  * instance will boot with env variable `another` set to `one`
  * instance will get 1234 MB of memory
  * note that run must take **exactly** one --vol argument for each mount point defined in the image specification

Flags:
  *  `--env value`             (string,repeated) set any number of environment variables for the instance. must be in the format KEY=VALUE (default [])
  *  `--imageName string`      (string,required) image to use
  *  `--instanceName string`   (string,required) name to give the instance. must be unique
  *  `--vol value`             (string,repeated) each --vol flag specifies one volume id and the corresponding mount point to attach to the instance at boot time. volumes must be attached to the instance for each mount point expected by the image. run 'unik image (image_name)' to see the mount points required for the image. specified in the format 'volume_id:mount_point' (default [])
  * `--instanceMemory`      (int, optional) amount of memory (in MB) to assign to the instance. if none is given, the provider default will be used
  * `--no-cleanup`          (bool, optional) tell UniK not to clean up any artifacts from the launch instance process if launching fails. for debugging purposes.

---

#### List available instances
```
unik instances
```
Lists all available unikernel instances across providers.

---

#### Get JSON representation of a specifig instance:
```
unik describe-instance --instance INSTANCE_NAME
```

---

#### Delete an instance
```
unik delete-instance --instance INSTANCE_NAME
```
Use `--force` to force deleting an instance that is powered on

---

#### Power Off an Instance
```
unik stop --instance INSTANCE_NAME
```
Powering off an instance is a necessary step to attach or detach volumes after an instance has been created.

----

#### Power On an Instance
```
unik start --instance INSTANCE_NAME
```

---

#### Retrieve or Follow Instance Logs
```
unik start --instance INSTANCE_NAME
```
Retrieves logs from a running unikernel instance.

Cannot be used on an instance in powered-off state.
Use the `--follow` flag to attach to the instance's stdout
Use `--delete` in combination with `--follow` to force automatic instance
deletion when the HTTP connection to the instance is broken (by client
disconnect). The `--delete` flag is typically intended for use with
orchestration software such as cluster managers which may require
a persistent http connection managed instances.

Example usage:
```
unik logs --instancce myInstance
```
* will return captured stdout from myInstance since boot time

```
unik logs --instance myInstance --follow --delete
```
* will open an http connection between the cli and unik backend which streams stdout from the instance to the client
* when the client disconnects (i.e. with Ctrl+C) unik will automatically power down and terminate the instance

---

##### Create a Volume

```

```
reate a data volume which can be attached to and detached from
unik-managed instances.

Volumes can be created from a directory, which will copy the contents
of the directory onto the voume. Empty volume can also be created.

Volumes will persist after instances are deleted, allowing application data
to be persisted beyond the lifecycle of individual instances.

If specifying a data folder (with --data), specifying a size for the volume is
not necessary. UniK will automatically size the volume to fit the data provided.
A larger volume can be requested with the --size flag.

If no data directory is provided, --size is a required parameter to specify the
desired size for the empty volume to be createad.

Volumes are created for a specific provider, specified with the --provider flag.
Volumes can only be attached to instances of the same provider type.
To see a list of available providers, run 'unik providers'

Volume names must be unique. If a volume exists with the same name, you will be
required to remove the volume with 'unik delete-volume' before the new volume
can be created.

--size parameter uses MB

Example usage:
unik create-volume --name myVolume --data ./myApp/data --provider aws

* will create an EBS-backed AWS volume named myVolume using the data found in ./myApp/src,
* the size will be either 1GB (the default minimum size on AWS) or greater, if the size of the
volume is greater


Another example (empty volume):
unik create-volume -name anotherVolume --size 500 -provider vsphere

* will create a 500mb sparse vmdk file and upload it to the vsphere datastore,
where it can be attached to a vsphere instance

Flags:
*  `--compiler int`      (int,special) size to create volume in MB. optional if --data is provided
*  `--data string`       (string,special) path to data folder. optional if --size is provided
*  `--name string`       (string,required) name to give the unikernel. must be unique
*  `--provider string`   (string,required) name of the target infrastructure to compile for
* `--no-cleanup`         (bool, optional) tell UniK not to clean up any artifacts from the build process if building fails. for debugging purposes.

---

##### List Volumes

```
unik volumes
```
Lists all available unik-managed volumes across providers.

`ATTACHED-INSTANCE` gives the instance ID of the instance a volume
is attached to, if any. Only volumes that have no attachment are
available to be attached to an instance.

---

##### Attach a Volume

```
unik attach-volume --instance INSTANCE_ID --volume VOLUME_ID --mountPoint MOUNT_POINT
```

Attaches a volume to a stopped instance at a specified mount point.
You specify the volume by name or id.

The volume must be attached to an available mount point on the instance.
Mount points are image-specific, and are determined when the image is compiled.

For a list of mount points on the image for this instance, run `unik images`, or
`unik describe image`

If the specified mount point is occupied by another volume, the command will result
in an error

Flags:
  *  `--force`               (bool, optional) force deleting volume in the case that it is running
  *  `--instance string`     (string,required) name or id of instance to attach to. unik accepts a prefix of the name or id
  *  `--mountPoint string`   (string,required) mount path for volume. this should reflect the mappings specified on the image. run 'unik describe-image' to see expected mount points for the image
  *  `--volume string`       (string,required) name or id of volume to attach. unik accepts a prefix of the name or id

---

##### Detach a Volume

```
unik detach-volume --volume VOLUME_ID
```

Detaches a volume to a stopped instance at a specified mount point.
You specify the volume by name or id.

After detaching the volume, the volume can be mounted to another instance.

If the instance is not stopped, detach will result in an error.

Aliases:
detach-volume, detach


Flags:
  * `--volume string`   (string,required) name or id of volume to detach. unik accepts a prefix of the name or id

---

##### Delete a Volume

```
unik delete-volume --volume VOLUME_NAME [--force]
```

* `--force` forces detaching the volume before deletion if it is currently attached.
