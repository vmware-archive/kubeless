using Kubeless.Core.Interfaces;
using Kubeless.Functions;
using Microsoft.AspNetCore.Http;
using Newtonsoft.Json;
using System;
using System.Collections.Generic;
using System.IO;
using System.Text;

namespace Kubeless.Core.Handlers
{
    public class DefaultParameterHandler : IParameterHandler
    {
        public (Event, Context) GetFunctionParameters(HttpRequest request)
        {
            var _event = GetEvent(request);
            var _context = GetContext();

            return (_event, _context);
        }

        private Event GetEvent(HttpRequest request)
        {
            if (request.Body.CanSeek)
                request.Body.Position = 0;

            var contentType = request.Headers["content-type"];

            object data = new StreamReader(request.Body).ReadToEnd();
            if (contentType == "application/json" && data.ToString().Length > 0)
                data = JsonConvert.DeserializeObject(data.ToString());

            string eventId = request.Headers["event-id"];
            string eventType = request.Headers["event-type"];
            string eventTime = request.Headers["event-time"];
            string eventNamespace = request.Headers["event-namespace"];

            var extensions = new Extensions(request);

            return new Event(data, eventId, eventType, eventTime, eventNamespace, extensions);
        }

        private Context GetContext()
        {
            var moduleName = Environment.GetEnvironmentVariable("MOD_NAME");
            var functionName = Environment.GetEnvironmentVariable("FUNC_HANDLER");
            var functionPort = Environment.GetEnvironmentVariable("FUNC_PORT");
            var timeout = Environment.GetEnvironmentVariable("FUNC_TIMEOUT");
            var runtime = Environment.GetEnvironmentVariable("FUNC_RUNTIME");
            var memoryLimit = Environment.GetEnvironmentVariable("FUNC_MEMORY_LIMIT");

            return new Context(moduleName, functionName, functionPort, timeout, runtime, memoryLimit);
        }
    }
}
