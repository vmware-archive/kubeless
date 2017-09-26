namespace Kubeless.Core.Interfaces
{
    public interface IFunction
    {
        IFunctionSettings FunctionSettings { get; }

        bool IsCompiled();
    }
}
