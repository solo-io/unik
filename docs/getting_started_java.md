# Getting Started: OSv on Java Edition!

In this tutorial we'll be:
  1. [installing UniK](getting_started_java.md#installing-unik)
  2. [writing a simple HTTP Daemon in Java](getting_started_java.md#write-a-java-http-server-using-maven)
  3. [compiling to a unikernel and launching an instance on AWS](getting_started_java.md#compile-an-image-and-run-on-aws)

### Installing UniK
#### Prerequisites
Ensure that each of the following are installed
- [Docker](http://www.docker.com/) installed and running with at least 4GB available space for building images
- [`make`](https://www.gnu.org/software/make/)
- [Maven](https://maven.apache.org/install.html)

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

2. Configure your AWS credentials in your environment, or in the file `~/.aws/credentials` (see [AWS provder](providers/aws.md))

3. Configure UniK daemon
  * Using a text editor, create and save the following to `$HOME/.unik/daemon-config.yaml`:
  ```yaml
  providers:
    aws:
      - name: up-to-you
        region: AWS_REGION
        zone: AWS_AVAILABILITY_ZONE
  ```
  * replacing `AWS_REGION` with the name of the EC2 region you would like to deploy instances to (e.g.: `us-east-1`)
  * and replacing `AWS_AVAILABILITY_ZONE` wit the name of the EC2 availability zone you would like to deploy instances to. Note that the zone must be within the region you chose (e.g.: `us-east-1a`)

4. Launch UniK!
  * Open a new terminal window/tab. This terminal will be where we leave the UniK daemon running.
  * from any directory, run `unik daemon` (Optional: `unik daemon --debug` for more verbose output)
  * After this is finished, UniK is running and ready to accept commands from the cli.
  * Open a new terminal and type `unik target --host localhost` to set the CLI target to the your local machine.
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

3. Let's add some necessary information to the project's `pom.xml`:
  * First, we'll need to specify some required build plugins. Start by adding the following anywhere between your outermost `<project>...</project>` tags:
    ```xml
<build>
          <plugins>
          <!-- required plugins will go here -->
          </plugins>
</build>
    ```
  * Now let's add a plugin that tells the compiler what version of Java we're building for. Java projects run with UniK should be built for Java 1.7. Add the `maven-compiler-plugin` between the `<plugins>...</plugins>` tags
    ```xml
<plugin>
          <artifactId>maven-compiler-plugin</artifactId>
          <version>2.3.2</version>
          <configuration>
                    <source>1.7</source>
                    <target>1.7</target>
          </configuration>
</plugin>
    ```

  * Now we need to tell the compiler to produce our application as a single `jar` file with all of the dependencies bundled together, with no need for a `libs/` folder. Add the `maven-assembly-plugin` between the `<plugins>...</plugins>` tags:
    ```xml
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

2. If you have Java1.7 or later installed, you can try running this code with `mvn package && java -jar target/my-app-1.0-SNAPSHOT-jar-with-dependencies.jar`. Visit [http://localhost:4000/](http://localhost:4000/) to see that the server is running.
3. Great! Now we're ready to compile this code to a unikernel.

---

#### Compile an image and run on AWS

1. run the following command from the directory where your `pom.xml` is located:
  ```
  unik build --name myJavaImage --path ./ --compiler osv-java-aws --provider aws
  ```
  this command will instruct UniK to compile the sources found in the working directory (`./`) using the `osv-java-aws` compiler, and stage the image for running the `aws` provider.
2. You can watch the output of the `build` command in the terminal window running the daemon.
3. When `build` finishes, the resulting disk image will exist as an AMI on your AWS account. You can see the AMI-ID in the output of `unik images`.
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
