# IncludeOS Unikernels

UniK uses [IncludeOS](http://www.includeos.org/) as a platform for compiling C++ applications to unikernels. Building IncludeOS unikernels requires conforming to the structure of an IncludeOS project, documented [here](https://github.com/hioa-cs/IncludeOS/wiki/Creating-your-first-IncludeOS-service), with an example  [here](https://github.com/hioa-cs/IncludeOS/tree/master/seed).

You can also see the [example in the examples directory](../examples/example-cpp-includeos)

Your application code will be called from ```Service::start()``` in `service.cpp`

Note the line `unik::register_instance();` in [service.cpp#L41](../examples/example-cpp-includeos/service.cpp#L41): this line (and the imported file `#include <unik>`) are required for registering instances of your application to UniK. Without this, UniK will be unable to IPs of your instances (they will run otherwise normally).
