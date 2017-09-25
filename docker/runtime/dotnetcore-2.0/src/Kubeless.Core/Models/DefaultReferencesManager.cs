using Kubeless.Core.Interfaces;
using System;
using System.Collections.Generic;
using Microsoft.CodeAnalysis;

namespace Kubeless.Core.Models
{
    public class DefaultReferencesManager : IReferencesManager
    {
        private IEnumerable<IReferencesManager> referencesManager;

        public DefaultReferencesManager()
        {
            referencesManager = new List<IReferencesManager>()
            {
                new SharedReferencesManager(),
                new StoreReferencesManager()
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
