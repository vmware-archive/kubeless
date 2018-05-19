using Kubeless.Core.Interfaces;
using Microsoft.CodeAnalysis;
using System;
using System.Collections.Generic;
using System.IO;
using System.Linq;
using System.Reflection;
using System.Runtime.InteropServices;
using System.Text.RegularExpressions;

namespace Kubeless.Core.References
{
    internal class SharedReferencesManager : IReferencesManager
    {
        private static readonly string BaseSharedPath = RuntimeInformation.IsOSPlatform(OSPlatform.Windows) ?
            Path.Combine(Environment.GetEnvironmentVariable("ProgramFiles"), $@"dotnet\shared\Microsoft.NETCore.App\") :
            Path.Combine("/usr/share", $@"dotnet/shared/Microsoft.NETCore.App/");

        private static readonly string SharedPath = Path.Combine(BaseSharedPath, GetInstalledNetCoreVersion());

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

        private static string GetInstalledNetCoreVersion()
        {
            var versionPattern = new Regex(@"(\d\.\d\.\d)");
            var dirs = Directory.GetDirectories(BaseSharedPath, "*", SearchOption.TopDirectoryOnly);
            var dirNames = from d in dirs select versionPattern.Match(d).Groups[1].Value;
            return dirNames.OrderByDescending(d => d).First();
        }
    }
}
