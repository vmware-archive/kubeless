using System;
using Microsoft.AspNetCore;
using Microsoft.AspNetCore.Hosting;

namespace kubeless_netcore_runtime
{
    public class Program
    {
        public static void Main(string[] args)
        {
            BuildWebHost(args).Run();
        }

        public static IWebHost BuildWebHost(string[] args)
        {
            var envPortStr = Environment.GetEnvironmentVariable("FUNC_PORT");
            var port = string.IsNullOrWhiteSpace(envPortStr) ? "8080" : envPortStr;

            return WebHost.CreateDefaultBuilder(args)
                .UseStartup<Startup>()
                .UseUrls($"http://*:{port}")
                .Build();
        }
    }
}
