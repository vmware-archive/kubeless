<?php

require 'vendor/autoload.php';

use Monolog\Logger;
use Monolog\Handler\StreamHandler;

function foo() {
  // create a log channel
  $log = new Logger('name');
  $log->pushHandler(new StreamHandler("php://stdout", Logger::INFO));

  // add records to the log
  $log->info('Hello');
  $log->info('World');
  print "hello world";
}