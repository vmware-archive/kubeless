using System;
using Microsoft.AspNetCore.Http;

public class helloget
{
    public string foo(HttpRequest request)
    {
        return "hello world";
    }
}
