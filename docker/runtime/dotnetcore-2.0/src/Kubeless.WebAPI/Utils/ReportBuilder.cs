using Kubeless.Core.Interfaces;
using Kubeless.Core.Models;
using Microsoft.Extensions.Configuration;
using System;
using System.Text;

namespace Kubeless.WebAPI.Utils
{
    public class ReportBuilder
    {
        private IFunctionSettings _functionSettings;
        private IConfiguration _configuration;

        public ReportBuilder(IFunctionSettings functionSettings, IConfiguration configuration)
        {
            _functionSettings = functionSettings;
            _configuration = configuration;
        }

        public string GetReport()
        {
            var builder = new StringBuilder();

            BuildHeader(builder);
            BuildFunctionReport(builder);
            BuildConfigurationReport(builder);
            BuildFooter(builder);

            return builder.ToString();
        }


        private void BuildHeader(StringBuilder builder)
        {
            builder.AppendLine("*** .NET Core Kubeless Runtime - Report ***");
            builder.AppendLine("--------------------------------------------------");
            builder.AppendLine();
        }

        private void BuildFunctionReport(StringBuilder builder)
        {
            builder.AppendKeyValue("Module/Class name", _functionSettings.ModuleName);
            builder.AppendKeyValue("Function Handler name", _functionSettings.FunctionHandler);
            builder.AppendKeyValue("Code file path", _functionSettings.Code.FilePath);
            builder.AppendCode("Code content", _functionSettings.Code.Content);
            builder.AppendKeyValue("Requirements file path", _functionSettings.Requirements.FilePath);
            builder.AppendCode("Requirements content", _functionSettings.Requirements.Content);
            builder.AppendKeyValue("Assembly file path", _functionSettings.Assembly.FilePath);
            builder.AppendKeyValue("Assembly exists", ((BinaryContent)_functionSettings.Assembly).Exists.ToString());
        }

        private void BuildConfigurationReport(StringBuilder builder)
        {
            builder.AppendKeyValue("CodePath", _configuration["Compiler:CodePath"]);
            builder.AppendKeyValue("RequirementsPath", _configuration["Compiler:RequirementsPath"]);
            builder.AppendKeyValue("FunctionAssemblyPath", _configuration["Compiler:FunctionAssemblyPath"]);
        }

        private void BuildFooter(StringBuilder builder)
        {
            builder.AppendLine();
            builder.AppendLine("--------------------------------------------------");
            builder.AppendLine(DateTime.Now.ToString());
        }

    }
}