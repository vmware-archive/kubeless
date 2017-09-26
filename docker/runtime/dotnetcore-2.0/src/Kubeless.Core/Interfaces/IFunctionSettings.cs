namespace Kubeless.Core.Interfaces
{
    public interface IFunctionSettings
    {
        string ModuleName { get; }
        string FunctionHandler { get; }
        IFileContent<string> Code { get; }
        IFileContent<string> Requirements { get; }
        IFileContent<byte[]> Assembly { get; }
    }
}