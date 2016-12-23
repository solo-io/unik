UniK has a built-in support to directly run application packages provided in
[MIKELANGELO Project](https://www.mikelangelo-project.eu)'s
[format](https://github.com/mikelangelo-project/capstan/blob/develop/Documentation/ApplicationManagement.md#package-management).

### Dynamic compiler
UniK offers a dynamic OSv compiler to run arbitrary Capstan package.
To use dynamic compiler, you must wrap your binaries into a Capstan package.
Consult [Capstan documentation](https://github.com/mikelangelo-project/capstan/blob/develop/Documentation/ApplicationManagement.md#package-management)
to learn how to prepare a valid Capstan package. To sum it up, you need to perform these two steps:

1. compile your code, if needed (e.g. for NodeJS this step is not needed, but for Java it is)
2. create `meta` folder in your project root, with two files in it (
[package.yaml](https://github.com/mikelangelo-project/capstan/blob/develop/Documentation/ApplicationManagement.md#manual-initialisation-of-a-package) and
[run.yaml](https://github.com/mikelangelo-project/capstan/blob/develop/Documentation/ApplicationManagement.md#runyaml-optional)
)

That's it! Now you can build unikernel with your application using this command:
```
$ unik build
    --name myImage
    --path ./
    --base osv
    --language dynamic # <------ dynamic compiler
    --provider openstack
```
NOTE: Currently, only `openstack` and `qemu` providers are supported by dynamic OSv compiler.
