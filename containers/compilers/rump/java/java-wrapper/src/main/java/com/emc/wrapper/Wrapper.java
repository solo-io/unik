package com.emc.wrapper;

import java.io.File;
import java.io.FileNotFoundException;
import java.io.IOException;
import java.lang.reflect.Method;
import java.net.URL;
import java.net.URLClassLoader;
import java.util.jar.JarFile;
import java.util.jar.Manifest;
import java.util.zip.ZipException;

public class Wrapper {
    public static void main(String[] args) throws Exception {
        String bootstrapType = System.getenv().get("BOOTSTRAP_TYPE");
        if (bootstrapType == null) {
            System.out.println("Must provide env var BOOTSTRAP_TYPE");
            System.exit(-1);
        }
        if (bootstrapType.contains("no-stub")) {
            System.out.println("skipping stub");
        } else if (bootstrapType.contains("ec2")) {
            new EC2Bootstrap().bootstrap();
        } else {
            new UDPBootstrap().bootstrap();
        }

        String jarName = System.getenv().get("MAIN_FILE");
        if (jarName == null) {
            System.out.println("Must provide env var MAIN_FILE");
            System.exit(-1);
        }
        if (jarName.endsWith(".jar")) {
            //Jar Bootstrap
            String mainClass = getMainClass("/bootpart/"+jarName);

            URLClassLoader loader = (URLClassLoader)ClassLoader.getSystemClassLoader();
            MyClassLoader l = new MyClassLoader(loader.getURLs());
            l.addURL(new URL("file:/bootpart/"+jarName));
            Class<?> c = l.loadClass(mainClass);
            System.out.println("succesfully loaded "+c.getName());

            Method main = c.getMethod("main", String[].class);

            System.out.println("calling "+c.getName()+".main("+args+")");
            main.invoke(null, new Object[]{args});
            System.out.println("main finished");
        } else {
            //Jetty Bootstrap
            System.getProperties().put("java.io.tmpdir", "/bootpart/");
            String port = System.getenv().get("PORT");
            if (port != null) {
                System.out.printf("using custom port %s\n", port);
            } else {
                port = "8080";
            }
            String jettyJar = "/bootpart/jetty/start.jar";
            String mainClass = getMainClass(jettyJar);
            URLClassLoader loader = (URLClassLoader)ClassLoader.getSystemClassLoader();
            MyClassLoader l = new MyClassLoader(loader.getURLs());
            l.addURL(new URL("file:"+jettyJar));
            Class<?> c = l.loadClass(mainClass);
            System.out.println("succesfully loaded "+c.getName());

            Method main = c.getMethod("main", String[].class);
            args = new String[3];
            args[0] = "jetty.home=/bootpart/jetty/";
            args[1] = "jetty.base=/bootpart/jetty/";
            args[2] = "jetty.http.port=" + port;
            System.out.println("calling "+c.getName()+".main("+args+")");
            main.invoke(null, new Object[]{args});
            System.out.println("main finished");
        }
    }

    private static String getMainClass(String jarName) throws IOException {
        String mainClass;
        File jarFile = new File(jarName);
        try {
            JarFile jar = new JarFile(jarFile);
            Manifest mf = jar.getManifest();
            jar.close();
            mainClass = mf.getMainAttributes().getValue("Main-Class");
            if (mainClass == null) {
                throw new IllegalArgumentException("No 'Main-Class' attribute in manifest of " + jarName);
            }
        } catch (FileNotFoundException e) {
            throw new IllegalArgumentException("File not found: " + jarName);
        } catch (ZipException e) {
            throw new IllegalArgumentException("File is not a jar: " + jarName, e);
        }
        return mainClass;
    }

    public static class MyClassLoader extends URLClassLoader{

        /**
         * @param urls, to carryforward the existing classpath.
         */
        public MyClassLoader(URL[] urls) {
            super(urls);
        }

        @Override
        /**
         * add ckasspath to the loader.
         */
        public void addURL(URL url) {
            super.addURL(url);
        }

    }
}
