using Kubeless.Core.Interfaces;
using Microsoft.CodeAnalysis;
using System.Collections.Generic;
using System;
using System.Runtime.InteropServices;
using System.IO;
using System.Reflection;
using System.Linq;
using Kubeless.Core.Filters;


namespace Kubeless.Core.References
{
    internal class StoreReferencesManager : IReferencesManager
    {
        private static readonly string StorePath = RuntimeInformation.IsOSPlatform(OSPlatform.Windows) ?
            Path.Combine(Environment.GetEnvironmentVariable("ProgramFiles"), @"dotnet\store\x64\netcoreapp2.0\") :
            Path.Combine("/usr/share", @"dotnet/store/x64/netcoreapp2.0/");

        public MetadataReference[] GetReferences()
        {
            var dlls = Directory
                .EnumerateFiles(StorePath, "*.dll", SearchOption.AllDirectories)
                .ApplyFilterOnDllVersion();

            var dllFiles = from d in dlls select new FileInfo(d);

            var references = new List<MetadataReference>();

            //Not every .dll on the directory can be used during compilation. Some of them are just metadata.
            //The following try-catch statement ensures only usable assemblies will be added to compilation process.
            foreach (var dll in dlls)
            {
                try
                {
                    var assembly = Assembly.LoadFile(dll);
                    references.Add(MetadataReference.CreateFromFile(dll));
                }
                catch (BadImageFormatException) { }
                catch
                {
                    throw;
                }
            }

            return references.ToArray();
        }
    }
}
