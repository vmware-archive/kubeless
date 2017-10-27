#!/bin/bash
init() {
  local dir=`mktemp -u`
  echo `git init $dir` 
}
