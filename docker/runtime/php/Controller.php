<?php

namespace Kubeless;

use \Psr\Http\Message\ServerRequestInterface as Request;
use \Psr\Http\Message\ResponseInterface as Response;

require 'vendor/autoload.php';
$_SERVER['SCRIPT_NAME'] = '/';

class TimeoutFunctionException extends \RuntimeException {}

class Controller
{
  private $app;
  private $timeout;
  private $root;
  private $function;
  private $currentDir;

  public function __construct()
  {

    $this->app = new \Slim\App();
    $this->timeout = (!empty(getenv('FUNC_TIMEOUT')) ? getenv('FUNC_TIMEOUT') : 180);
    $this->root = (!empty(getenv('MOD_ROOT_PATH')) ? getenv('MOD_ROOT_PATH') : '/kubeless/');
    $this->file = sprintf("/kubeless/%s.php", getenv('MOD_NAME'));
    $this->function = getenv('FUNC_HANDLER');
    $this->currentDir = getcwd();
    $this->functionContext = (object) array(
      'function-name' => $this->function,
      'timeout' => $this->timeout,
      'runtime' => getenv('FUNC_RUNTIME'),
      'memory-limit' => getenv('FUNC_MEMORY_LIMIT'),
    );
  }

  /**
   * Run the injected function.
   *
   * @return void
   */
  private function runFunction(Request $request, Response $response)
  {
      set_time_limit($this->timeout);
      ob_start();
      chdir($this->root);
      include $this->file;
      if (!function_exists($this->function)) {
        throw new \Exception(sprintf("Function %s not exist", $this->function));
      }
      $pid = pcntl_fork();
      if ($pid == 0) {
        $data = $request->getBody()->getContents();
        if ($_SERVER['HTTP_CONTENT_TYPE'] == 'application/json') {
          $data = json_decode($data);
        }
        $event = (object) array(
          'data' => $data,
          'event-type' => $_SERVER['HTTP_EVENT_TYPE'],
          'event-id' => $_SERVER['HTTP_EVENT_ID'],
          'event-time' => $_SERVER['HTTP_EVENT_TIME'],
          'event-namespace' => $_SERVER['HTTP_EVENT_NAMESPACE'],
          'extensions' => (object) array(
            'request' => $request,
            'response' => $response,
          )
        );
        $res = call_user_func($this->function, $event, $this->functionContext);
        $event->extensions->response->getBody()->write($res);
        ob_end_clean();
        chdir($this->currentDir);
        return $event->extensions->response;
      } else {
          sleep($this->timeout);
          posix_kill($pid, SIGKILL);
          throw new TimeoutFunctionException();
      }
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

      return $this->runFunction($request, $response);
    } catch (\Kubeless\TimeoutFunctionException $e) {
      $res = $response->withStatus(408);

      return $res;
    } catch (\Exception $e) {
      $response->getBody()->write($e->getMessage() . "\n");
      $res = $response->withStatus(500);

      return $res;
    }
  }

  /**
   * Healthz route.
   *
   * @param Request $request
   * @param Response $response
   * @param array $args
   * @return Response $response
   */
  public function healthz(Request $request, Response $response, array $args)
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
      $this->app->any('/', [$this, 'root']);
      $this->app->any('/healthz', [$this, 'healthz']);
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
