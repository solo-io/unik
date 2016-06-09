# UniK Changelog

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
*  Container versioning:
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
