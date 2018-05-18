using System;

namespace Kubeless.Functions
{
    public class Extensions
    {
        public object HttpRequest { get; }

        public Extensions(object httpRequest)
        {
            HttpRequest = httpRequest ?? throw new ArgumentNullException(nameof(httpRequest));
        }
    }
}
