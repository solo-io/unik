# Example C/C++ application for OSv (MySQL)

This example demonstrates use of pre-compiled applications\* that are available to OSv compiler.
To use these applications (e.g. MySQL) only configuration files are needed since the actual
application is automatically downloaded from public repository.

\* * You can, naturally, also compile your own C/C++ application and use it instead of
pre-compiled ones. Make sure that you compile it with Ubuntu14 compile tools.
into PIC (position independent code) shared object. *

## Configuration files

### manifest.yaml
```
image_size: 20GB
```
This file tells UniK what logical filesystem size would we like to use for the unikernel.

### meta/package.yaml
```
name: my.test
title: My Test
author: I am
require:
 - eu.mikelangelo-project.app.mysql-5.6.21
```
This file tells UniK what pre-comiled packages to download and include into the unikernel.
It also provides some additional information about our unikenrel that is not very relevant
for the build process, but is required by underlying Capstan compiler.

### meta/run.yaml
```
config_set:
   conf1:
      bootcmd: /usr/bin/mysqld --datadir=/usr/data --user=root --init-file=/etc/mysql-init.txt

config_set_default: conf1
```
This file specifies boot command to boot unikernel with. A single configuration set
`conf1` (that is also set as default) is specified. It tells UniK what executable to run
(/usr/bin/mysqld) and what arguments to use (--datadir=/usr/data --user=root --init-file=/etc/mysql-init.txt)

## Build command
```
$ unik build \
   --name osv-mysql-example \
   --path ./                \
   --base osv \
   --language native \
   --provider [qemu|openstack]
```
NOTE: Unikernel with MySQL application needs at least 1GB of memory to boot properly.
