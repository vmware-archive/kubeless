using Kubeless.Functions;
using System.Threading;

namespace Kubeless.Core.Interfaces
{
    public interface IInvoker
    {
        object Execute(IFunction function, CancellationTokenSource cancellationSource, Event kubelessEvent, Context kubelessContext);
    }
}
