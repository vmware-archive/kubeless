using System;
using Microsoft.AspNetCore.Http;
using YamlDotNet.Serialization;

public class module
{
    public string handler(HttpRequest request)
    {
        var person = new Person()
        {
            Name = "Michael J. Fox",
            Age = 56
        };

        var serializer = new SerializerBuilder().Build();
        return serializer.Serialize(person);
    }
}

public class Person
{
    public string Name { get; set; }
    public int Age { get; set; }
}