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

        //instance listener bootstrap
        Thread udpListenThread = new Thread(new Runnable() {
            @Override
            public void run() {
                System.out.printf("attempting to bootstrap with udp brodacst..\n");
                try {
                    String listenerIp = getListenerIp(); //needs to be closed
                    Map<String, String> env = registerWithListener(listenerIp);
//                    setEnv(env);
                    System.out.println("udp bootstrap successful.");
                } catch (Exception ex) {
                    System.out.printf("udp bootstrap failed: "+ex.toString());
                }
            }
        });

//        //bootstrap from ec2 metadata
//        Thread ec2BootstrapThread = new Thread(new Runnable() {
//            @Override
//            public void run() {
//                System.out.printf("attempting to bootstrap with ec2 metadata..");
//                try {
//                    Map<String, String> env = getEnvEc2();
////                    setEnv(env);
//                    envSet.value = true;
//                    System.out.println("ec2 bootstrap successful.");
//                } catch (IOException ex) {
//                    System.out.printf("ec2 bootstrap failed: "+ex.toString());
//                }
//            }
//        });

        udpListenThread.setDaemon(true);
//        ec2BootstrapThread.setDaemon(true);

        udpListenThread.start();
//        ec2BootstrapThread.start();

        System.out.println("waiting for env to be set");
        try {
            udpListenThread.join();
        } catch (InterruptedException ex) {

            System.out.println("failed to wait for udp thread to complete!");
        }

//        System.out.println(ec2BootstrapThread.isAlive());
        System.out.println(udpListenThread.isAlive());

        System.out.printf("calling main\n");
    }

    private static Map<String, String> getEnvEc2() throws IOException {
        String resp = getHTTP("http://169.254.169.254/latest/user-data");
        Gson gson = new Gson();
        Type stringStringMap = new TypeToken<Map<String, String>>() {
        }.getType();
        return gson.fromJson(resp, stringStringMap);
    }

    private static String getListenerIp() throws IOException {
        System.out.println("listening for udp heartbeat...");
        DatagramSocket serverSocket = new DatagramSocket(9876);
        byte[] receiveData = new byte[1024];
        while (true) {
            System.out.println("creating datagram receive packet...");
            DatagramPacket receivePacket = new DatagramPacket(receiveData, receiveData.length);
            System.out.println("trying to receive packet...");
            serverSocket.receive(receivePacket);
            System.out.println("converting bytes to string");
            String unikMessage = new String(receivePacket.getData());
            System.out.println("reading source ip...");
            InetAddress IPAddress = receivePacket.getAddress();
            System.out.println("RECEIVED: " + unikMessage + " FROM " + IPAddress.getHostName());
            if (unikMessage.contains("unik")) {
                unikMessage = unikMessage.replaceAll("\\x00", "").trim();
                String[] parts = unikMessage.split(":");
                return parts[1];
            }
        }
    }

    private static Map<String, String> registerWithListener(String listenerIp) throws IOException, MacAddressNotFoundException {
        String macAddress = getMacAddress();
        String resp = postHTTP("http://" + listenerIp + ":3000/register?mac_address=" + macAddress);
        Gson gson = new Gson();
        Type stringStringMap = new TypeToken<Map<String, String>>() {
        }.getType();
        return gson.fromJson(resp, stringStringMap);
    }


    private static String getMacAddress() throws UnknownHostException, SocketException, MacAddressNotFoundException {
        byte[] mac = new byte[1];
        Enumeration<NetworkInterface> ifaces = NetworkInterface.getNetworkInterfaces();
        while (ifaces.hasMoreElements()) {
            NetworkInterface network = ifaces.nextElement();
            System.out.println("Interface name: " + network.getName());
            if (!network.getDisplayName().contains("eth0")) {
                System.out.println("Seeking first network card, skipping interface " + network.getName());
                continue;
            }
            if (network.getHardwareAddress() != null) {
                mac = network.getHardwareAddress();
                break;
            }
        }
        if (mac.length == 1) {
            throw new MacAddressNotFoundException();
        }

        System.out.print("Current MAC address : " + new String(mac));

        StringBuilder sb = new StringBuilder();
        for (int i = 0; i < mac.length; i++) {
            String macString = String.format("%02X%s", mac[i], (i < mac.length - 1) ? ":" : "");
            sb.append(macString.toLowerCase());
        }
        System.out.println(sb.toString());
        return sb.toString();
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

    private static String postHTTP(String urlToRead) throws IOException {
        System.out.printf("url: %s\n", urlToRead);
        StringBuilder result = new StringBuilder();
        URL url = new URL(urlToRead);
        HttpURLConnection conn = (HttpURLConnection) url.openConnection();
        conn.setRequestMethod("POST");
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

    private static class MacAddressNotFoundException extends Exception {}

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
