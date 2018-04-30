using System;
using Kubeless.Functions;

public class helloget
{
    public string foo(Event k8Event, Context k8Context)
    {
        return "hello world";
    }
}
