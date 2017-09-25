namespace Kubeless.Core.Interfaces
{
    public interface ICompiler
    {
        IParser Parser { get; }
        IReferencesManager ReferenceManager { get; }

        void Compile(IFunction function);
    }
}
