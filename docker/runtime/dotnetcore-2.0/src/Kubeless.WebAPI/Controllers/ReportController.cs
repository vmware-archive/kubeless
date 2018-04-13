using Kubeless.Core.Interfaces;
using Kubeless.WebAPI.Utils;
using Microsoft.AspNetCore.Mvc;
using Microsoft.Extensions.Configuration;
using System;
using System.Collections.Generic;
using System.Linq;
using System.Threading.Tasks;

namespace Kubeless.WebAPI.Controllers
{
    [Route("/report")]
    public class ReportController : Controller
    {
        private IFunction _function;
        private IConfiguration _configuration;

        public ReportController(IFunction function, IConfiguration configuration)
        {
            _function = function;
            _configuration = configuration;
        }

        [HttpGet]
        public string FunctionReport()
        {
            return new ReportBuilder(_function.FunctionSettings, _configuration).GetReport();
        }
    }
}
