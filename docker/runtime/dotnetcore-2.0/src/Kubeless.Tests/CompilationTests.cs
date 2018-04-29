using Kubeless.Core.Compilers;
using Kubeless.Core.References;
using Kubeless.Core.Tests.Utils;
using System.IO;
using Xunit;

namespace Kubeless.Core.Tests
{
    public class CompilationTests
    {
        private const string BASE_PATH = @".\Functions";


        [InlineData("nodependency")]
        [Theory]
        public void BuildWithoutDependency(string functionFileName)
        {
            var environment = EnvironmentManager.CreateEnvironment(BASE_PATH, functionFileName);

            var functionFile = environment.FunctionFile;
            var assemblyFile = environment.AssemblyFile;

            Assert.True(File.Exists(functionFile));
            Assert.False(File.Exists(assemblyFile));

            var compiler = new DefaultCompiler(new DefaultParser(), new WithoutDependencyReferencesManager());
            var function = FunctionCreator.CreateFunction(functionFile);

            Assert.False(function.IsCompiled());

            compiler.Compile(function);

            Assert.True(function.IsCompiled());

            Assert.True(File.Exists(assemblyFile));
            Assert.NotEmpty(File.ReadAllBytes(assemblyFile));
        }

        //[InlineData("dependency-json")] //TODO: Run both tests in parallel
        [InlineData("dependency-yaml")]
        [Theory]
        public void BuildWithDependency(string functionFileName)
        {
            var environment = EnvironmentManager.CreateEnvironment(BASE_PATH, functionFileName);

            var functionFile = environment.FunctionFile;
            var projectFile = environment.ProjectFile;
            var assemblyFile = environment.AssemblyFile;

            Assert.True(File.Exists(functionFile));
            Assert.True(File.Exists(projectFile));
            Assert.False(File.Exists(assemblyFile));

            var restorer = new DependencyRestorer(environment);
            restorer.CopyAndRestore();

            var compiler = new DefaultCompiler(new DefaultParser(), new WithDependencyReferencesManager());
            var function = FunctionCreator.CreateFunction(functionFile, projectFile);

            Assert.False(function.IsCompiled());

            compiler.Compile(function);

            Assert.True(function.IsCompiled());

            Assert.True(File.Exists(assemblyFile));
            Assert.NotEmpty(File.ReadAllBytes(assemblyFile));
        }
    }
}
