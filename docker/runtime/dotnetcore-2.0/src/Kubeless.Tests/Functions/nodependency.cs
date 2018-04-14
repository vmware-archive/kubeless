using System;
using Microsoft.AspNetCore.Http;

public class module
{
    public string handler(HttpRequest request)
    {
        return "hello world";
    }
}