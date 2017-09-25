using Kubeless.Core.Interfaces;
using System;
using System.Reflection;

namespace Kubeless.Core.Models
{
    public class DefaultInvoker : IInvoker
    {
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
    }
}
