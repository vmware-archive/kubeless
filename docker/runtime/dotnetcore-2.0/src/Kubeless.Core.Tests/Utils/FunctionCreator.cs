using Kubeless.Core.Interfaces;
using Kubeless.Core.Models;
using System;
using System.Collections.Generic;
using System.IO;
using System.Text;

namespace Kubeless.Core.Tests.Utils
{
    public static class FunctionCreator
    {
        private static IFunctionSettings BuildFunctionSettings(string functionFile, string moduleName, string functionHandler, string requirementsFile = "")
        {
            var basePath = Path.GetDirectoryName(functionFile);
            var baseName = Path.GetFileNameWithoutExtension(functionFile);

            var assembly = new BinaryContent(Path.Combine(basePath, $"{baseName}.dll"));

            return new FunctionSettings(moduleName, functionHandler, assembly);
        }

        public static IFunction CreateFunction(string functionFile, string requirementsFile = "", string moduleName = "module", string functionHandler = "handler")
        {
            var settings = BuildFunctionSettings(functionFile, moduleName, functionHandler, requirementsFile);
            return new Function(settings);
        }
    }
}
