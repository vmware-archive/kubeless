using System;

public class mycode
{
    public int execute(object request)
    {
        var result = DoSomeMath(6, 7);
        return result;
    }

    public int DoSomeMath(int x, int y) => x * y;
}
