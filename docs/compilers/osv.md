# OSv Unikernels

UniK uses OSv as a platform for compiling Java to unikernels.

---

### Java

Compiling Java on the OSv platform requires the following parameters be met:
* One `main` class in your project
* Project compiles to Java version 1.7
* Use of `maven` with a `pom.xml` file in the root project folder
  * built to compile to a `.jar`
  * with the following format in the project configuration:

    ```xml
<project>
      <!--...-->
      <groupId>PROJECT_GROUP_ID</groupId>
      <artifactId>PROJECT_ARTIFACT_ID</artifactId>
      <version>1.0-SNAPSHOT</version>
      <packaging>jar</packaging>
      <!--...-->
</project>
    ```

  * and following plugins confiugred:

    ```xml
<build>
      <plugins>
      <!--...-->
      <plugin>
      	<artifactId>maven-compiler-plugin</artifactId>
      	<version>2.3.2</version>
      	<configuration>
      		<source>1.7</source>
      		<target>1.7</target>
      	</configuration>
      </plugin>
      <!--...-->
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
              <mainClass>PROJECT_GROUP_ID.YOUR_MAIN_CLASS</mainClass>
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
      <!--...-->
      </plugins>
</build>

    ```
