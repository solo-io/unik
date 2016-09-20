# Mirage Unikernels

Compile Mirage Ocaml unikernels with unik.

---

Mirage supports only the Xen hypervisor for now (support for Solo5 is planned). So make sure to configure a [Xen provider](../providers/xen.md) for unik (and run the unik daemon on Dom0).

## Build an Image

To compile examples from mirage-skeleton:
```
git clone https://github.com/mirage/mirage-skeleton
unik build --name sw --path ./mirage-skeleton/static_website/  --base mirage --language ocaml --provider xen
```

## Volumes

Unik will automatically detect if the unikernel needs data volumes mounted, and will autogenerate mountpoints. 
Since there are no directory structure, the mount point will indicate the name of the xen device (rather than a path in the file system)

You can see the auto generated mountpoints with `unik images` or `unik describe-instance`.

Example output of `unik images`:
```
NAME                 ID                   INFRASTRUCTURE  CREATED                        SIZE(MB) MOUNTPOINTS
sw                   sw                   XEN             2016-09-20 17:34:40.884854395  35       xen:xvdc
```

In our example, the created mountpoint is named "xen:xvdc"

Now create the volume using unik, and make sure to use the "mirage-fat" type:
```
unik create-volume --name websitedata1 --data ./mirage-skeleton/static_website/htdocs/ --type mirage-fat --provider xen
```

## Run it
To run and image with attached volumes, use the run command:
```
unik run --instanceName sw1 --imageName sw --vol websitedata1:xen:xvdc
```
