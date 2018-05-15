using System;
using System.Collections.Generic;
using System.Linq;
using System.Threading.Tasks;

namespace Kubeless.CSharpX
{
	public sealed class Globals
	{
		public Event KubelessEvent { get; }
		public Context KubelessContext { get; }

		public Globals( Event ev, Context ctx )
		{
			KubelessEvent = ev;
			KubelessContext = ctx;
		}
	}
}
