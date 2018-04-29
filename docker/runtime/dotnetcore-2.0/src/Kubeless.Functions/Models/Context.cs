using System;

namespace Kubeless.Functions
{
    public class Context
    {
        public string ModuleName { get; }
        public string FunctionName { get; }
        public string FunctionPort { get; }
        public string Timeout { get; }
        public string Runtime { get; }
        public string MemoryLimit { get; }

        public Context(string moduleName, string functionName, string functionPort, string timeout, string runtime, string memoryLimit)
        {
            ModuleName = moduleName ?? throw new ArgumentNullException(nameof(moduleName));
            FunctionName = functionName ?? throw new ArgumentNullException(nameof(functionName));
            FunctionPort = functionPort ?? throw new ArgumentNullException(nameof(functionPort));
            Timeout = timeout ?? throw new ArgumentNullException(nameof(timeout));
            Runtime = runtime ?? throw new ArgumentNullException(nameof(runtime));
            MemoryLimit = memoryLimit ?? throw new ArgumentNullException(nameof(memoryLimit));
        }
    }
}
