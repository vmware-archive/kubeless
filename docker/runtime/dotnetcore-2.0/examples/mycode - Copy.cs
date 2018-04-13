using System;
using System.Text;
using System.IO;
using System.Linq;
using Microsoft.AspNetCore.Http;

public class mycode
{
    public string execute(HttpRequest request)
    {
        var builder = new StringBuilder();
        var files = Directory.EnumerateFiles(@"C:\Users\altargin\Desktop\dotnetcore-2.0\src\Kubeless.WebAPI\bin\Debug\netcoreapp2.0", "*", SearchOption.AllDirectories);

        foreach (var file in files)
        {
            builder.AppendLine(file);
        }

        return builder.ToString();

        // return "hello world";
    }
}