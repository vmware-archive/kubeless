import kubeless/kubeless;
import ballerina/io;
import ballerina/config;

public function bar(kubeless:Event event, kubeless:Context context) returns (string|error) {
    io:println(event);
    io:println(context);
    return config:getAsString("hello.userid");
}
