using Kubeless.Core.Interfaces;
using System;
using System.Collections.Generic;
using Microsoft.CodeAnalysis;

namespace Kubeless.Core.References
{
    public class WithDependencyReferencesManager : IReferencesManager
    {
        private IEnumerable<IReferencesManager> referencesManager;

        public WithDependencyReferencesManager()
        {
            referencesManager = new List<IReferencesManager>()
            {
                new SharedReferencesManager(),
                new StoreReferencesManager(),
                new KubelessReferencesManager()
            };
        }

        public MetadataReference[] GetReferences()
        {
            var references = new List<MetadataReference>();
            foreach (var manager in referencesManager)
                references.AddRange(manager.GetReferences());
            return references.ToArray();
        }
    }
}
