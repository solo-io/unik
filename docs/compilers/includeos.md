# IncludeOS Unikernels

UniK uses [IncludeOS](http://www.includeos.org/) as a platform for compiling C++ applications to unikernels. Building IncludeOS unikernels requires conforming to the structure of an IncludeOS project, [documented here](https://github.com/hioa-cs/IncludeOS/wiki/Creating-your-first-IncludeOS-service), with an [example here](https://github.com/hioa-cs/IncludeOS/tree/master/seed).

You can also see an [example here](https://github.com/includeos/unik_test_service)

Your application code will be called from ```Service::start()``` in `service.cpp`
