# Example NodeJS application for OSv

Simple NodeJS application is provided (server.js) with configuration files that are
needed by UniK.

## Configuration files

### manifest.yaml
```
image_size: 1GB
```
This file tells UniK what logical filesystem size would we like to use for the unikernel.

### meta/run.yaml
```
config_set:
   conf1:
      main: server.js
      env:
         PORT: 3000

config_set_default: conf1
```
This file describes our NodeJS project structure. A single configuration set
`conf1` (that is also set as default) is specified. It tells UniK what JavaScript file
to run on boot (server.js) and what environment variables to set inside unikernel
(PORT=3000).

## Build command

### Prepare application
You must install some NodeJS libraries prior using UniK to build unikernel for you.
Make sure that you have version 4.x of node installed:
```
$ node -v
4.3.0
```
Then install required libraries:
```
$ npm install
```
That's it, application is prepared. Go on, build unikernel for it.

### Build
```
$ unik build \
   --name osv-nodejs-example \
   --path ./ \
   --base osv \
   --language nodejs \
   --provider [qemu|openstack]
```

