package com.emc.testapp;

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
      System.out.println("value of property1: "+System.getProperty("property1"));
      System.out.println("value of property2: "+System.getProperty("property2"));
      HttpServer server = HttpServer.create(new InetSocketAddress(3000), 0);
      server.createContext("/test", new MyHandler());
      server.setExecutor(null); // creates a default executor
      server.start();
  }

  static class MyHandler implements HttpHandler {
      @Override
      public void handle(HttpExchange t) throws IOException {
          String response = "This is the response";
          t.sendResponseHeaders(200, response.length());
          OutputStream os = t.getResponseBody();
          os.write(response.getBytes());
          os.close();
      }
  }
}
