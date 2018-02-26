<?php

function foo($event, $context) {
  return json_encode($event->data);
}

