using Microsoft.AspNetCore.Http;
using Microsoft.CodeAnalysis;
using Microsoft.CodeAnalysis.CSharp;
using Microsoft.CodeAnalysis.Emit;
using System;
using System.Collections.Generic;
using System.IO;
using System.Linq;
using System.Net.Http;
using System.Reflection;
using System.Text;

namespace kubeless_netcore_runtime.Util
{
    public class Compiler
    {
        public string Code { get; }
        public string ClassName { get; }
        public string FunctionName { get; }
        private IEnumerable<Diagnostic> Failures { get; set; }
        private Assembly Assembly { get; set; }

        public Compiler(string code, string className, string functionName)
        {
            Code = code;
            ClassName = className;
            FunctionName = functionName;
        }

        #region compilation

        public bool Start()
        {
            var assemblyName = Path.GetRandomFileName();
            var syntaxTree = ParseText(Code);
            var references = GetReferences();
            var compilation = CreateCompilation(assemblyName, syntaxTree, references);
            return StartCompilation(compilation);
        }

        public SyntaxTree ParseText(string code)
        {
            return CSharpSyntaxTree.ParseText(code);
        }

        public MetadataReference[] GetReferences()
        {
            return new MetadataReference[]
                        {
                            MetadataReference.CreateFromFile(typeof(object).Assembly.Location),
                            MetadataReference.CreateFromFile(typeof(Enumerable).Assembly.Location),
                            MetadataReference.CreateFromFile(typeof(HttpClient).Assembly.Location),
                            MetadataReference.CreateFromFile(typeof(HttpRequest).Assembly.Location),
                        };
        }

        public CSharpCompilation CreateCompilation(string assemblyName, SyntaxTree syntaxTree, MetadataReference[] references)
        {
            return CSharpCompilation.Create(
                assemblyName,
                syntaxTrees: new[] { syntaxTree },
                references: references,
                options: new CSharpCompilationOptions(OutputKind.DynamicallyLinkedLibrary));
        }

        public bool StartCompilation(CSharpCompilation compilation)
        {
            using (var ms = new MemoryStream())
            {
                EmitResult result = compilation.Emit(ms);
                if (!result.Success)
                {
                    Failures = result.Diagnostics.Where(diagnostic =>
                                            diagnostic.IsWarningAsError ||
                                            diagnostic.Severity == DiagnosticSeverity.Error);
                }
                else
                {
                    ms.Seek(0, SeekOrigin.Begin);
                    Assembly = Assembly.Load(ms.ToArray());
                }

                return result.Success;
            }
        }

        #endregion

        #region output

        public object Execute(object[] arguments)
        {
            Type type = Assembly.GetType(ClassName);
            object obj = Activator.CreateInstance(type);
            var returnedValue = type.InvokeMember(FunctionName,
                                    BindingFlags.Default | BindingFlags.InvokeMethod,
                                    null,
                                    obj,
                                    arguments);

            return returnedValue;
        }

        public string GetErrors()
        {
            var builder = new StringBuilder("Errors during validation:\n");
            foreach (Diagnostic diagnostic in Failures)
                builder.AppendLine(string.Format("{0}: {1}", diagnostic.Id, diagnostic.GetMessage()));
            return builder.ToString();
        }

        #endregion

    }
}