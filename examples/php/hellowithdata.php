<?php

function foo($request) {
   /** @var \Psr\Http\Message\ServerRequestInterface $request */
  $body = $request->getBody()->getContents();
  print $body;
}

