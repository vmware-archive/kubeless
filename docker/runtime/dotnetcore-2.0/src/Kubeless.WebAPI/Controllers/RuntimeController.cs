using Kubeless.Core.Interfaces;
using Microsoft.AspNetCore.Http;
using Microsoft.AspNetCore.Mvc;
using Microsoft.Extensions.Configuration;

namespace Kubeless.WebAPI.Controllers
{
    [Route("/")]
    public class RuntimeController : Controller
    {
        private IFunction _function;
        private ICompiler _compiler;
        private IInvoker _invoker;
        private IConfiguration _configuration;

        public RuntimeController(IFunction function, ICompiler compiler, IInvoker invoker, IConfiguration configuration)
        {
            _function = function;
            _compiler = compiler;
            _invoker = invoker;
            _configuration = configuration;
        }

        [HttpPost]
        public object Post([FromBody]object data)
        {
            if (Request.Body.CanSeek)
                Request.Body.Position = 0;

            return CallFunction();
        }

        [HttpGet]
        public object Get()
        {
            return CallFunction();
        }

        private object CallFunction()
        {
            if (!_function.IsCompiled())
                _compiler.Compile(_function);
            return _invoker.Execute(_function, Request);
        }
        
    }
}
