namespace Kubeless.Core.Interfaces
{
    public interface IInvoker
    {
        object Execute(IFunction function, params object[] parameters);
    }
}
