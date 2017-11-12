using System;
using System.Collections.Generic;
using System.IO;
using System.Linq;
using System.Threading.Tasks;
using Microsoft.AspNetCore;
using Microsoft.AspNetCore.Hosting;
using Microsoft.Extensions.Configuration;
using Microsoft.Extensions.Logging;

namespace kubeless_netcore_runtime
{
    public class Program
    {
        public static void Main(string[] args)
        {
            BuildWebHost(args).Run();
        }

        public static IWebHost BuildWebHost(string[] args) =>
            var envPortStr = Environment.GetEnvironmentVariable("FUNC_PORT");
            var urls = "http://*:8080";
            if (!string.IsNullOrEmpty(envPortStr))
                urls = string.Concat("http://*:", envPortStr)
            WebHost.CreateDefaultBuilder(args)
                .UseStartup<Startup>()
                .UseUrls(urls)
                .Build();
    }
}
