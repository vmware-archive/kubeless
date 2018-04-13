using Kubeless.Core.Interfaces;
using Microsoft.AspNetCore.Mvc;

namespace Kubeless.WebAPI.Controllers
{
    [Route("/")]
    public class RuntimeController : Controller
    {
        private readonly IFunction _function;
        private readonly IInvoker _invoker;

        public RuntimeController(IFunction function, IInvoker invoker)
        {
            _function = function;
            _invoker = invoker;
        }

        [HttpPost]
        public object Post([FromBody]object data)
        {
            if (Request.Body.CanSeek)
                Request.Body.Position = 0;

            return _invoker.Execute(_function, Request);
        }

        [HttpGet]
        public object Get() => _invoker.Execute(_function, Request);

        [HttpGet("/healthz")]
        public IActionResult Health() => Ok();
                
    }
}
