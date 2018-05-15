using System;
using System.Collections.Generic;
using System.IO;
using System.Linq;
using System.Threading;
using System.Threading.Tasks;
using Microsoft.AspNetCore.Mvc;
using Microsoft.CodeAnalysis.Scripting;
using Microsoft.CodeAnalysis.CSharp.Scripting;

namespace Kubeless.CSharpX
{
	[Route( "/" )]
	public class RuntimeController : Controller
	{
		private Context _context;

		public RuntimeController( Context context )
		{
			_context = context;
		}

		[HttpGet]
		[HttpPost]
		public string Get()
		{
			Console.WriteLine( $"{DateTime.Now}: Request: {Request.Method} {Request.Path}" );

			var globals = new Globals( new Event( HttpContext ), _context );

			try
			{
				// The function to load can be specified using an environment variable MOD_NAME
				string modulePath = $"/kubeless/{_context.ModuleName}.csx";
				string funcCode = System.IO.File.ReadAllText( modulePath );    // throw exception if not exist

				// The function to load can be specified using an environment variable FUNC_HANDLER.
				// Functions should receive two parameters: event and context and should return the value that will be used as HTTP response.
				string code =
					$"{funcCode}\n" +
					$"return {_context.FunctionName}( KubelessEvent, KubelessContext );\n";

				// Functions should run FUNC_TIMEOUT as maximum.
				var cancelTokenSource = new CancellationTokenSource( TimeSpan.FromSeconds( _context.Timeout ) );
				var cancelToken = cancelTokenSource.Token;

				// evaluate function
				Console.WriteLine( $"{DateTime.Now}: Evaluate: MOD_NAME:{_context.ModuleName}, FUNC_HANDLER:{_context.FunctionName}" );

				var evalFunction = CSharpScript.EvaluateAsync<string>(
					code,
					ScriptOptions.Default
						.WithReferences( new[] {
						typeof( Event ).Assembly,
						typeof( Context ).Assembly,
						} ),
					globals,
					typeof( Globals ) );

				evalFunction.Wait( cancelToken );

				string result = evalFunction.Result;

				Response.StatusCode = 200;
				Console.WriteLine( $"{DateTime.Now}: Response: {Response.StatusCode}" );

				return result;
			}
			catch( OperationCanceledException )
			{
				Response.StatusCode = 408;
				Console.WriteLine( $"{DateTime.Now}: Response: {Response.StatusCode} - Timeout" );
				return "TimeOut";
			}
			catch( Exception ex )
			{
				// Exceptions in the function should be caught. The server should not exit due to a function error.
				Response.StatusCode = 500;
				Console.WriteLine( $"{DateTime.Now}: Response: {Response.StatusCode} - {ex.Message}" );
				return $"Internal Server Error: {ex.Message}";
			}
		}

		[HttpGet( "/healthz" )]
		public void Health()
		{
			Console.WriteLine( $"{DateTime.Now}: Request: {Request.Method} {Request.Path}" );

			Response.StatusCode = 200;  // The server should return 200 - OK	 to requests at /healthz.

			Console.WriteLine( $"{DateTime.Now}: Response: {Response.StatusCode}" );
		}
	}
}
