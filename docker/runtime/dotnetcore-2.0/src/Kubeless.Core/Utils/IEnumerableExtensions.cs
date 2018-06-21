using System;
using System.Collections.Generic;
using System.Text;
using System.Linq;
using System.IO;
using System.Text.RegularExpressions;

namespace Kubeless.Core.Utils
{
    public static class IEnumerableExtensions
    {
        private static Regex targetPattern = new Regex(@"lib\\(.+)\\.*\.dll");
        private static Regex versionPattern = new Regex(@"netstandard(\d\.\d)");

        public static IEnumerable<string> ApplyFilterForNetStandard(this IEnumerable<string> dllPaths)
        {
            var files = from dll in dllPaths select new FileInfo(dll);

            var netstandardAssembies = from f in files
                                       where f.FullName.Contains("netstandard")
                                       group f.FullName by f.Name into g
                                       select new { DLLName = g.Key, DLLList = MatchVersionedPattern(g.ToList()) };

            var output = new List<string>();
            foreach (var assembly in netstandardAssembies)
                output.Add(assembly.DLLList.OrderByDescending(d => d.Version).First().Assembly);

            return output;
        }

        private static IEnumerable<NetstandardAssembly> MatchVersionedPattern(IEnumerable<string> assemblyList)
        {
            return from assembly in assemblyList
                   select new NetstandardAssembly
                   {
                       Assembly = assembly,
                       FrameworkTarget = targetPattern.Match(assembly).Groups[1].Value,
                       Version = decimal.Parse(versionPattern.Match(assembly).Groups[1].Value)
                   };
        }


        public class NetstandardAssembly
        {
            public string Assembly { get; set; }
            public decimal Version { get; set; }
            public string FrameworkTarget { get; set; }
        }
    }
}
