using System;
using System.Collections.Generic;
using System.Linq;
using System.Threading;
using System.Threading.Tasks;
using Microsoft.AspNetCore.Builder;
using Microsoft.AspNetCore.Hosting;
using Microsoft.Extensions.Configuration;
using Microsoft.Extensions.DependencyInjection;
using Microsoft.Extensions.Logging;
using Microsoft.Extensions.Options;
using Microsoft.AspNetCore.Http;
using Microsoft.AspNetCore.Http.Internal;
using Microsoft.CodeAnalysis.Scripting;
using Microsoft.CodeAnalysis.CSharp.Scripting;

namespace Kubeless.CSharpX
{
	public class Startup
	{
		public Startup( IConfiguration configuration, IHostingEnvironment env )
		{
		}

		public void ConfigureServices( IServiceCollection services )
		{
		}

		public void Configure( IApplicationBuilder app, IHostingEnvironment env )
		{
			app.Use( next =>
				httpContext => {
					httpContext.Request.EnableRewind();
					return next( httpContext );
				} );

			app.Map( "/healthz", 
				builder => builder.Run( this.GetHealthz ) );

			app.Run( this.Get );
		}

		async Task Get( HttpContext httpContext )
		{
			await Console.Out.WriteLineAsync( $"{DateTime.Now}: Request: {httpContext.Request.Method} {httpContext.Request.Path}" );

			var _globals = new Globals( new Event( httpContext ), new Context() );

			try
			{
				// The function to load can be specified using an environment variable MOD_NAME
				string modulePath = $"/kubeless/{_globals.KubelessContext.ModuleName}.csx";
				string funcCode = await System.IO.File.ReadAllTextAsync( modulePath );    // throw exception if not exist

				// The function to load can be specified using an environment variable FUNC_HANDLER.
				// Functions should receive two parameters: event and context and should return the value that will be used as HTTP response.
				string code =
					$"{funcCode}\n" +
					$"return {_globals.KubelessContext.FunctionName}( KubelessEvent, KubelessContext );\n";

				// Functions should run FUNC_TIMEOUT as maximum.
				var timeoutTokenSource = new CancellationTokenSource( TimeSpan.FromSeconds( _globals.KubelessContext.Timeout ) );
				var requestAbortToken = httpContext.RequestAborted;
				var linkedToken = CancellationTokenSource.CreateLinkedTokenSource( timeoutTokenSource.Token, requestAbortToken );   // Timeout & Aborted

				// evaluate function
				await Console.Out.WriteLineAsync( $"{DateTime.Now}: Evaluate: MOD_NAME:{_globals.KubelessContext.ModuleName}, FUNC_HANDLER:{_globals.KubelessContext.FunctionName}" );

				var scriptResult = await CSharpScript.EvaluateAsync<string>(
					code,
					ScriptOptions.Default
						.WithReferences( new[] {
						typeof( Event ).Assembly,
						typeof( Context ).Assembly,
						} ),
					_globals,
					typeof( Globals ), linkedToken.Token );

				httpContext.Response.StatusCode = 200;
				await httpContext.Response.WriteAsync( scriptResult );

				await Console.Out.WriteLineAsync( $"{DateTime.Now}: Response: {httpContext.Response.StatusCode}" );
			}
			catch( OperationCanceledException )
			{
				httpContext.Response.StatusCode = 408;
				await Console.Out.WriteLineAsync( $"{DateTime.Now}: Response: {httpContext.Response.StatusCode} - Timeout" );
			}
			catch( Exception ex )
			{
				// Exceptions in the function should be caught. The server should not exit due to a function error.
				httpContext.Response.StatusCode = 500;
				await Console.Out.WriteLineAsync( $"{DateTime.Now}: Response: {httpContext.Response.StatusCode} - {ex.Message}" );
				await httpContext.Response.WriteAsync( $"Internal Server Error: {ex.Message}" );
			}
		}

		async Task GetHealthz( HttpContext httpContext )
		{
			await Console.Out.WriteLineAsync( $"{DateTime.Now}: Request: {httpContext.Request.Method} {httpContext.Request.Path}" );

			httpContext.Response.StatusCode = 200;
			await httpContext.Response.WriteAsync( "Ok" );

			await Console.Out.WriteLineAsync( $"{DateTime.Now}: Response: {httpContext.Response.StatusCode}" );
		}
	}
}
