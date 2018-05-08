package io.kubeless;

import org.joda.time.LocalTime;

public class Hello {
    public String sayHello(io.kubeless.Event event, io.kubeless.Context context) {
        LocalTime currentTime = new LocalTime();
        return "Hello world! Current local time is: " + currentTime;
    }
}
