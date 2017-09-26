using Microsoft.CodeAnalysis;
using System;
using System.Collections.Generic;
using System.Text;
using System.Linq;

namespace Kubeless.Core.Exceptions
{
    [Serializable]
    public class CompilationException : Exception
    {

        private IEnumerable<Diagnostic> Errors { get; set; }

        public CompilationException() { }
        public CompilationException(string message) : base(message) { }
        public CompilationException(string message, Exception inner) : base(message, inner) { }
        protected CompilationException(
          System.Runtime.Serialization.SerializationInfo info,
          System.Runtime.Serialization.StreamingContext context) : base(info, context) { }

        public CompilationException(IEnumerable<Diagnostic> errors)
        {
            Errors = errors;
        }

        public override string Message
        {
            get
            {
                var builder = new StringBuilder();
                builder.AppendLine($"Compilation has failed.");
                builder.AppendLine($"Error quantity: {Errors.Count()}");
                foreach (var message in FormatErrors(Errors))
                    builder.AppendLine(message);
                return builder.ToString();
            }
        }

        private IEnumerable<string> FormatErrors(IEnumerable<Diagnostic> errors)
        {
            return from e in errors
                   select $"{e.Id}: {e.GetMessage()}";
        }
    }
}
