using System;
using System.Collections.Generic;
using System.IO;
using System.Linq;
using System.Threading.Tasks;
using Microsoft.AspNetCore;
using Microsoft.AspNetCore.Hosting;
using Microsoft.Extensions.Configuration;
using Microsoft.Extensions.Logging;

namespace Kubeless.CSharpX
{
	public class Program
	{
		public static void Main( string[] args )
		{
			BuildWebHost( args ).Run();
		}

		public static IWebHost BuildWebHost( string[] args )
		{
			var port = int.TryParse( Environment.GetEnvironmentVariable( "FUNC_PORT" ), out int p ) ? p : 8080;

			return WebHost.CreateDefaultBuilder( args )
				.UseStartup<Startup>()
				.UseUrls( $"http://*:{port}" )      // The port used to expose the service can be modified using an environment variable FUNC_PORT
				.Build();
		}
	}
}
