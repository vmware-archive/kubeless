using Microsoft.CodeAnalysis;

namespace Kubeless.Core.Interfaces
{
    public interface IReferencesManager
    {
        MetadataReference[] GetReferences();
    }
}
