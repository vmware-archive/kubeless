using Kubeless.Functions;
using Microsoft.AspNetCore.Http;
using System;
using System.Collections.Generic;
using System.Text;

namespace Kubeless.Core.Interfaces
{
    public interface IParameterHandler
    {
        (Event, Context) GetFunctionParameters(HttpRequest request);
    }
}
