# Getting Started: OSv on Java Edition!

In this tutorial we'll be:
  1. [installing UniK](getting_started_java.md#installing-unik)
  2. [writing a simple HTTP Daemon in Java](getting_started_java.md#write-a-java-http-server-using-maven)
  3. [compiling to a unikernel and launching an instance on Virtualbox](getting_started_java.md#compile-an-image-and-run-on-virtualbox)

### Installing UniK
#### Prerequisites
Ensure that each of the following are installed
- [Docker](http://www.docker.com/) installed and running with at least 4GB available space for building images
- [`make`](https://www.gnu.org/software/make/)
- [Virtualbox](https://www.virtualbox.org/)
- [Maven](https://maven.apache.org/download.cgi)

#### Install, configure, and launch UniK
1. Install UniK
  ```
  $ git clone https://github.com/emc-advanced-dev/unik.git
  $ cd unik
  $ make
  ```
  note: `make` will take quite a few minutes the first time it runs. the UniK `Makefile` is pulling all of the Docker images that bundle UniK's dependencies.

  Then, place the `unik` executable in your `$PATH` to make running UniK commands easier:
  ```
  $ mv _build/unik /usr/local/bin/
  ```

  2. Configure a Host-Only Network on Virtualbox
    * Open Virtualbox
    * Open **Preferences** > **Network** > **Host-only Networks**
    * Click the green add button on the right side of the UI
    * Record the name of the new Host-Only adapter. You will need this in your UniK configuration
    * Ensure that the Virtualbox DHCP Server is Enabled for this Host-Only Network:
      * With the Host-Only Network selected, Click the edit button (screwdriver image)
      * In the **Adapter** tab, note the IPv4 address and netmask of the adapter.
      * In the **DHCP Server** tab, check the **Enable Server** box
      * Set **Server Address** an IP on the same subnet as the Adapter IP. For example, if the adapter IP is `192.168.100.1`, make set the DHCP server IP as `192.168.100.X`, where X is a number between 2-254.
      * Set **Server Mask** to the netmask you just noted
      * Set **Upper / Lower Address Bound** to a range of IPs on the same subnet. We recommend using the range `X-254` where X is one higher than the IP you used for the DHCP server itself. E.g., if your DHCP server is `192.168.100.2`, you can set the lower and upper bounds to `192.168.100.3` and `192.168.100.254`, respectively.

  3. Configure UniK daemon
    * Using a text editor, create and save the following to `$HOME/.unik/daemon-config.yaml`:
    ```yaml
    providers:
      virtualbox:
        - name: my-vbox
          adapter_type: host_only
          adapter_name: NEW_HOST_ONLY_ADAPTER
    ```
    replacing `NEW_HOST_ONLY_ADAPTER` with the name of the network adapter you created.

    4. Launch UniK and automatically deploy the *Virtualbox Instance Listener*
      * Open a new terminal window/tab. This terminal will be where we leave the UniK daemon running.
      * `cd` to the `_build` directory created by `make`
      * run `./unik daemon --debug` (the `--debug` flag is optional, if you want to see more verbose output)
      * UniK will compile and deploy its own 30 MB unikernel. This unikernel is the [Unik Instance Listener](./instance_listener.md). The instance listener uses udp broadcast to detect instance ips and bootstrap instances running on Virtualbox.
      * After this is finished, UniK is running and ready to accept commands.
      * Open a new terminal window and type `unik target --host localhost` to set the CLI target to the your local machine.

---

#### Write a Java HTTP server using Maven
0. Open a new terminal window, but leave the window with the daemon running. This window will be used for running UniK CLI commands.
1. `mkdir` a new directory to create this sample app in & `cd` to it.
2. Generate a new maven project with the following command:

  ```
  mvn -B archetype:generate \
    -DarchetypeGroupId=org.apache.maven.archetypes \
    -DgroupId=com.mycompany.app \
    -DartifactId=my-app
  ```

    Great! we've got the project structure created. Let's `cd` into the new project folder `my-app`.

3. We need to add a plugin to our project's `pom.xml` so it can be built as a fat jar (all dependencies packaged into one `.jar` file):
  * Add the `maven-assembly-plugin` between the `<plugins>...</plugins>` tags:
    ```xml
<plugins>
       <plugin>
          <groupId>org.apache.maven.plugins</groupId>
          <artifactId>maven-assembly-plugin</artifactId>
          <version>2.2-beta-4</version>
          <configuration>
            <descriptorRefs>
              <descriptorRef>jar-with-dependencies</descriptorRef>
            </descriptorRefs>
            <archive>
              <manifest>
                <mainClass>com.mycompany.app.App</mainClass>
              </manifest>
            </archive>
          </configuration>
          <executions>
            <execution>
              <phase>package</phase>
              <goals>
                <goal>single</goal>
              </goals>
            </execution>
          </executions>
       </plugin>
<plugins>
    ```

    * Now our application is UniK-ready. Let's add some code to our `App.java` source file. Open up `src/main/java/com/mycompany/app/App.java` and replace its contents with the following:

      ```java
      package com.mycompany.app;

      import java.io.IOException;
      import java.io.OutputStream;
      import java.net.InetSocketAddress;

      import com.sun.net.httpserver.HttpExchange;
      import com.sun.net.httpserver.HttpHandler;
      import com.sun.net.httpserver.HttpServer;

      public class App
      {
        public static void main(String[] args) throws Exception {
            System.out.println("Started!");
            HttpServer server = HttpServer.create(new InetSocketAddress(4000), 0);
            server.createContext("/", new MyHandler());
            server.setExecutor(null); // creates a default executor
            server.start();
        }

        static class MyHandler implements HttpHandler {
            @Override
            public void handle(HttpExchange t) throws IOException {
                String response = "Java running inside a unikernel!";
                t.sendResponseHeaders(200, response.length());
                OutputStream os = t.getResponseBody();
                os.write(response.getBytes());
                os.close();
            }
        }
      }     
      ```

2. If you have Java installed, you can try running this code with `mvn package && java -jar target/my-app-1.0-SNAPSHOT-jar-with-dependencies.jar`. Visit [http://localhost:4000/](http://localhost:4000/) to see that the server is running.

3. We have to add a manifest file to tell unik how to build our application into a unikernel. Create a file named `manifest.yaml` in the same directory as the `pom.xml` (the java project root) and paste the following inside:
  ```yaml
artifact_filename: target/my-app-1.0-SNAPSHOT-jar-with-dependencies.jar
build_command: mvn package
  ```
  This will tell UniK to build our project with the `mvn package` commmand, and that the resulting jar file will be located at `target/my-app-1.0-SNAPSHOT-jar-with-dependencies.jar`.

4. Great! Now we're ready to compile this code to a unikernel.

---

#### Compile an image and run on Virtualbox

1. Make sure that the UniK daemon is still running, and run the following:
  ```
  unik build --name myJavaImage --path PATH_TO_JAVA_PROJECT --compiler osv-java-virtaulbox --provider virtaulbox
  ```
  Replacing `PATH_TO_JAVA_PROJECT` with the path to the root of the java project we created. (This will be the directory containing the `pom.xml` file).
2. You can watch the output of the `build` command in the terminal window running the daemon.
3. When `build` finishes, the resulting disk image will exist as a virtual disk image in your `$HOME/.unik` directory.
4. Run an instance of this image with
  ```
  unik run --instanceName myJavaInstance --imageName myJavaImage
  ```
5. When the instance finishes launching, let's check its IP and see that it is running our application.
6. Run `unik instances`. The instance IP Address should be listed.
7. Direct your browser to `http://instance-ip:4000` and see that your instance is running!
8. To clean up your image and the instance you created
  ```
  unik rmi --force --image myImage
  ```
