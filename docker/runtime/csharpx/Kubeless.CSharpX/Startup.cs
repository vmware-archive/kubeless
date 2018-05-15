using System;
using System.Collections.Generic;
using System.Linq;
using System.Threading.Tasks;
using Microsoft.AspNetCore.Builder;
using Microsoft.AspNetCore.Hosting;
using Microsoft.Extensions.Configuration;
using Microsoft.Extensions.DependencyInjection;
using Microsoft.Extensions.Logging;
using Microsoft.Extensions.Options;
using Microsoft.AspNetCore.Http.Internal;

namespace Kubeless.CSharpX
{
	public class Startup
	{
		public Startup( IConfiguration configuration, IHostingEnvironment env )
		{
			Configuration = configuration;

			if( env.IsDevelopment() )
			{
				// defaults for debug
				Environment.SetEnvironmentVariable( "MOD_NAME", "mycode" );
				Environment.SetEnvironmentVariable( "FUNC_HANDLER", "func" );
				Environment.SetEnvironmentVariable( "FUNC_TIMEOUT", "180" );
			}
		}

		public IConfiguration Configuration { get; }

		// This method gets called by the runtime. Use this method to add services to the container.
		public void ConfigureServices( IServiceCollection services )
		{
			services.AddMvc();

			// constructor injection
			services.AddSingleton<Context>( new Context() );
		}

		// This method gets called by the runtime. Use this method to configure the HTTP request pipeline.
		public void Configure( IApplicationBuilder app, IHostingEnvironment env )
		{
			if( env.IsDevelopment() )
			{
				app.UseDeveloperExceptionPage();
			}

			//Its important the rewind us added before UseMvc
			app.Use( next => context => { context.Request.EnableRewind(); return next( context ); } );

			app.UseMvc();
		}
	}
}
