using System;
using System.Collections.Generic;
using System.Text;

namespace Kubeless.Tests.Utils
{
    public class FunctionEnvironment
    {
        public string Path { get; set; }
        public string FunctionFileName { get; set; }

        public FunctionEnvironment(string path, string functionFileName)
        {
            Path = path;
            FunctionFileName = functionFileName;
        }

        public string PackagesPath
            => $@"{Path}\Packages";

        public string FunctionFile
            => $@"{Path}\{FunctionFileName}.cs";

        public string AssemblyFile
            => $@"{Path}\{FunctionFileName}.dll";

        public string ProjectFile
            => $@"{Path}\{FunctionFileName}.csproj";
    }
}
