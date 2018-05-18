using System;
using System.Collections.Generic;
using System.Diagnostics;
using System.IO;
using System.Text;

namespace Kubeless.Core.Tests.Utils
{
    /// <summary>
    /// Simulates dependency restore executed in an init container by Kubeless.
    /// </summary>
    public class DependencyRestorer
    {
        private readonly string _projectFile;
        private readonly string _packagesPath;
        private readonly string _copiedProjectFile;


        public DependencyRestorer(FunctionEnvironment environment)
        {
            _projectFile = environment.ProjectFile;
            _packagesPath = environment.PackagesPath;

            _copiedProjectFile = Path.Combine(_packagesPath, Path.GetFileName(_projectFile));
        }

        private void CopyProjectFile()
        {
            Directory.CreateDirectory(_packagesPath);

            var from = _projectFile;
            var to = _copiedProjectFile;

            File.Copy(from, to);
        }

        private void DeleteProjectFile()
        {
            File.Delete(_copiedProjectFile);
        }

        private void NugetRestore()
        {
            var process = new Process()
            {
                StartInfo = new ProcessStartInfo
                {
                    FileName = "dotnet",
                    Arguments = $"restore --packages .",
                    RedirectStandardOutput = true,
                    UseShellExecute = false,
                    CreateNoWindow = true,
                    WorkingDirectory = _packagesPath
                }
            };

            process.Start();

            string result = process.StandardOutput.ReadToEnd();

            process.WaitForExit();

            if (result.ToLower().Contains("error"))
                throw new Exception("Error during nuget restore.");
        }

        public void CopyAndRestore()
        {
            CopyProjectFile();
            NugetRestore();
            DeleteProjectFile();
        }
    }
}
