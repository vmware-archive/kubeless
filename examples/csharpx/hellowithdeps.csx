#r "/kubeless/packages/newtonsoft.json/11.0.2/lib/netstandard2.0/Newtonsoft.Json.dll"

using Newtonsoft.Json;
using Kubeless.CSharpX;

string Hello( Event ev, Context ctx )
{
	var account = new {
		Name = "John Doe",
		Email = "john@nuget.org",
	};

	return JsonConvert.SerializeObject( account, Formatting.Indented );
}
