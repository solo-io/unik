package com.emc.wrapper;

import java.io.File;
import java.io.FileNotFoundException;
import java.io.IOException;
import java.lang.reflect.Method;
import java.util.jar.JarFile;
import java.util.jar.Manifest;
import java.util.zip.ZipException;

public class Wrapper {
    public static void main(String[] args) throws Exception {
        String appArgs[] = new String[1];
        for (String arg : args) {
            if (arg.startsWith("-bootstrapType")) {
                if (arg.contains("ec2")) {
                    new EC2Bootstrap().bootstrap();
                } else {
                    new UDPBootstrap().bootstrap();
                }
            }
            if (arg.startsWith("-appArgs=")) {
                appArgs = arg.replaceFirst("-appArgs", "").split(",,");
            }
            //only one or the other
            if (arg.startsWith("-jarName=")) {
                String jarName = arg.replaceFirst("-jarName=", "");
                String mainClass = getMainClass(jarName);
                Class<?> klass = Thread.currentThread().getContextClassLoader().loadClass(mainClass);
                Method main = klass.getMethod("main", String[].class);
                main.invoke(null, new Object[]{appArgs});
            } else if (arg.startsWith("-tomcat")) {
                System.getProperties().put("java.util.logging.config.file", "/usr/tomcat/conf/logging.properties");
                System.getProperties().put("java.util.logging.manager", "org.apache.juli.ClassLoaderLogManager");
                System.getProperties().put("java.util.logging.manager", "org.apache.juli.ClassLoaderLogManager");
                System.getProperties().put("catalina.base", "/usr/tomcat");
                System.getProperties().put("catalina.home", "/usr/tomcat");
                System.getProperties().put("java.io.tmpdir", "/usr/tomcat/temp");
                Class<?> klass = Thread.currentThread().getContextClassLoader().loadClass("org.apache.catalina.startup.Bootstrap");
                Method main = klass.getMethod("main", String[].class);
                args = new String[1];
                args[0] = "start";
                main.invoke(null, new Object[]{args});
            } else {
                System.err.println("Neet to provide either 'tomcat' or 'jarName' to run!");
                System.out.println("args provided: "+String.join(",", args));
            }
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
}
