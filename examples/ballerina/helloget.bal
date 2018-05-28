import kubeless/kubeless;

public function foo(kubeless:Event event, kubeless:Context context) returns (string|error) {
    return "Hello World Ballerina";
}
