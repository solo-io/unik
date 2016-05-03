package com.emc.wrapper;

import sinetja.Action;
import sinetja.Request;
import sinetja.Response;
import sinetja.Server;

public class WrapperServer  {
    private static final String LOGS_PATH="/logs";

    public void run() {
        try {
            new Server()
                    .GET(LOGS_PATH, new Action() {
                        public void run(Request request, Response response) throws Exception {
                            try {
                                String logs = Bootstrap.logBuffer.toString();
                                response.respondText(logs);
                            } catch (Exception e) {
                                e.printStackTrace();
                                throw e;
                            }
                        }
                    })
                    .start(9876);
        } catch (NoClassDefFoundError ex) {
            //ex.printStackTrace();
        }
    }

    public static class ServerThread extends Thread {
        private final WrapperServer server;
        public ServerThread(WrapperServer server){
            this.server = server;
        }
        @Override
        public void run(){
            server.run();
        }
    }
}

