#!/usr/bin/env node

const http = require('http')
const path = require('path')

const modName = process.env.MOD_NAME
const funcHandler = process.env.FUNC_HANDLER

const modRootPath = process.env.MOD_ROOT_PATH ? process.env.MOD_ROOT_PATH : '/kubeless/'
const modPath = path.join(modRootPath, modName + '.js')
console.log('Loading', modPath)
try {
  var mod = require(modPath)
} catch (e) {
  console.error('No valid module found for the name: lambda, Failed to import module')
  process.exit(1)
}

console.log('mod[funcHandler]', funcHandler, mod[funcHandler])
const server = http.createServer((req, res) => {
  return mod[funcHandler](req, res)
})

server.listen(8080, '0.0.0.0')
