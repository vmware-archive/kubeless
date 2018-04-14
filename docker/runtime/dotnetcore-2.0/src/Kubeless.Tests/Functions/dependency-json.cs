using System;
using Microsoft.AspNetCore.Http;
using Newtonsoft.Json;

public class module
{
    public string handler(HttpRequest request)
    {
        var person = new Person()
        {
            Name = "Allan",
            Age = 24
        };

        return JsonConvert.SerializeObject(person);
    }
}

public class Person
{
    public string Name { get; set; }
    public int Age { get; set; }
}