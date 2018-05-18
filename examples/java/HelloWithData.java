package io.kubeless;

import io.kubeless.Event;
import io.kubeless.Context;

public class Foo {
    public String foo(io.kubeless.Event event, io.kubeless.Context context) {
        System.out.println(event.Data);
        return event.Data;
    }
}
