using System;
using System.Text;
using System.IO;
using System.Linq;
using Microsoft.AspNetCore.Http;
using Newtonsoft.Json;
using YamlDotNet.Serialization;

public class mycode
{
    public string execute(HttpRequest request)
    {
        var person = new Person(){
            Name="Allan",
            Age=24
        };
        
        var serializer = new SerializerBuilder().Build();
        return serializer.Serialize(person);
        //return JsonConvert.SerializeObject(person);
    }
}

public class Person{
    public string Name { get; set; }
    public int Age { get; set; }
}