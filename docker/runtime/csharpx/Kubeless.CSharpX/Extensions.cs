using System;
using System.Collections.Generic;
using System.Linq;
using System.Threading.Tasks;
using Microsoft.AspNetCore.Http;

namespace Kubeless.CSharpX
{
	public class Extensions
	{
		public HttpRequest Request { get; }
		public HttpContext Context { get; }
		public HttpResponse Response { get; }

		public Extensions( HttpRequest request, HttpContext context, HttpResponse response )
		{
			Request = request;
			Context = context;
			Response = response;
		}
	}
}
