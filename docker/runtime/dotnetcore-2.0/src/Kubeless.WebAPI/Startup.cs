using System;
using Microsoft.AspNetCore.Builder;
using Microsoft.AspNetCore.Hosting;
using Microsoft.Extensions.Configuration;
using Microsoft.Extensions.DependencyInjection;
using Kubeless.Core.Interfaces;
using Kubeless.WebAPI.Utils;
using Kubeless.Core.Invokers;
using System.IO;
using Kubeless.Core.Handlers;

namespace kubeless_netcore_runtime
{
    public class Startup
    {
        public Startup(IConfiguration configuration, IHostingEnvironment env)
        {
            Configuration = configuration;

            if (env.IsDevelopment())
            {
                //Set fixed enviroment variables for example function:
                Environment.SetEnvironmentVariable("MOD_NAME", "hellowithdata");
                Environment.SetEnvironmentVariable("FUNC_HANDLER", "handler");
                Environment.SetEnvironmentVariable("FUNC_TIMEOUT", "180");
                Environment.SetEnvironmentVariable("FUNC_PORT", "8080");
                Environment.SetEnvironmentVariable("FUNC_RUNTIME", "dotnetcore2.0");
                Environment.SetEnvironmentVariable("FUNC_MEMORY_LIMIT", "0");
            }
        }

        public IConfiguration Configuration { get; }

        public void ConfigureServices(IServiceCollection services)
        {
            services.AddMvc();

            var function = FunctionFactory.BuildFunction(Configuration);

            if (!function.IsCompiled())
                throw new FileNotFoundException(nameof(function.FunctionSettings.ModuleName));

            services.AddSingleton<IFunction>(function);

            int timeout = int.Parse(VariablesUtils.GetEnvironmentVariable("FUNC_TIMEOUT", "180"));

            services.AddSingleton<IInvoker>(new CompiledFunctionInvoker(timeout * 1000)); // seconds
            services.AddSingleton<IParameterHandler>(new DefaultParameterHandler());
        }

        public void Configure(IApplicationBuilder app, IHostingEnvironment env)
        {
            if (env.IsDevelopment())
            {
                app.UseDeveloperExceptionPage();
            }

            app.UseCors(builder =>
                builder.
                AllowAnyHeader().
                AllowAnyOrigin().
                AllowAnyMethod()
                );

            app.UseMvc();
        }
    }
}
