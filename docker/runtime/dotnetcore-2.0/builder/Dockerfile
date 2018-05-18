FROM microsoft/aspnetcore-build:2.0.0

WORKDIR /app

COPY kubeless.csproj .

RUN dotnet restore

RUN mv /root/.dotnet /root/.nuget /

RUN chown 1000 /.dotnet/ /.nuget/