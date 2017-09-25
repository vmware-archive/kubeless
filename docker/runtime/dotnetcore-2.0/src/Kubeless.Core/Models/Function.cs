using Kubeless.Core.Interfaces;
using System;
using System.Collections.Generic;
using System.IO;
using System.Linq;
using System.Threading.Tasks;

namespace Kubeless.Core.Models
{
    public class Function : IFunction
    {
        public IFunctionSettings FunctionSettings { get; }

        public Function(IFunctionSettings functionSettings)
        {
            FunctionSettings = functionSettings;
        }

        public bool IsCompiled()
        {
            return ((BinaryContent)FunctionSettings.Assembly).Exists;
        }
    }
}
