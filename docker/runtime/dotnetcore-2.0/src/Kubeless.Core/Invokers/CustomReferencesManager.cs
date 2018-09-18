using Kubeless.Core.Utils;
using System;
using System.Collections.Generic;
using System.IO;

namespace Kubeless.Core.Invoker
{
    public class CustomReferencesManager
    {
        private static readonly string directory = Environment.GetEnvironmentVariable("DOTNETCORE_HOME");

        public IEnumerable<string> GetReferences()
        {
            if (Directory.Exists(directory))
                return Directory
                    .EnumerateFiles(directory, "*.dll", SearchOption.AllDirectories)
                    .ApplyFilterForNetStandard();
            else
                return new List<string>();
        }
    }
}
