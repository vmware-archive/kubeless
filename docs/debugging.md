# Debugging Kubeless

As a developer you'll probably be interested on the investigation of Kubeless code. A possible result of this investigation process is the proposition of a new feature or any additional contribution that could make any sense.

In this context, debugging tools raises as a fundamental part of this mentioned understanding process. This document will describe the process that developers must execute in order to be able to debug Kubeless code.

## 1. Delve

Delve is the component that allows you to debug Go code. This way, the first thing you need to do is install the solution in your computer.

You can find the procedure to install and configure Delve in you PC (Linux, Mac and Windows) following [this link](https://github.com/derekparker/delve/tree/master/Documentation/installation).

### Important Note

Some versions of Mac OS has been facing some troubles related to injection of auto-generated digital certificate required by Delve installation via Homebrew process. The error seems like that one presented below.

```console
==> Tapping go-delve/delve
Cloning into '/usr/local/Homebrew/Library/Taps/go-delve/homebrew-delve'...
remote: Counting objects: 7, done.
remote: Compressing objects: 100% (6/6), done.
remote: Total 7 (delta 0), reused 5 (delta 0), pack-reused 0
Unpacking objects: 100% (7/7), done.
Tapped 1 formula (33 files, 41.4KB)
==> Installing delve from go-delve/delve
==> Using the sandbox
==> Downloading https://github.com/derekparker/delve/archive/v1.0.0-rc.1.tar.gz
==> Downloading from https://codeload.github.com/derekparker/delve/tar.gz/v1.0.0-rc.1
######################################################################## 100.0%
security: SecKeychainSearchCopyNext: The specified item could not be found in the keychain.
==> Generating dlv-cert
==> openssl req -new -newkey rsa:2048 -x509 -days 3650 -nodes -config dlv-cert.cfg
-extensions codesign_reqext -batch -out dlv-cert.cer -keyout dlv-cert.key
==> [SUDO] Installing dlv-cert as root
==> sudo security add-trusted-cert -d -r trustRoot -k /Library/Keychains/System.keychain
dlv-cert.cer
Last 15 lines from /Users/gta/Library/Logs/Homebrew/delve/02.sudo:
2017-08-02 17:06:05 +0200

sudo
security
add-trusted-cert
-d
-r
trustRoot
-k
/Library/Keychains/System.keychain
dlv-cert.cer
```

This error commonly occurs because the installer wasn't able (for some reason) to auto-generate the required certificate for Delve installer.

You can manually fix the error installing the certificate by yourself. To do that please follow the steps described below.

**Unzip the delve-1.0.0-rc.1 file**

```console
$ tar  /Users/{you_user}/Library/Caches/Homebrew/delve-1.0.0-rc.1
```

**Navigate to Delve/Scripts directory**

```console
$ cd /Users/{your_user}/Library/Caches/Homebrew/delve-1.0.0-rc.1/scripts
```

**Execute gencert and provide your admin password**

```console
$ ./gencert.sh
```

Done. Now you can try to install Delve again via Homebrew. You'll see that the operation will be completed successfully.

## 2. Configure Visual Studio Code

In order to demonstrate the debug process I'll use Visual Studio Code. Visual Studio Code is a lightweight but powerful source code editor which runs on your desktop and is available for Windows, macOS and Linux. It comes with built-in support for JavaScript, TypeScript and Node.js and has a rich ecosystem of extensions for other languages (such as C++, C#, Python, PHP, Go) and runtimes (such as .NET and Unity).

To know more about VS Code, follow [this link](https://code.visualstudio.com/docs).

Microsoft already did a great job describing the process to configure Delve on top of VS Code. In order to accomplish that, please, follow [this link](https://github.com/Microsoft/vscode-go/wiki/Debugging-Go-code-using-VS-Code).

## 3. Debugging Kubeless

If you was successful VS Code debug setup task, you now have a new directory with one file called "launch.json" inside. This file must contain the follow content inside.

```json
{
	"version": "0.2.0",
	"configurations": [
		{
			"name": "Launch",
			"type": "go",
			"request": "launch",
			"mode": "debug",
			"remotePath": "",
			"port": 2345,
			"host": "127.0.0.1",
			"program": "${workspaceRoot}",
			"env": {},
			"args": [],
			"showLog": true
		}
	]
}
```

In order to debug a Go code, Delve looks for a "main" method, once that is the method that starts the entire execution flow. This way, could be a good practice replace the value of "program" property (currently "`${workspaceRoot}`") by the static path to the "main" file. In this case, the "program" property could be similar to this:

```json
 "program": "$/Users/{your_user}/Documents/Projects/.../kubeless/cmd/kubeless/"
```

Done. Now Kubeless code is done to be debugged.
