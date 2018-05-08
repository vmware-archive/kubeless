/*
Copyright (c) 2016-2017 Bitnami

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package io.kubeless;

import java.io.IOException;
import java.io.OutputStream;
import java.net.InetSocketAddress;

import com.sun.net.httpserver.HttpExchange;
import com.sun.net.httpserver.HttpHandler;
import com.sun.net.httpserver.HttpServer;

import io.prometheus.client.Counter;
import io.prometheus.client.Histogram;
import java.lang.reflect.InvocationTargetException;
import io.kubeless.Event;
import io.kubeless.Context;

public class Handler {
    public static void main(String[] args) {
        String funcPort = System.getenv("FUNC_PORT");
        if(funcPort == null || funcPort.isEmpty()) {
            funcPort = "8080";
        }
        int port = Integer.parseInt(funcPort);
        try {
            HttpServer server = HttpServer.create(new InetSocketAddress(port), 0);
            server.createContext("/", new FunctionHandler());
            server.createContext("/healthz", new HealthHandler());
            server.setExecutor(null);
            server.start();
        } catch (java.io.IOException e) {
            System.out.println("Failed to starting listener.");
        }
    }

    static class FunctionHandler implements HttpHandler {
        String className = System.getenv("MOD_NAME");
        String methodName = System.getenv("FUNC_HANDLER");
        static final Counter requests = Counter.build().name("function_calls_total").help("Total function calls.").register();
        static final Counter failures = Counter.build().name("function_failures_total").help("Total function call failuress.").register();
        static final Histogram requestLatency = Histogram.build().name("function_duration_seconds").help("Duration of time user function ran in seconds.").register();

        @Override
        public void handle(HttpExchange t) throws IOException {
            Histogram.Timer requestTimer = requestLatency.startTimer();
            try {
                requests.inc();
                Class<?> c = Class.forName("io.kubeless."+className);
                Object obj = c.newInstance();
                java.lang.reflect.Method method = c.getMethod(methodName, io.kubeless.Event.class, io.kubeless.Context.class);

                io.kubeless.Event event = new io.kubeless.Event("", "", "", "", "");
                io.kubeless.Context context = new io.kubeless.Context("", "", "", "");

                Object returnValue = method.invoke(obj, event, context);
                String response = (String)returnValue;
                System.out.println("Response: " + response);
                t.sendResponseHeaders(200, response.length());
                OutputStream os = t.getResponseBody();
                os.write(response.getBytes());
                os.close();
            } catch (Exception e) {
                failures.inc();
                if (e instanceof ClassNotFoundException) {
                    System.out.println("Class: " + className + " not found");
                } else if (e instanceof NoSuchMethodException) {
                    System.out.println("Method: " + methodName + " not found");
                } else if (e instanceof InvocationTargetException) {
                    System.out.println("Failed to Invoke Method: " + methodName);
                } else if (e instanceof InstantiationException) {
                    System.out.println("Failed to instantiate method: " + methodName);
                } else {
                    System.out.println("An exception occured running Class: " + className + " method: " + methodName);
                }
            } finally {
                requestTimer.observeDuration();
            }
        }
    }

    static class HealthHandler implements HttpHandler {
        @Override
        public void handle(HttpExchange t) throws IOException {
            String response = "OK";
            t.sendResponseHeaders(200, response.length());
            OutputStream os = t.getResponseBody();
            os.write(response.getBytes());
            os.close();
        }
    }
}
