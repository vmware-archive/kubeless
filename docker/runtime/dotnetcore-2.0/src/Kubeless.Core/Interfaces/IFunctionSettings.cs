namespace Kubeless.Core.Interfaces
{
    public interface IFunctionSettings
    {
        string ModuleName { get; }
        string FunctionHandler { get; }
        IFileContent<byte[]> Assembly { get; }
    }
}