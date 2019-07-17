# API Description

The Unik daemon provides a RESTful interface. 

* [Compilers](#Compilers)
* [Providers](#Providers)
* [Images](#Images)
* [Instances](#Instances)
* [Volumes](#Volumes)

## Compilers

### GET `/available_compilers`

Get all compilers available to the targeted unik backend.

Returns a list with the names of all compilers. 


### GET `/describe_compiler`

Describes a compiler identified by provider, base and language.

The fallowing query parameters are available to specify the compiler:

Parameter | Description | Mandatory
----------| ----------- | :--------:
provider  |  The (cloud/hypervisor) provider the image should be built for. Supported providers depend on the unikernel base.     | yes
base      | The Unikernel Base to build the image on. These include different unikernel implementations, such as rumprun, IncludeOS, etc.          | yes
lang      | The language/runtime the image should be built with. E.g. for a golang project, specify `go`. Languages supported depend on the unikernel base.        | yes

A string describing the compiler is returned, or `<missing compiler description>` if no description is available. 

## Providers

### GET `/available_providers`

Get all providers available to the targeted unik backend.

Retruns a list with the names of all providers. 


## Images

### GET `/images`

Lists all available unikernel images across providers. Includes important information for running and managing instances, including the required mount points for the image.

Returns a JSON list of objects, describing all images. 
See description of the request for a single image for more [details](#GET-/images/{image_name}) about the object.

### GET `/images/{image_name}`

Describe one image identified by its name.

A JSON of the following format is returned, describing the image: 

```javascript
{
    "Id": string,
    "Name": string,
    "SizeMb": number,
    "Infrastructure": string,
    "Created": string,
    "StageSpec": {
        "ImageFormat": string,
        "XenVirtualizationType": string
    },
    "RunSpec": {
        "DeviceMappings": [
            {
                "MountPoint": string,
                "DeviceName": string
            },
            ...
        ],
        "DefaultInstanceMemory": number,
        "MinInstanceDiskMB": number,
        "StorageDriver": string,
        "VsphereNetworkType": string,
        "Compiler": string
    }
}
```

### POST `/images/{image_name}/create`

Builds  new image on the targeted unik backend, with the specified name.

The body must be of the type `multipart/form-data` with the fallowing values:

Parameter | Description | Mandatory
----------| ----------- | :--------:
tarfile   | A file of the type tar.gz containing the application source. | yes
provider  |  The (cloud/hypervisor) provider the image should be built for. Supported providers depend on the unikernel base.     | yes
base      | The Unikernel Base to build the image on. These include different unikernel implementations, such as rumprun, IncludeOS, etc.          | yes
lang      | The language/runtime the image should be built with. E.g. for a golang project, specify `go`. Languages supported depend on the unikernel base.        | yes
force   | If set to true, the image creation is inforced | no
no_cleanup   | If set to true, the files send to the server will not be removed after the image is created | no
mounts   | A comma separeted list of mountpoint where on runtime a volume is expected to be mounted | no
args   | A list of arguments to be passed to the application at runtime  | no

Returns a JSON object, describing the newly created image.
See description of the request for a single image for more [details](#GET-/images/{image_name}) about the object.


### DELETE `/images/{image_name}`

Deletes an image identified by its name. 

The fallowing query parameters are available:

Parameter | Description | Mandatory
----------| ----------- | :--------:
force   | If set to true, the image removal is inforced | no

Returns HTTP Status `204` on success. 

### POST `/images/push/{image_name}`

Pushes an image from the server to the specified hub. 

The body must be of the type `application/json` with the fallowing format:

```javascript
{
	"url":string,
	"user":string,
	"pass":string
}
```
It describes the Hub configuration, which is the url and credentials for a AWS S3 instance. 

Returns HTTP Status `202` on success. 

### POST `/images/pull/{image_name}`

Pulls an image from the specified hub to the server.

The fallowing query parameters are available:

Parameter | Description | Mandatory
----------| ----------- | :--------:
provider  |  The (cloud/hypervisor) provider the image should be built for. Supported providers depend on the unikernel base.     | yes
force     | If set to true, a locally exsting image might get overwritten | no

The body must contain the description of the Hub, for more details [see](#POST-/images/push/{image_name}).

Returns HTTP Status `202` on success. 


### POST `/images/remote-delete/{image_name}`

Deletes an image from the specified hub. 

The body must contain the description of the Hub, for more details [see](#POST-/images/push/{image_name}).

Returns HTTP Status `202` on success. 
    
## Instances

### GET `/instances`

Lists all available unikernel instances across providers. Includes important information about the state of the instances.

Returns a JSON list of objects, describing all instances. 
See description of the request for a single instance for more [details](#GET-/instances/{instance_id}) about the object.

### GET `/instances/{instance_id}`

Describe an instance identified by its ID or name.

A JSON of the following format is returned, describing the instance: 

```javascript
{
    "Id":string,
    "Name":string,
    "State":string,
    "IpAddress":string,
    "ImageId":string,
    "Infrastructure":string,
    "Created":string
}
```

### DELETE `/instances/{instance_id}`

Deletes an instance identified by its ID or name. 

The fallowing query parameters are available:

Parameter | Description | Mandatory
----------| ----------- | :--------:
force   | If set to true, the instance removal is inforced | no

Returns HTTP Status `204` on success. 

### GET `/instances/{instance_id}/logs`

Retrieves logs from a running unikernel instance. Cannot be used on an instance in powered-off state. 

The fallowing query parameters are available:

Parameter | Description | Mandatory
----------| ----------- | :--------:
fallow   | If set to true, the logs will be continously streamed to the client | no
delete   | Only valid with fallow=true. If set to true, the instance will be removed as soon as the connection is terminated  | no

Returns a string containing all available logs, or a stream if fallow=true. 

### POST `/instances/run`

Deploys a running instance from a unik-compiled unikernel disk image. The instance will be deployed on the provider the image was compiled for. e.g. if the image was compiled for virtualbox, unik will attempt to deploy the image on the configured virtualbox environment.

As body a JSON of the fallowing format is expected:

```javascript
{
	"InstanceName": string,
	"ImageName": string,
	"Mounts": {
        string:string,
        string:string,
        ...
    },
	"Env": {
        string:string,
        string:string,
        ...
    },
	"MemoryMb": number,
	"NoCleanup": boolean,
	"DebugMode": boolean
}
```
Returns a JSON object describing the newly created instance.
See description of the request for a single instance for more [details](#GET-/instances/{instance_id}) about the object.

### POST `/instances/{instance_id}/start`

Powers on an existing instance identified by its name.

Returns HTTP Status `200` on success. 

### POST `/instances/{instance_id}/stop`

Powers of an instance identified by its name. 

Returns HTTP Status `200` on success. 

## Volumes

### GET `/volumes`

Lists all available unik-managed volumes across providers.

Returns a JSON list of objects, describing all volumes. 
See description of the request for a single volume for more [details](#GET-/volumes/{volume_name}) about the object.

### GET `/volumes/{volume_name}`

Describes a volume identified by its name. 

A JSON of the following format is returned, describing the volume: 

```javascript
{
    "Id":string,
    "Name":string,
    "SizeMb":number,
    "Attachment":string,  //instanceId
    "Infrastructure":string,
    "Created":string
}
```

### POST `/volumes/{volume_name}`

Creates a data volume which can be attached to and detached from unik-managed instances.

Every request must contain at least the fallowing form value in its body:

Parameter | Description | Mandatory
----------| ----------- | :--------:
type   | A docker image used for the creation of the volume (?)  | true

If data for the volume is already send with the request, the body must be of the type `multipart/form-data` and must contain the rest of the parameters as well. 
If no data is send with the request, the body can be simple url encoded and the rest of the parameters must be send as query parameters.

Parameter | Description | In Body  | In Query |
----------| ----------- | -------- | -------- |
tarfile   | The data added initially to the volume  | yes | no
provider  |  The (cloud/hypervisor) provider the volume should be built for.   | yes | yes
raw | If ture, the data is send as raw bytes, otherwise as tar.gz | optional | no
no_cleanup |  If set to true, the files send to the server will not be removed after the volume is created | optional | optional 
size | The size of the volume in MB | no | yes 

The size, unlike the other parameters, must always be send as query parameter. 

Returns a JSON describing the newly created volume. 
See description of the request for a single volume for more [details](#GET-/volumes/{volume_name}) about the object.

### DELETE `/volumes/{volume_name}`

Deletes a volume indentified by its name.

The fallowing query parameters are available:

Parameter | Description | Mandatory
----------| ----------- | :--------:
force   | If set to true, the volume removal is inforced | no

Returns HTTP Status `204` on success. 

### POST `/volumes/{volume_name}/attach/{instance_id}`

Attatches an existing volume, identified by its name, to a running instance, identified by its id. 

The fallowing query parameters are available:

Parameter | Description | Mandatory
----------| ----------- | :--------:
mount   | The mount point of the instance for the volume | true

Returns the name of the attached volume. 

### POST `/volumes/{volume_name}/detach`

Detaches a volume from its instance.

Returns the name of the detached volume.





