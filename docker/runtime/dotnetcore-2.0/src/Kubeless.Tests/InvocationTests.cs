using Kubeless.Core.Compilers;
using Kubeless.Core.Invokers;
using Kubeless.Core.References;
using Kubeless.Tests.Utils;
using Microsoft.AspNetCore.Http;
using Microsoft.AspNetCore.Http.Internal;
using Xunit;
[assembly: CollectionBehavior(DisableTestParallelization = true)]

namespace Kubeless.Tests
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

            var args = GetHttpRequest();

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

            var args = GetHttpRequest();

            object result = invoker.Execute(function, args);
        }

        private HttpRequest GetHttpRequest()
           => new DefaultHttpContext().Request;
    }
}
