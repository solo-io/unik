# OSv Unikernels

UniK uses OSv as a platform for compiling Java to unikernels.

---

### Java

Compiling Java on the OSv platform requires the following parameters be met:
* Project compiles to Java version 1.8 or earlier
* A `manifest.yaml` file in the root directory of the project specifying the following information:
  * An optional build command for unpackaged sources (required if the project is not already packaged as a `.jar` or `.war` file).
  * The name of the project artifact
  * An optional list of properties (normally set with the `-Dproperty=value` in java) to pass to the application
  * See the [example java project](../examples/example_java_project) or the [example java servlet](../examples/example_java_project) for an example.
* Either:
  * Project packaged as a fat `.jar` file or `.war` file *or*
  * Project uses **Gradle** or **Maven** and able to be built as a fat `.jar` or `.war`
