using System;
using Microsoft.AspNetCore.Http;
//using Kubeless.Functions;

public class mycode
{
    public string execute(HttpRequest request)
    {
        return "hello world";
    }
}