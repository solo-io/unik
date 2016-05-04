# UniK Design

The UniK Daemon consists of 3 major components.
* The **API server**
* **Compilers**
* **Providers**

The **API Server** handles requests from the CLI / any HTTP Client, then determines which is the appropriate **provider** and/or **compiler** to service the request.

When the **API Server** receives a *build* request (`POST /images/:image_name/create`), it calls the specified **compiler** to build the raw image, and then passes the raw image to the specified **provider**, who processes the raw image with the `Stage()` method, turning it into an infrastructure-specific bootable image (e.g. an *Amazon AMI* on AWS)

The provider for all subsequent operations on the image are determined by a reference to the provider the image was built on.

For more on adding providers, see [providers](providers/README.md)

For more on adding compilers, see [compilers](compilers/README.md)
