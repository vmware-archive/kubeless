using System;
using Microsoft.AspNetCore.Http;

public class mycode
{
    public int execute(HttpRequest request)
    {
        var result = DoSomeMath(6, 7);
        return result;
    }

    public int DoSomeMath(int x, int y) => x * y;
}
