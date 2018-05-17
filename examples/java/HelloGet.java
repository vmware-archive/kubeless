package io.kubeless;

import io.kubeless.Event;
import io.kubeless.Context;

public class Foo {
    public String foo(io.kubeless.Event event, io.kubeless.Context context) {
        return "Hello world!";
    }
}
