using Microsoft.CodeAnalysis;
using System;
using System.Collections.Generic;
using System.Text;

namespace Kubeless.Core.Interfaces
{
    public interface IParser
    {
        SyntaxTree ParseText(string code);
    }
}
