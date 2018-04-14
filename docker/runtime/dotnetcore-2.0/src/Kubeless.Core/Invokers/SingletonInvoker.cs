using Kubeless.Core.Interfaces;
using Kubeless.Core.Invoker;
using System;
using System.Collections.Generic;
using System.Reflection;
using System.Text;

namespace Kubeless.Core.Invokers
{
    public class SingletonInvoker : IInvoker
    {
        private readonly Type _type;
        private readonly object _instance;

        public SingletonInvoker(IFunction function)
        {
            var invocationManager = new CustomReferencesManager();
            var references = invocationManager.GetReferences();

            foreach (var r in references)
                Assembly.LoadFrom(r);

            var assembly = Assembly.Load(function.FunctionSettings.Assembly.Content);

            _type = assembly.GetType(function.FunctionSettings.ModuleName);

            _instance = Activator.CreateInstance(_type);
        }

        public object Execute(IFunction function, params object[] parameters)
        {
            var returnedValue = _type.InvokeMember(function.FunctionSettings.FunctionHandler,
                                     BindingFlags.Default | BindingFlags.InvokeMethod,
                                     null,
                                     _instance,
                                     parameters);

            return returnedValue;
        }
    }
}
