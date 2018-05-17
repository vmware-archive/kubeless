using System;
using Kubeless.Functions;

public class helloget
{
    public string handler(Event k8Event, Context k8Context)
    {
        return "hello world";
    }
}
