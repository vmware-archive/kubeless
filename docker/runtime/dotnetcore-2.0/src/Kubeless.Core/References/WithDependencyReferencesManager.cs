using Kubeless.Core.Interfaces;
using Microsoft.CodeAnalysis;
using System.Collections.Generic;
using System.IO;
using System.Linq;

namespace Kubeless.Core.References
{
    public class WithDependencyReferencesManager : IReferencesManager
    {
        private IEnumerable<IReferencesManager> basicReferencesManager;
        private IEnumerable<IReferencesManager> dependenciesManager;

        public WithDependencyReferencesManager()
        {
            basicReferencesManager = new List<IReferencesManager>()
            {
                new SharedReferencesManager(),
                new StoreReferencesManager()
            };

            dependenciesManager = new List<IReferencesManager>()
            {
                new KubelessReferencesManager()
            };
        }

        public MetadataReference[] GetReferences()
        {
            var references = new List<MetadataReference>();

            // Add all native assemblies 
            foreach (var manager in basicReferencesManager)
                references.AddRange(manager.GetReferences());

            // Add external referenced assemblies, but check if them aren't a native one
            foreach (var manager in dependenciesManager)
                foreach (var reference in manager.GetReferences())
                {
                    var assemblyName = Path.GetFileNameWithoutExtension(reference.Display);
                    if (references.FirstOrDefault(r => r.Display.Contains(assemblyName)) == null)
                        references.Add(reference);
                }

            return references.ToArray();
        }
    }
}
