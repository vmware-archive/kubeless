// Copyright (c) 2016-2017 Bitnami

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
import kubeless;

// A service endpoint represents a listener
endpoint http:Listener listener {
    // Read listner port via ENV.
    port:config: getAsInt("FUNC_PORT", default = 8080)
};

@http:ServiceConfig {
    basePath: "/"
}

service<http:Service> controller bind listener {

    @http:ResourceConfig {
        methods: ["GET", "POST", "DELETE", "PATCH"],
        path: "/"
    }
    handler(endpoint caller, http:Request request) {
        kubeless:Context context = {};
        kubeless:Event event = {};

        //Read environment variables and set them to context
        context.function_name = config:getAsString("FUNC_HANDLER");
        context.time_out = config:getAsString("FUNC_TIMEOUT");
        context.runtime = config:getAsString("FUNC_RUNTIME");
        context.memory_limit = config:getAsString("FUNC_MEMORY_LIMIT");

        //Read requests and set them to event
        match request.getPayloadAsString() {
            string payload => {
                event.data = payload;
            }
            error e => {
                io:println("Error in Payload :" + e.message);
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
        response.setTextPayload(response_payload);

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
