using System;
using System.IO;
using Kubeless.CSharpX;

string Hello( Event ev, Context ctx )
{
    Console.WriteLine( ev );
    return ev.Data;
}
