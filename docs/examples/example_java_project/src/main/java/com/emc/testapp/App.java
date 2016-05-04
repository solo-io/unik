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
      HttpServer server = HttpServer.create(new InetSocketAddress(8000), 0);
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
