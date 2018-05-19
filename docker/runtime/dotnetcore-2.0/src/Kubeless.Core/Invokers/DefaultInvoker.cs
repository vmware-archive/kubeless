using Kubeless.Core.Interfaces;
using Kubeless.Core.Invoker;
using Kubeless.Functions;
using System;
using System.IO;
using System.Reflection;
using System.Threading;

namespace Kubeless.Core.Invokers
{
    public class DefaultInvoker : IInvoker
    {
        public DefaultInvoker()
        {
            var invocationManager = new CustomReferencesManager();
            var references = invocationManager.GetReferences();

            foreach (var r in references)
                Assembly.LoadFrom(r);
        }

        public object Execute(IFunction function, params object[] parameters)
        {
            var assembly = Assembly.Load(function.FunctionSettings.Assembly.Content);

            Type type = assembly.GetType(function.FunctionSettings.ModuleName);

            object instance = Activator.CreateInstance(type);

            var returnedValue = type.InvokeMember(function.FunctionSettings.FunctionHandler,
                                     BindingFlags.Default | BindingFlags.InvokeMethod,
                                     null,
                                     instance,
                                     parameters);

            return returnedValue;
        }

        public object Execute(IFunction function, CancellationTokenSource cancellationSource, Event kubelessEvent, Context kubelessContext)
        {
            throw new NotImplementedException();
        }
    }
}
