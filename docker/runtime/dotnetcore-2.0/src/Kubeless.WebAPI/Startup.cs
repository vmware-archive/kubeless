using System;
using System.Collections.Generic;
using System.Linq;
using System.Threading.Tasks;
using Microsoft.AspNetCore.Builder;
using Microsoft.AspNetCore.Hosting;
using Microsoft.Extensions.Configuration;
using Microsoft.Extensions.DependencyInjection;
using Microsoft.Extensions.Logging;
using Microsoft.Extensions.Options;
using Kubeless.Core.Interfaces;
using Kubeless.WebAPI.Utils;
using Kubeless.Core.Models;

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
            }
        }

        public IConfiguration Configuration { get; }

        // This method gets called by the runtime. Use this method to add services to the container.
        public void ConfigureServices(IServiceCollection services)
        {
            services.AddMvc();

            services.AddSingleton<IFunction>(FunctionFactory.BuildFunction(Configuration));
            services.AddSingleton<ICompiler>(new DefaultCompiler(new DefaultParser(), new DefaultReferencesManager()));
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
