package com.emc.wrapper;


import com.google.gson.Gson;
import com.google.gson.reflect.TypeToken;
import com.sun.jna.Library;
import com.sun.jna.Native;
import com.sun.net.httpserver.HttpExchange;
import com.sun.net.httpserver.HttpHandler;
import com.sun.net.httpserver.HttpServer;

import java.io.*;
import java.lang.reflect.Type;
import java.net.*;
import java.util.Enumeration;
import java.util.Map;

public class Bootstrap {
    public static ByteArrayOutputStream logBuffer = new ByteArrayOutputStream();
    public static void bootstrap() {
        //connect stdout to logs
        MultiOutputStream multiOut = new MultiOutputStream(System.out, logBuffer);
        MultiOutputStream multiErr = new MultiOutputStream(System.err, logBuffer);

        PrintStream stdout = new PrintStream(multiOut);
        PrintStream stderr = new PrintStream(multiErr);

        System.setOut(stdout);
        System.setErr(stderr);

        //listen to requests for logs
        try {
            HttpServer server = HttpServer.create(new InetSocketAddress(9876), 0);
            server.createContext("/logs", new ServeLogs());
            server.setExecutor(null); // creates a default executor
            server.start();
        } catch (IOException ex) {
            ex.printStackTrace();
            System.out.println("starting logs server failed, exiting...");
            try {
                Thread.sleep(15000);
            } catch (Exception e) {
                //ignore
            }
            System.exit(-1);
        }

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

    private static String getHTTP(String urlToRead) throws IOException {
        System.out.printf("url: %s\n", urlToRead);
        StringBuilder result = new StringBuilder();
        URL url = new URL(urlToRead);
        HttpURLConnection conn = (HttpURLConnection) url.openConnection();
        conn.setRequestMethod("GET");
        BufferedReader rd = new BufferedReader(new InputStreamReader(conn.getInputStream()));
        String line;
        while ((line = rd.readLine()) != null) {
            result.append(line);
        }
        rd.close();
        return result.toString();
    }

    private static void setEnv(Map<String, String> env) {
        LibC libc = (LibC) Native.loadLibrary("c", LibC.class);
        for (String key : env.keySet()) {
            String value = env.get(key);
            int result = libc.setenv(key, value, 1);
            System.out.println("set " + key + "=" + value + ": " + result);
        }
    }


    private interface LibC extends Library {
        int setenv(String name, String value, int overwrite);
    }

    private static class ServeLogs implements HttpHandler {
        @Override
        public void handle(HttpExchange t) throws IOException {
            byte[] bytes = Bootstrap.logBuffer.toByteArray();
            System.out.println("Response length: "+bytes.length);
            OutputStream os = t.getResponseBody();
            t.sendResponseHeaders(200, bytes.length);
            os.write(bytes);
            os.close();
        }
    }
}
