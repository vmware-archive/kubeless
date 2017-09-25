using Kubeless.Core.Interfaces;
using Microsoft.CodeAnalysis;
using System;
using System.Collections.Generic;
using System.IO;
using System.Reflection;
using System.Runtime.InteropServices;

namespace Kubeless.Core.Models
{
    public class SharedReferencesManager : IReferencesManager
    {
        private static readonly string SharedPath = RuntimeInformation.IsOSPlatform(OSPlatform.Windows) ?
            Path.Combine(Environment.GetEnvironmentVariable("ProgramFiles"), @"dotnet\shared\Microsoft.NETCore.App\2.0.0\") :
            Path.Combine("/usr/share", @"dotnet/shared/Microsoft.NETCore.App/2.0.0/");

        public MetadataReference[] GetReferences()
        {
            var dlls = Directory.EnumerateFiles(SharedPath, "*.dll");

            var references = new List<MetadataReference>()
            {
                MetadataReference.CreateFromFile(typeof(object).Assembly.Location)
            };

            //Not every .dll on the directory can be used during compilation. Some of them are just metadata.
            //The following try-catch statement ensures only usable assemblies will be added to compilation process.
            foreach (var dll in dlls)
            {
                try
                {
                    var assembly = Assembly.LoadFile(dll);
                    references.Add(MetadataReference.CreateFromFile(dll));
                }
                catch { }
            }

            return references.ToArray();
        }
    }
}
