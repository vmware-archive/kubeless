using Kubeless.Core.Interfaces;
using Microsoft.CodeAnalysis;
using System.Collections.Generic;

namespace Kubeless.Core.Models
{
    class NugetReferencesManager : IReferencesManager
    {
        public MetadataReference[] GetReferences()
        {
            var references = new List<MetadataReference>()
            {
            };
            return references.ToArray();
        }
    }
}
