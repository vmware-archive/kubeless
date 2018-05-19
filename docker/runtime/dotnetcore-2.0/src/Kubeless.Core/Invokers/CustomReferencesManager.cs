using Kubeless.Core.Filters;
using System;
using System.Collections.Generic;
using System.IO;
using System.Linq;
using System.Text;

namespace Kubeless.Core.Invoker
{
    public class CustomReferencesManager
    {
        private static string directory = Environment.GetEnvironmentVariable("DOTNETCORE_HOME");

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
