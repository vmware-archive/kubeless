using Kubeless.Core.Interfaces;
using Kubeless.Core.Models;
using Microsoft.Extensions.Configuration;
using System;

namespace Kubeless.WebAPI.Utils
{
    public class FunctionFactory
    {
        public static IFunctionSettings BuildFunctionSettings(IConfiguration configuration)
        {
            var moduleName = Environment.GetEnvironmentVariable("MOD_NAME");
            if (string.IsNullOrEmpty(moduleName))
                throw new ArgumentNullException("MOD_NAME");

            var functionHandler = Environment.GetEnvironmentVariable("FUNC_HANDLER");
            if (string.IsNullOrEmpty(moduleName))
                throw new ArgumentNullException("FUNC_HANDLER");

            var codePathSetting = configuration["Compiler:CodePath"];
            if (string.IsNullOrEmpty(codePathSetting))
                throw new ArgumentNullException("Compiler:CodePath");
            var codePath = string.Concat(codePathSetting, moduleName, ".cs");
            var code = new StringContent(codePath);

            var requirementsPathSetting = configuration["Compiler:RequirementsPath"];
            if (string.IsNullOrEmpty(requirementsPathSetting))
                throw new ArgumentNullException("Compiler:RequirementsPath");
            var requirementsPath = string.Concat(requirementsPathSetting, "requirements", ".xml");
            var requirements = new StringContent(requirementsPath);

            var assemblyPathConfiguration = configuration["Compiler:FunctionAssemblyPath"];
            if (string.IsNullOrEmpty(assemblyPathConfiguration))
                throw new ArgumentNullException("Compiler:FunctionAssemblyPath");
            var assemblyPath = string.Concat(assemblyPathConfiguration, moduleName, ".dll");
            var assembly = new BinaryContent(assemblyPath);

            return new FunctionSettings(moduleName, functionHandler, code, requirements, assembly);
        }

        public static IFunction BuildFunction(IConfiguration configuration)
        {
            var settings = BuildFunctionSettings(configuration);
            return new Function(settings);
        }
    }
}
