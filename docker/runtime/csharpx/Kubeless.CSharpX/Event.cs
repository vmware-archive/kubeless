using System;
using System.Collections.Generic;
using System.IO;
using System.Linq;
using System.Threading.Tasks;
using Microsoft.AspNetCore.Http;
using Microsoft.AspNetCore.Http.Internal;

namespace Kubeless.CSharpX
{
	public class Event
	{
		public string Data { get; }
		public string EventId { get; }
		public string EventType { get; }
		public string EventTime { get; }
		public string EventNamespace { get; }
		public Extensions Extensions { get; }

		public Event( HttpContext context )
		{
			context.Request.EnableRewind();
			Data = new StreamReader( context.Request.Body ).ReadToEnd();
			EventId = context.Request.Headers[ "event-id" ];
			EventType = context.Request.Headers[ "event-type" ];
			EventTime = context.Request.Headers[ "event-time" ];
			EventNamespace = context.Request.Headers[ "event-namespace" ];
			Extensions = new Extensions(
				context.Request,
				context,
				context.Response );
		}

		public override string ToString()
			=> $"{Data} {EventId} {EventType} {EventTime} {EventNamespace} {Extensions}";
	}
}
