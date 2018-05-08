package io.kubeless;

public class Foo {
    public String Foo(io.kubeless.Event event, io.kubeless.Context context) {
        System.out.println(event.Data);
        return event.Data;
    }
}
