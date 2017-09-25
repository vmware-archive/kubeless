using Kubeless.Core.Interfaces;
using System.IO;

namespace Kubeless.Core.Models
{
    public class StringContent : IFileContent<string>
    {
        public string FilePath { get; }
        public string Content { get; }

        public StringContent(string filePath)
        {
            this.FilePath = filePath;
            if (File.Exists(filePath))
                this.Content = File.ReadAllText(filePath);
        }
    }
}
