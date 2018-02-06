<?php

namespace Kubeless;

use \Psr\Http\Message\ServerRequestInterface as Request;
use \Psr\Http\Message\ResponseInterface as Response;

require 'vendor/autoload.php';

class Controller
{
  private $app;
  private $timeout;
  private $root;
  private $function;

  public function __construct()
  {
    $this->app = new \Slim\App();
    $this->timeout = (!empty(getenv('FUNC_TIMEOUT')) ? getenv('FUNC_TIMEOUT') : 180);
    $this->root = (!empty(getenv('MOD_ROOT_PATH')) ? getenv('MOD_ROOT_PATH') : '/kubeless/');
    $this->file = sprintf("/kubeless/%s.php", getenv('MOD_NAME'));
    $this->function = getenv('FUNC_HANDLER');
  }

  /**
   * Run the injected function.
   *
   * @return void
   */
  private function runFunction(Request $request)
  {
      ob_start();
      $currentDir = getcwd();
      chdir($this->root);
      include $this->file;
      if (!function_exists($this->function)) {
        throw new \Exception(sprintf("Function %s not exist", $this->function));
      }
      call_user_func_array($this->function, [$request]);
      $response = ob_get_contents();
      ob_end_clean();
      chdir($currentDir);

      return $response;
  }

  /**
   * Validate some required variables.
   *
   * @return void
   */
  private function validate()
  {
    if (empty(getenv('FUNC_HANDLER'))) {
      throw new \Exception("FUNC_HANDLER is empty");
    }
    if (empty(getenv('MOD_NAME'))) {
      throw new \Exception("MOD_NAME is empty");
    }
    if (!file_exists($this->file)) {
      throw new \Exception(sprintf("%s cannot be found", $this->file));
    }
  }

  /**
   * Root route.
   *
   * @param Request $request
   * @param Response $response
   * @param array $args
   * @return Response $repsonse
   */
  public function root(Request $request, Response $response, array $args)
  {
    try {
      $this->validate();
      $ret = $this->runFunction($request);
      $response->getBody()->write($ret);

      return $response;
    } catch (\Exception $e) {
      $response->getBody()->write($e->getMessage() . "\n");
      $res = $response->withStatus(500);

      return $res;
    }
  }

  /**
   * Healtz route.
   *
   * @param Request $request
   * @param Response $response
   * @param array $args
   * @return Response $response
   */
  private function healtz(Request $request, Response $response, array $args)
  {
    try {
      $this->validate();
      $response->getBody()->write("OK");

      return $response;
    } catch (\Exception $e) {
      $response->getBody()->write($e->getMessage() . "\n");
      $res = $response->withStatus(500);

      return $res;
    }
  }

  /**
   * Run the slim framework.
   */
  public function run()
  {
    try {
      $this->app->get('/', [$this, 'root']);
      $this->app->get('/healtz', [$this, 'healtz']);
      $this->app->run();
    } catch (\Exception $e) {
      ob_end_flush();
      ob_start();
      print $e->getMessage();
      header($_SERVER['SERVER_PROTOCOL'] . ' 500 Internal Server Error', true, 500);
    }

  }
}

$server = new \Kubeless\Controller();
$server->run();
