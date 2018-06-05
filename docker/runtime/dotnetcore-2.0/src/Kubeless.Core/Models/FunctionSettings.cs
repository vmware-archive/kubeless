using Kubeless.Core.Interfaces;

namespace Kubeless.Core.Models
{
    public sealed class FunctionSettings : IFunctionSettings
    {
        public string ModuleName { get; private set; }
        public string FunctionHandler { get; private set; }

        public IFileContent<byte[]> Assembly { get; private set; }

        public FunctionSettings(string moduleName, string functionHandler, IFileContent<byte[]> assembly)
        {
            ModuleName = moduleName;
            FunctionHandler = functionHandler;
            Assembly = assembly;
        }

    }
}
