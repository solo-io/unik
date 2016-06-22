package com.emc.wrapper;


import com.google.gson.Gson;
import com.google.gson.reflect.TypeToken;
import com.sun.jna.Library;
import com.sun.jna.Native;

import java.io.*;
import java.lang.reflect.Type;
import java.util.Map;

public class EC2Bootstrap extends Bootstrap {
    public void bootstrap() {
        //connect stdout to logs
        MultiOutputStream multiOut = new MultiOutputStream(System.out, logBuffer);
        MultiOutputStream multiErr = new MultiOutputStream(System.err, logBuffer);

        PrintStream stdout = new PrintStream(multiOut);
        PrintStream stderr = new PrintStream(multiErr);

        System.setOut(stdout);
        System.setErr(stderr);

        //listen to requests for logs
        listenForLogs();

        System.out.printf("unik v0.0 bootstrapping beginning...");

        //bootstrap from ec2 metadata
        Thread ec2BootstrapThread = new Thread(new Runnable() {
            @Override
            public void run() {
                System.out.printf("attempting to bootstrap with ec2 metadata..");
                try {
                    Map<String, String> env = getEnvEc2();
                    setEnv(env);
                    System.out.println("ec2 bootstrap successful.");
                } catch (IOException ex) {
                    System.out.printf("ec2 bootstrap failed: "+ex.toString());
                }
            }
        });

        ec2BootstrapThread.setDaemon(true);
        ec2BootstrapThread.start();

        System.out.println("waiting for env to be set");
        try {
            ec2BootstrapThread.join();
        } catch (InterruptedException ex) {

            System.out.println("failed to wait for ec2 thread to complete!");
        }

        System.out.println(ec2BootstrapThread.isAlive());
        System.out.printf("calling main\n");
    }

    private static Map<String, String> getEnvEc2() throws IOException {
        String resp = getHTTP("http://169.254.169.254/latest/user-data");
        Gson gson = new Gson();
        Type stringStringMap = new TypeToken<Map<String, String>>() {
        }.getType();
        return gson.fromJson(resp, stringStringMap);
    }

    private interface LibC extends Library {
        int setenv(String name, String value, int overwrite);
    }

}
