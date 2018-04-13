using System;
using Microsoft.AspNetCore.Builder;
using Microsoft.AspNetCore.Hosting;
using Microsoft.Extensions.Configuration;
using Microsoft.Extensions.DependencyInjection;
using Kubeless.Core.Interfaces;
using Kubeless.WebAPI.Utils;
using Kubeless.Core.Models;
using System.Reflection;

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
                Environment.SetEnvironmentVariable("MOD_NAME", "mycode");
                Environment.SetEnvironmentVariable("FUNC_HANDLER", "execute");
                Environment.SetEnvironmentVariable("DOTNETCORE_HOME", @"C:\Users\altargin\Desktop\packages");
                Environment.SetEnvironmentVariable("DOTNETCORESHAREDREF_VERSION", "2.0.6"); //TODO: Get Higher available version
            }
            else
            {
                Environment.SetEnvironmentVariable("DOTNETCORESHAREDREF_VERSION", "2.0.0");
            }
        }

        public IConfiguration Configuration { get; }

        // This method gets called by the runtime. Use this method to add services to the container.
        public void ConfigureServices(IServiceCollection services)
        {
            services.AddMvc();

            //Compile Function.
            var function = FunctionFactory.BuildFunction(Configuration);
            var compiler = new DefaultCompiler(new DefaultParser(), new DefaultReferencesManager());

            if (!function.IsCompiled())
                compiler.Compile(function);

            services.AddSingleton<IFunction>(function);
            services.AddSingleton<IInvoker>(new DefaultInvoker());
        }

        // This method gets called by the runtime. Use this method to configure the HTTP request pipeline.
        public void Configure(IApplicationBuilder app, IHostingEnvironment env)
        {
            if (env.IsDevelopment())
            {
                app.UseDeveloperExceptionPage();
            }

            app.UseMvc();
        }
    }
}
