using System;
using System.Threading;
using Kubeless.Functions;

public class module
{
    public string handler(Event k8Event, Context k8Context)
    {
        Thread.Sleep(5 * 1000);
        return "This is a long run!";
    }
}
