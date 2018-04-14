using System;
using System.Collections.Generic;
using System.IO;
using System.Text;

namespace Kubeless.Tests.Utils
{
    public static class EnvironmentManager
    {
        public static void InitVariables()
        {
            //Environment.SetEnvironmentVariable("MOD_NAME", "mycode");
            //Environment.SetEnvironmentVariable("FUNC_HANDLER", "execute");
            Environment.SetEnvironmentVariable("DOTNETCORE_HOME", @".\packages");
            Environment.SetEnvironmentVariable("DOTNETCORESHAREDREF_VERSION", "2.0.6"); //TODO: Get Higher available version
        }

        public static void ClearCompilations(string basePath)
        {
            if (!Directory.Exists(basePath))
                return;

            var compilations = Directory.EnumerateFiles(basePath, "*.dll", SearchOption.AllDirectories);

            foreach (var c in compilations)
                File.Delete(c);
        }
    }
}
