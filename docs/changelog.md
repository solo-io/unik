# UniK Changelog

## Tue Aug 2 2016
* Update version of Rump in all Rump containers. Fixes I/O panic running Gorump on AWS

## Thu, Jul 28 2016
* UniK now supports running [IncludeOS](http://www.includeos.org/) Unikernels (for C++). Currently only the Virtualbox provider can run IncludeOS unikernels.

## Fri, Jun 17 2016
* UniK's Containers are now (automatically) versioned by the first 16 characters of their SHA256 checksum.
* Changed OSv / Java support:
  * The OSv/Java compiler in UniK will now build a unikernel from any `.jar` or `.war` file.
  * There are two options for building Java unikernels:
    - directly from a compiled fat `.jar` or `.war` file
    - using UniK to invoke the build on your Java sources (e.g. with `mvn package` or `gradle build`)
  * UniK now requires that a `manifest.yaml` file be present in the root directory of Java projects. See the [OSv Java Compiler Documentation](compilers/osv.md#java) or the [Getting Started with Java (New)](getting_started_java.md) for more information.

---

### Tue, Jun 14 2016
* Added QEMU as a provider.
  * Note that the QEMU provider does not provide bridged networking support. This means that QEMU instances will not be reachable from the host network.
  * The QEMU provider includes support for debugging unikernels with `gdb`. See [the qemu provider docs](./providers/qemu.md) for more information.

---

### Tue, Jun 7 2016
*This update features a merge of* `develop` *into* `master`*, which includes a large number of changes, all pushed as a single bundle of features and fixes. Future merges will be more granular.*

**Major Changes:**
* Added support for building Python unikernels on rumprun using Python 3.5
* Boot Volumes are now mountable on rumprun unikernels
  * This enables UniK unikernels to serve static files (HTML, .js, etc.), as well as make scripting language-based unikernels (Javascript, Python) less memory-consumptive (as files no longer have to be loaded into memory at boot time).
  * Example fileservers available in `docs/examples`
* Testing:
  * An integration test suite based on [`ginkgo`](https://onsi.github.io/ginkgo/) has been added to UniK.
  * Tests are located in the `pkg/client` package, with helper functions and scripts in `test`.
  * To run tests, install `go`, `ginkgo` and `gomega`, and run
    ```
    bash test/scripts/test_ginkgo.sh
    ```
    tests currently run against Virtualbox, and Virtualbox is therefore also required to run tests. Note that the tests assume `host_only` networking using `vboxnet1`. To change the VBox network used for tests, modify the values in [`test/scripts/test_ginkgo.sh`](../test/scripts/test_ginkgo.sh)
* Container versioning:
    * UniK's dockerized dependencies (all `projectunik` containers) now use version tags.
    * The purpose of this is to keep older versions of UniK stable while permitting containers hosted on Docker Hub to be updated.
    * To upgrade containers, you only need to run
    ```
    git checkout master
    git pull
    make
    ```
    which will pull the latest versioned containers and recompile unik to utilize that version.

**Minor Changes:**
* Increased timeout when waiting for instance listener UDP packet
* Properly clean up build artifacts from `unik build`, `unik run`, `unik create-volume`
* Fix formatting of CLI output
* Do not auto-delete instances that do not reply to UDP broadcast before a specified timeout
