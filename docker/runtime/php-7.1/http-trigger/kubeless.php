<?php

namespace Kubeless;

use \Psr\Http\Message\ServerRequestInterface as Request;
use \Psr\Http\Message\ResponseInterface as Response;
use Symfony\Component\Process\PhpProcess;
use Symfony\Component\Process\Exception\RuntimeException as RuntimeException;

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
  private function runFunction()
  {
      include $this->file;
      if (!function_exists($this->function)) {
        throw new \Exception(sprintf("Function %s not exist", $this->function));
      }
      $php = file_get_contents($this->file);
      $php .= "\n{$this->function}();";
      $process = new PhpProcess($php);
      $process->setTimeout($this->timeout);
      $process->run();

      return $process->getOutput();
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
      $ret = $this->runFunction();
      $response->getBody()->write($ret);

      return $response;
    } catch (RuntimeException $e) {
      $response->getBody()->write($e->getMessage());
      $response->withStatus(408);

      return $response;
    } catch (\Exception $e) {
      $response->getBody()->write($e->getMessage());
      $response->withStatus(500);

      return $response;
    }
  }

  /**
   * Heltz route.
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
      $response->getBody()->write($e->getMessage());
      $response->withStatus(500);

      return $response;
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
      var_dump($e->getMessage()); die;
    }

  }
}

$server = new Kubeless\Controller();
$server->run();
