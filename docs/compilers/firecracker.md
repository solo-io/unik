make compilers-firecracker
make localbuild
reset;./unik daemon --debug
./unik build --name myImage --path ./t/ --base firecracker --language go --provider firecracker --force
./unik run --instanceName myInstance --imageName myImage


./unik instances


./unik delete-instance --instance myInstance
./unik delete-image --image myImage