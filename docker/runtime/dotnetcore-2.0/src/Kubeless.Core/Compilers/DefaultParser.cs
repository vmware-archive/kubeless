using Kubeless.Core.Interfaces;
using System;
using System.Collections.Generic;
using System.Text;
using Microsoft.CodeAnalysis;
using Microsoft.CodeAnalysis.CSharp;

namespace Kubeless.Core.Compilers
{
    public class DefaultParser : IParser
    {
        public SyntaxTree ParseText(string code)
        {
            return CSharpSyntaxTree.ParseText(code);
        }
    }
}
