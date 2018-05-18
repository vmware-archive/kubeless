using Microsoft.AspNetCore.Http;
using System;
using System.Collections.Generic;
using System.Text;

namespace Kubeless.Core.Tests.Utils
{
    public static class WebManager
    {
        public static HttpRequest GetHttpRequest()
            => new DefaultHttpContext().Request;
    }
}
