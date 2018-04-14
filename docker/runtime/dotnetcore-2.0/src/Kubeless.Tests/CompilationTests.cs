using Kubeless.Core.Interfaces;
using Kubeless.Core.Models;
using Kubeless.Core.References;
using Kubeless.Tests.Utils;
using System;
using System.IO;
using Xunit;

namespace Kubeless.Tests
{
    public class CompilationTests
    {
        private const string BASE_PATH = @".\Functions";

        public CompilationTests()
        {
            EnvironmentManager.ClearCompilations(BASE_PATH);
            EnvironmentManager.InitVariables();
        }

        [Fact]
        public void BuildWithoutDependency()
        {
            var functionFileName = "nodependency";
            var functionFile = GetFunctionFile(functionFileName);
            var assemblyFile = GetAssemblyFile(functionFileName);

            var compiler = new DefaultCompiler(new DefaultParser(), new WithoutDependencyReferencesManager());
            var function = FunctionCreator.CreateFunction(functionFile);

            compiler.Compile(function);

            Assert.True(File.Exists(assemblyFile));
            Assert.NotEmpty(File.ReadAllBytes(assemblyFile));
        }

        [Fact]
        public void BuildWithDependencyNewtonsoft()
        {

        }

        private string GetFunctionFile(string functionFileName)
            => ($@"{BASE_PATH}\{functionFileName}.cs");

        private string GetAssemblyFile(string assemblyFileName)
            => ($@"{BASE_PATH}\{assemblyFileName}.dll");


    }
}
