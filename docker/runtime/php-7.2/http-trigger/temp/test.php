<?php

require 'vendor/autoload.php';

use Monolog\Logger;
use Monolog\Handler\StreamHandler;

function foo() {
  sleep(5);
  print "ciao";
}