using Kubeless.Core.Interfaces;
using Kubeless.Functions;
using System;
using System.Collections.Generic;
using System.Text;
using System.Threading;

namespace Kubeless.Core.Invokers
{
    public class ScriptedFunctionInvoker : IInvoker
    {
        public object Execute(IFunction function, CancellationTokenSource cancellationSource, Event kubelessEvent, Context kubelessContext)
        {
            throw new NotImplementedException();
        }
    }
}
