namespace Kubeless.Core.Interfaces
{
    public interface IFileContent<T>
    {
        string FilePath { get; }
        T Content{ get; }
    }
}