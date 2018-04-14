using System;
using System.IO;
using Microsoft.AspNetCore.Http;

public class mycode
{
    public int execute(HttpRequest request)
    {
        var n = int.Parse(new StreamReader(request.Body).ReadToEnd());

        return fibonacci(n);
    }
    
    public int fibonacci(int n)
    {
        if ((n == 0) || (n == 1))
            return n;
        return fibonacci(n - 1) + fibonacci(n - 2);
    }
}