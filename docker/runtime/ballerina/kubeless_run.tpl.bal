// Copyright (c) 2017-2018 Bitnami

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

//     http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

import ballerina/http;
import ballerina/io;
import ballerina/config;
import func;
import kubeless/kubeless;

// A service endpoint represents a listener
endpoint http:Listener listener {
    //Listner port is 8090 redirected by proxy
    port: 8090
};

@http:ServiceConfig {
    basePath: "/",
    compression: "NEVER"
}

service<http:Service> controller bind listener {
    //Initialize context from environment variables
    kubeless:Context context = {
        function_name: config:getAsString("FUNC_HANDLER"),
        time_out: config:getAsString("FUNC_TIMEOUT"),
        runtime: config:getAsString("FUNC_RUNTIME"),
        memory_limit: config:getAsString("FUNC_MEMORY_LIMIT")
    };

    @http:ResourceConfig {
        methods: ["GET", "POST", "DELETE", "PATCH", "PUT"],
        path: "/"
    }
    handler(endpoint caller, http:Request request) {
        kubeless:Event event = {};
        //Read requests and set them to event
        if (request.hasHeader("Content-Length")){
            int|error result = <int>request.getHeader("Content-Length");
            match result {
                int legnth => {
                    if (legnth > 0){
                        match request.getPayloadAsString() {
                            string payload => {
                                event.data = payload;
                            }
                            error e => {
                                io:println("Error while extracting payload: " + e.message);
                            }
                        }
                    }
                }
                error e => {
                    io:println("Error while reading header: " + e.message);
                }
            }
        }
        //Read headers and set event info.
        if (request.hasHeader("event-id")){
            event.event_id = request.getHeader("event-id");
        }
        if (request.hasHeader("event-type")){
            event.event_type = request.getHeader("event-type");
        }
        if (request.hasHeader("event-time")){
            event.event_time = request.getHeader("event-time");
        }
        if (request.hasHeader("event-namespace")){
            event.event_namespace = request.getHeader("event-namespace");
        }
        // Create object to carry data back to caller
        http:Response response = new;

        // Invoke function and set the response payload.
        string response_payload;
        match func:<<FUNCTION>>(event, context) {
            string return_value => {
                response_payload = return_value;
            }
            error e => {
                response_payload = e.message;
            }
        }
        response.setPayload(untaint response_payload);
        // Send a response back to caller
        // Errors are ignored with '_'
        // -> indicates a synchronous network-bound call
        _ = caller->respond(response);
    }

    // Health check resource
    @http:ResourceConfig {
        methods: ["GET", "POST"],
        path: "/healthz"
    }
    health(endpoint caller, http:Request request) {

        http:Response response = new;
        response.statusCode = http:OK_200;

        // Send a response back to caller
        // Errors are ignored with '_'
        // -> indicates a synchronous network-bound call
        _ = caller->respond(response);
    }
}
