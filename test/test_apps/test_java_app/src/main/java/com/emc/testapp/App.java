package com.emc.testapp;

import java.io.BufferedReader;
import java.io.FileReader;
import java.io.IOException;
import java.io.OutputStream;
import java.net.InetSocketAddress;

import com.sun.net.httpserver.HttpExchange;
import com.sun.net.httpserver.HttpHandler;
import com.sun.net.httpserver.HttpServer;

public class App 
{
  public static void main(String[] args) throws Exception {
      System.out.println("started!");
      HttpServer server = HttpServer.create(new InetSocketAddress(8080), 0);
      server.createContext("/ping_test", new PingHandler());
      server.createContext("/env_test", new EnvHandler());
      server.createContext("/mount_test", new MountHandler());
      server.setExecutor(null); // creates a default executor
      server.start();
  }

  private static class PingHandler implements HttpHandler {
      @Override
      public void handle(HttpExchange t) throws IOException {
          String response = "{\"message\":\"pong\"}";
          t.sendResponseHeaders(200, response.length());
          OutputStream os = t.getResponseBody();
          os.write(response.getBytes());
          os.close();
      }
  }

  private static class EnvHandler implements HttpHandler {
      @Override
      public void handle(HttpExchange t) throws IOException {
          String val = System.getenv("KEY");
          String response = "{\"message\":\""+val+"\"}";
          t.sendResponseHeaders(200, response.length());
          OutputStream os = t.getResponseBody();
          os.write(response.getBytes());
          os.close();
      }
  }

  private static class MountHandler implements HttpHandler {
      @Override
      public void handle(HttpExchange t) throws IOException {
          try(BufferedReader br = new BufferedReader(new FileReader("/data/data.txt"))) {
              StringBuilder sb = new StringBuilder();
              String line = br.readLine();
              while (line != null) {
                  sb.append(line);
                  sb.append(System.lineSeparator());
                  line = br.readLine();
              }
              String data = sb.toString();
              String response = "{\"message\":\""+data+"\"}";
              t.sendResponseHeaders(200, response.length());
              OutputStream os = t.getResponseBody();
              os.write(response.getBytes());
              os.close();
          }
      }
  }
}
