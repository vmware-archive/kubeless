using System;
using Kubeless.Functions;

public class module
{
    public object handler(Event k8Event, Context k8Context)
    {
        return k8Event.Data;
    }
}
