using Kubeless.Core.Compilers;
using Kubeless.Core.Handlers;
using Kubeless.Core.Interfaces;
using Kubeless.Core.Invokers;
using Kubeless.Core.References;
using Kubeless.Core.Tests.Utils;
using Kubeless.Functions;
using System;
using System.Threading;
using Xunit;
[assembly: CollectionBehavior(DisableTestParallelization = true)]

namespace Kubeless.Core.Tests
{
    public class InvocationTests
    {
        private const string BASE_PATH = @".\Functions";


        [InlineData("nodependency")]
        [Theory]
        public void BuildWithoutDependency(string functionFileName)
        {
            // Compile
            var environment = EnvironmentManager.CreateEnvironment(BASE_PATH, functionFileName);

            var functionFile = environment.FunctionFile;
            var assemblyFile = environment.AssemblyFile;

            var compiler = new DefaultCompiler(new DefaultParser(), new WithoutDependencyReferencesManager());
            var function = FunctionCreator.CreateFunction(functionFile);

            compiler.Compile(function);

            // Invoke
            var invoker = new DefaultInvoker();

            var args = WebManager.GetHttpRequest();

            object result = invoker.Execute(function, args);
        }

        //[InlineData("dependency-json")] //TODO: Run both tests in parallel
        [InlineData("dependency-yaml")]
        [Theory]
        public void BuildWithDependency(string functionFileName)
        {
            // Compile
            var environment = EnvironmentManager.CreateEnvironment(BASE_PATH, functionFileName);

            var functionFile = environment.FunctionFile;
            var projectFile = environment.ProjectFile;
            var assemblyFile = environment.AssemblyFile;

            var restorer = new DependencyRestorer(environment);
            restorer.CopyAndRestore();

            var compiler = new DefaultCompiler(new DefaultParser(), new WithDependencyReferencesManager());
            var function = FunctionCreator.CreateFunction(functionFile, projectFile);

            compiler.Compile(function);

            // Invoke
            var invoker = new DefaultInvoker();

            var args = WebManager.GetHttpRequest();

            object result = invoker.Execute(function, args);
        }

        #region Timeout

        [InlineData("timeout")]
        [Theory]
        public void RunWithTimeout(string functionFileName)
        {
            // Compile
            IFunction function = GetCompiledFunctionWithDepedencies(functionFileName);

            // Invoke
            var timeoutTriggered = false;
            try
            {
                var timeout = 4 * 1000; // Limits to 4 seconds
                object result = ExecuteCompiledFunction(function, timeout); // Takes 5 seconds
            }
            catch (Exception ex)
            {
                Assert.IsType<OperationCanceledException>(ex);
                timeoutTriggered = true;
            }

            Assert.True(timeoutTriggered);
        }

        [InlineData("timeout")]
        [Theory]
        public void RunWithoutTimeout(string functionFileName)
        {
            // Compile
            IFunction function = GetCompiledFunctionWithDepedencies(functionFileName);

            // Invoke
            var timeout = 6 * 1000; // Limits to 6 seconds
            object result = ExecuteCompiledFunction(function, timeout); // Takes 5 seconds

            Assert.Equal("This is a long run!", result.ToString());
        }

        #endregion

        private static object ExecuteCompiledFunction(IFunction function, int timeout = 180 * 1000)
        {
            // Invoke
            var invoker = new TimeoutInvoker(timeout);

            var cancellationSource = new CancellationTokenSource();

            var request = WebManager.GetHttpRequest();
            (Event _event, Context _context) = new DefaultParameterHandler().GetFunctionParameters(request);

            return invoker.Execute(function, cancellationSource, _event, _context); 
        }

        private static IFunction GetCompiledFunctionWithDepedencies(string functionFileName)
        {
            // Creates Environment
            var environment = EnvironmentManager.CreateEnvironment(BASE_PATH, functionFileName);

            var functionFile = environment.FunctionFile;
            var projectFile = environment.ProjectFile;
            var assemblyFile = environment.AssemblyFile;

            // Restore Dependencies
            var restorer = new DependencyRestorer(environment);
            restorer.CopyAndRestore();

            // Compile
            var compiler = new DefaultCompiler(new DefaultParser(), new WithDependencyReferencesManager());
            var function = FunctionCreator.CreateFunction(functionFile, projectFile);
            compiler.Compile(function);

            return function;
        }

    }
}
