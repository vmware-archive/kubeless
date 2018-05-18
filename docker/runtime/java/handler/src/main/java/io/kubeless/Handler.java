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

import com.sun.net.httpserver.HttpExchange;
import com.sun.net.httpserver.HttpHandler;
import com.sun.net.httpserver.HttpServer;
import com.sun.net.httpserver.Headers;

import java.io.IOException;
import java.io.OutputStream;
import java.io.InputStreamReader;
import java.io.BufferedReader;
import java.lang.reflect.Method;
import java.lang.reflect.InvocationTargetException;
import java.net.InetSocketAddress;
import java.nio.charset.StandardCharsets;
import java.util.List;
import java.util.stream.Collectors;
import io.prometheus.client.Counter;
import io.prometheus.client.Histogram;
import io.kubeless.Event;
import io.kubeless.Context;
import org.apache.log4j.Logger;
import org.apache.log4j.BasicConfigurator;

public class Handler {

    static String className   = System.getenv("MOD_NAME");
    static String methodName  = System.getenv("FUNC_HANDLER");
    static String timeout     = System.getenv("FUNC_TIMEOUT");
    static String runtime     = System.getenv("FUNC_RUNTIME");
    static String memoryLimit = System.getenv("FUNC_MEMORY_LIMIT");
    static Method method;
    static Object obj;
    static Logger logger = Logger.getLogger(Handler.class.getName());

    static final Counter requests = Counter.build().name("function_calls_total").help("Total function calls.").register();
    static final Counter failures = Counter.build().name("function_failures_total").help("Total function call failuress.").register();
    static final Histogram requestLatency = Histogram.build().name("function_duration_seconds").help("Duration of time user function ran in seconds.").register();

    public static void main(String[] args) {

        BasicConfigurator.configure();

        String funcPort = System.getenv("FUNC_PORT");
        if(funcPort == null || funcPort.isEmpty()) {
            funcPort = "8080";
        }
        int port = Integer.parseInt(funcPort);
        try {
            HttpServer server = HttpServer.create(new InetSocketAddress(port), 0);
            server.createContext("/", new FunctionHandler());
            server.createContext("/healthz", new HealthHandler());
            server.setExecutor(java.util.concurrent.Executors.newFixedThreadPool(50));
            server.start();

            Class<?> c = Class.forName("io.kubeless."+className);
            obj = c.newInstance();
            method = c.getMethod(methodName, io.kubeless.Event.class, io.kubeless.Context.class);
        } catch (Exception e) {
            failures.inc();
            if (e instanceof ClassNotFoundException) {
                logger.error("Class: " + className + " not found");
            } else if (e instanceof NoSuchMethodException) {
                logger.error("Method: " + methodName + " not found");
            } else if (e instanceof java.io.IOException) {
                logger.error("Failed to starting listener.");
            } else {
                logger.error("An exception occured running Class: " + className + " method: " + methodName);
                e.printStackTrace();
            }
        }
    }

    static class FunctionHandler implements HttpHandler {

        @Override
        public void handle(HttpExchange he) throws IOException {
            Histogram.Timer requestTimer = requestLatency.startTimer();
            try {
                requests.inc();

                InputStreamReader reader = new InputStreamReader(he.getRequestBody(), StandardCharsets.UTF_8.name());
                BufferedReader br = new BufferedReader(reader);
                String requestBody = br.lines().collect(Collectors.joining());
                br.close();
                reader.close();

                Headers headers = he.getRequestHeaders();
                String eventId  = getEventId(headers);
                String eventType = getEventType(headers);
                String eventTime = getEventTime(headers);
                String eventNamespace = getEventNamespace(headers);

                Event event = new Event(requestBody, eventId, eventType, eventTime, eventNamespace);
                Context context = new Context(methodName, timeout, runtime, memoryLimit);

                Object returnValue = Handler.method.invoke(Handler.obj, event, context);
                String response = (String)returnValue;
                logger.info("Response: " + response);
                he.sendResponseHeaders(200, response.length());
                OutputStream os = he.getResponseBody();
                os.write(response.getBytes());
                os.close();
            } catch (Exception e) {
                failures.inc();
                if (e instanceof ClassNotFoundException) {
                    logger.error("Class: " + className + " not found");
                } else if (e instanceof NoSuchMethodException) {
                    logger.error("Method: " + methodName + " not found");
                } else if (e instanceof InvocationTargetException) {
                    logger.error("Failed to Invoke Method: " + methodName);
                } else if (e instanceof InstantiationException) {
                    logger.error("Failed to instantiate method: " + methodName);
                } else {
                    logger.error("An exception occured running Class: " + className + " method: " + methodName);
                    e.printStackTrace();
                }
            } finally {
                requestTimer.observeDuration();
            }
        }

        private String getEventType(Headers headers) {
            if (headers.containsKey("event-type")) {
                List<String> values = headers.get("event-type");
                if (values != null) {
                    return values.get(0);
                }
            }
            return "";
        }

        private String getEventTime(Headers headers) {
            if (headers.containsKey("event-time")) {
                List<String> values = headers.get("event-time");
                if (values != null) {
                    return values.get(0);
                }
            }
            return "";
        }

        private String getEventNamespace(Headers headers) {
            if (headers.containsKey("event-namespace")) {
                List<String> values = headers.get("event-namespace");
                if (values != null) {
                    return values.get(0);
                }
            }
            return "";
        }

        private String getEventId(Headers headers) {
            if (headers.containsKey("event-id")) {
                List<String> values = headers.get("event-id");
                if (values != null) {
                    return values.get(0);
                }
            }
            return "";
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
