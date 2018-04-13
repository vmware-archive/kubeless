using Kubeless.Core.Filters;
using Kubeless.Core.Interfaces;
using Microsoft.CodeAnalysis;
using System;
using System.Collections.Generic;
using System.IO;
using System.Linq;
using System.Reflection;
using System.Text;

namespace Kubeless.Core.Models
{
    public class KubelessReferencesManager : IReferencesManager
    {
        private static string directory = Environment.GetEnvironmentVariable("DOTNETCORE_HOME");

        public MetadataReference[] GetReferences()
        {
            if (Directory.Exists(directory))
            {
                var dlls = Directory
                    .EnumerateFiles(directory, "*.dll", SearchOption.AllDirectories)
                    .ApplyFilterForNetStandard();

                var dllFiles = from d in dlls select new FileInfo(d);

                var references = new List<MetadataReference>();

                foreach (var dll in dlls)
                {
                    try
                    {
                        var assembly = Assembly.LoadFile(dll);
                        references.Add(MetadataReference.CreateFromFile(dll));
                    }
                    catch (BadImageFormatException ex)
                    {

                    }
                    catch
                    {
                        throw;
                    }
                }

                return references.ToArray();
            }

            return null;
        }
    }
}
