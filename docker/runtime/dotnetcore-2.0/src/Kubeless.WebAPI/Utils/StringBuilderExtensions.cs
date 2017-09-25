using System.Text;

namespace Kubeless.WebAPI.Utils
{
    public static class StringBuilderExtensions
    {
        public static void AppendKeyValue(this StringBuilder builder, string key, string value)
        {
            builder.AppendLine($"# {key}: {value}");
        }

        public static void AppendCode(this StringBuilder builder, string key, string code)
        {
            builder.AppendLine($"# {key}:");
            builder.AppendLine(code);
        }
    }
}
