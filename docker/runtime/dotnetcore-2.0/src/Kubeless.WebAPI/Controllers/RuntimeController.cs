using Kubeless.Core.Interfaces;
using Kubeless.Functions;
using Microsoft.AspNetCore.Mvc;
using Newtonsoft.Json.Linq;
using System;
using System.Threading;

namespace Kubeless.WebAPI.Controllers
{
    [Route("/")]
    public class RuntimeController : Controller
    {
        private readonly IFunction _function;
        private readonly IParameterHandler _parameterManager;
        private readonly IInvoker _invoker;

        public RuntimeController(IFunction function, IParameterHandler parameter, IInvoker invoker)
        {
            _function = function;
            _invoker = invoker;
            _parameterManager = parameter;
        }

        [AcceptVerbs("GET", "POST", "PUT", "PATCH", "DELETE")]
        public object Execute()
        {
            Console.WriteLine("{0}: Function Started. HTTP Method: {1}, Path: {2}.", DateTime.Now.ToString(), Request.Method, Request.Path);

            try
            {
                (Event _event, Context _context) = _parameterManager.GetFunctionParameters(Request);

                CancellationTokenSource _cancellationSource = new CancellationTokenSource();

                var output = _invoker.Execute(_function, _cancellationSource, _event, _context);

                Console.WriteLine("{0}: Function Executed. HTTP response: {1}.", DateTime.Now.ToString(), 200);
                return output;
            }
            catch (OperationCanceledException)
            {
                Console.WriteLine("{0}: Function Cancelled. HTTP Response: {1}. Reason: {2}.", DateTime.Now.ToString(), 408, "Timeout");
                return new StatusCodeResult(408);
            }
            catch (Exception ex)
            {
                Console.WriteLine("{0}: Function Corrupted. HTTP Response: {1}. Reason: {2}.", DateTime.Now.ToString(), 500, ex.Message);
                throw;
            }
        }

        [HttpGet("/healthz")]
        public IActionResult Health() => Ok();

    }
}