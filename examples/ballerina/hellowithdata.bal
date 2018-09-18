import kubeless/kubeless;
import ballerina/io;

public function foo(kubeless:Event event, kubeless:Context context) returns (string|error) {
    io:println(event);
    io:println(context);
    return <string>event.data;
}
