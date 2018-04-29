using Kubeless.Core.Interfaces;
using Kubeless.Core.Invoker;
using Kubeless.Functions;
using System;
using System.Reflection;
using System.Threading;
using System.Threading.Tasks;
using System.Timers;

namespace Kubeless.Core.Invokers
{
    public class TimeoutInvoker : IInvoker
    {
        private readonly int _functionTimeout;

        public TimeoutInvoker(int functionTimeout)
        {
            _functionTimeout = functionTimeout;

            // Load Depedencies:
            LoadAssemblyDepedencies();
        }

        private static void LoadAssemblyDepedencies()
        {
            var invocationManager = new CustomReferencesManager();
            var references = invocationManager.GetReferences();

            foreach (var r in references)
                Assembly.LoadFrom(r);
        }

        private static object InvokeFunction(IFunction function, object[] parameters, Type type, object instance)
        {
            // Execute the function
            return type.InvokeMember(function.FunctionSettings.FunctionHandler,
                                     BindingFlags.Default | BindingFlags.InvokeMethod,
                                     null,
                                     instance,
                                     parameters);
        }

        public object Execute(IFunction function, CancellationTokenSource cancellationSource, Event kubelessEvent, Context kubelessContext)
        {
            // Gets references to function assembly
            var assembly = Assembly.Load(function.FunctionSettings.Assembly.Content);
            Type type = assembly.GetType(function.FunctionSettings.ModuleName);

            // Instantiates a new function
            object instance = Activator.CreateInstance(type);

            // Sets function timeout
            cancellationSource.CancelAfter(_functionTimeout);
            var cancellationToken = cancellationSource.Token;

            // Invoke function
            object functionOutput = null;
            var task = Task.Run(() =>
            {
                functionOutput = InvokeFunction(function, new object[] { kubelessEvent, kubelessContext }, type, instance);
            });

            // Wait for function execution. If the timeout is achived, the invoker exits
            task.Wait(cancellationToken);

            return functionOutput;
        }
    }
}
