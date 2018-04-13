using Kubeless.Core.Interfaces;
using System;
using System.IO;
using System.Reflection;

namespace Kubeless.Core.Models
{
    public class DefaultInvoker : IInvoker
    {

        public DefaultInvoker()
        {
            Assembly.LoadFrom(@"C:\Users\altargin\Desktop\packages\yamldotnet\4.3.1\lib\netstandard1.3\YamlDotNet.dll");

        }
        public object Execute(IFunction function, params object[] parameters)
        {
            AppDomain domain = AppDomain.CurrentDomain;


            var assembly = Assembly.Load(function.FunctionSettings.Assembly.Content);

            var ass = domain.GetAssemblies();

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
