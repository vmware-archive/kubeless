using System;
using System.Collections.Generic;
using System.Linq;
using System.Threading.Tasks;

namespace Kubeless.CSharpX
{
	public class Context
	{
		public string ModuleName { get; }
		public string FunctionName { get; }
		public int FunctionPort { get; }
		public int Timeout { get; }
		public string Runtime { get; }
		public string MemoryLimit { get; }

		public Context()
		{
			ModuleName = Environment.GetEnvironmentVariable( "MOD_NAME" );
			FunctionName = Environment.GetEnvironmentVariable( "FUNC_HANDLER" );
			FunctionPort = int.TryParse( Environment.GetEnvironmentVariable( "FUNC_PORT" ), out int p ) ? p : 8080;
			Timeout = int.TryParse( Environment.GetEnvironmentVariable( "FUNC_TIMEOUT" ), out int t ) ? t : 180;
			Runtime = Environment.GetEnvironmentVariable( "FUNC_RUNTIME" );
			MemoryLimit = Environment.GetEnvironmentVariable( "FUNC_MEMORY_LIMIT" );
		}

		public override string ToString()
			=> $"{ModuleName} {FunctionName} {FunctionPort} {Timeout} {Runtime} {MemoryLimit}";
	}
}
