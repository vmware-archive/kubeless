FROM microsoft/aspnetcore:2.0.8
WORKDIR /app
COPY ./publish ./
ENTRYPOINT ["dotnet", "Kubeless.CSharpX.dll"]
