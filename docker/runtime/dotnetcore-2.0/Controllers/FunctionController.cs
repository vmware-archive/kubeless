using System;
using Microsoft.AspNetCore.Http;
using Microsoft.AspNetCore.Mvc;
using kubeless_netcore_runtime.Util;

namespace kubeless_netcore_runtime.Controllers
{
    //[Produces("application/json")]
    [Route("/")]
    public class FunctionController : Controller
    {

        [HttpPost]
        public object Post([FromBody]object data)
        {
            try
            {
                return CompileAndExecuteFunction(data);
            }
            catch (Exception ex)
            {
                return ex.Message;
            }
        }

        [HttpGet]
        public string Get()
        {
            return "NotImplemented";
        }

        private object CompileAndExecuteFunction(object data)
        {
            var className = Environment.GetEnvironmentVariable("MOD_NAME");
            Console.WriteLine($"==> Class name: {className}");
            var functionName = Environment.GetEnvironmentVariable("FUNC_HANDLER");
            Console.WriteLine($"==> Function name: {functionName}");

            var codeFile = string.Concat("/kubeless/", className, ".cs");
            Console.WriteLine($"==> Code file: {codeFile}");
            var code = System.IO.File.ReadAllText(codeFile);
            Console.WriteLine($"==> Code:\n{code}");

            var compiler = new Compiler(code, className, functionName);

            //TODO: Install dependencies from project.json file
            //var dependencies = "/kubeless/project.json";

            var result = compiler.Start();
            if (result)
                return compiler.Execute(new object[] { HttpContext.Request });
            else
                return compiler.GetErrors();
        }

    }
}
