using System;
using System.Collections.Generic;
using System.Linq;
using System.Threading.Tasks;

namespace Kubeless.WebAPI.Utils
{
    public static class VariablesUtils
    {

        public static string GetEnvironmentVariable(string environmentVariable, string defaultValue)
        {
            var variable = Environment.GetEnvironmentVariable(environmentVariable);
            return string.IsNullOrWhiteSpace(variable) ? defaultValue : variable;
        }

    }
}
