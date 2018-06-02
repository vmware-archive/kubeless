using Kubeless.Core.Interfaces;
using System.IO;

namespace Kubeless.Core.Models
{
    public class BinaryContent : IFileContent<byte[]>
    {
        public string FilePath { get; }
        public byte[] Content { get; private set; }
        public bool Exists { get; private set; }

        public BinaryContent(string filePath)
        {
            this.FilePath = filePath;
            if (File.Exists(filePath))
            {
                this.Content = File.ReadAllBytes(filePath);
                Exists = true;
            }
            else
            {
                Exists = false;
            }
        }
    }
}
