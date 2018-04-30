using System;
using System.Collections.Generic;
using System.Text;

namespace Kubeless.Core.Tests.Utils
{
    public class FunctionEnvironment
    {
        private const int DEFAULT_TIMEOUT = 180;

        public string Path { get; }
        public string FunctionFileName { get; }
        public int Timeout { get; }

        public FunctionEnvironment(string path, string functionFileName) : this(path, functionFileName, DEFAULT_TIMEOUT) { }

        public FunctionEnvironment(string path, string functionFileName, int timeout)
        {
            Path = path;
            FunctionFileName = functionFileName;
            Timeout = timeout;
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
