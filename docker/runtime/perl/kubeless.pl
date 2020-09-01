#!/usr/bin/env perl
# Copyright 2019 VMware, Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#    http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

use threads;
use JSON::MaybeXS qw(encode_json decode_json);
use Dancer2;
use Time::HiRes qw(usleep);

# Port on which the dancer app will run.
my $func_port = $ENV{FUNC_PORT} || 8080;

# Timeout in seconds for the main function to finish.
my $timeout = int($ENV{FUNC_TIMEOUT} || 10);

# Module name and function handler.
# While deploying the function in kubeless,
# the module name is specified in --from-file command line argument and
# the function name is specified in --handler command line option.
my $mod_name = $ENV{MOD_NAME};
my $func_handler = $ENV{FUNC_HANDLER};

# This is the standard convention followed for kubeless runtimes.
# The main module file will be located in /kubeless/ directory.
# The runtime script needs to prepare the proper path and load it.
my $mod_path = "/kubeless/" . $mod_name . ".pm";

# Only for testing. Need to remove for production purposes.
if (!$mod_name) {
   $mod_path = $ENV{MOD_PATH};
}
# TODO: Known bug in Perl Dancer app. Need to set the port here.
set port => $func_port;

# TODO: Set the proper serializer app for JSON. But known bug in Dancer app.
# set serializer => 'JSON';

my %function_context = (
    'function' => $func_handler,
    'module' => $mod_path,
    'timeout' => $timeout,
    'runtime' => $ENV{'FUNC_RUNTIME'},
    'memory-limit' => $ENV{'FUNC_MEMORY_LIMIT'},
);

# It's always better to load the module in the beginning itself. If there
# are any errors, they will be caught early while the container is loading.
# The same format / style is being followed in other runtimes i.e. python,
# nodejs, etc.
eval {
   require($mod_path);
}; if ($@) {
   print("Error while invoking the module at $mod_path. Error: " . $@ . " \n");
   exit(1);
}

# Endpoint for /healthz. This is one of the main requirements by kubeless
# runtimes.
get '/healthz' => sub {
   return 'OK';
};

# Endpoint for /metrics. This is one of the main requirements by kubeless
# runtimes.
get '/metrics' => sub {
   # TODO: Needs to implement. We need to use proper prometheus client
   # package in perl. But it seems to be complicated. Lets leave this as is
   # for now.
   return '!!!!! To be Implemented !!!!!';
};

# Wrapper around the main function execution. This function will run the
# main function in a thread and waits for the timeout. If the function doesn'test
# complete in the timeout, then it is explicitly killed and a proper error
# is returned to the client.
sub thread_wrap {
   my $event_ref = shift;
   my $func_ref = shift;

   my $return;

   # If the function takes a lot of time to finish, the Main
   # thread sends a kill signal to terminate the function.
   # Registering a new function for the KILL signal is the only cleanest
   # way to provide the timeout functionality.
   $SIG{'KILL'} = sub { threads->exit(); };

   eval {
      my $function =  $func_ref->{function};
      my $func_result = &{\&{$function}}($event_ref, $func_ref);
      $return = $func_result;
   }; if ($@) {
      $return = "Error while invoking the function. Error: '" . $@ . "' \n";
   }

   return $return;
}

# Main route mapping for all the incoming HTTP requests from the client.
any ['get', 'post', 'patch', 'delete'] => qr{/.*} => sub {
    my $content_type = request->header('content-type');
    my $data = request->body();
    if ($content_type eq 'application/json') {
       eval {
          $data = decode_json($data);
       }; if ($@) {
          response->status(500);
          return "Unable to decode the json body. Error: '" . $@ . "' \n";
       }
    }

    # TODO: Implement extensions
    # my %extensions = ( 'request' => request );

    my %event = (
        'data' => $data,
        'event-id' =>  request->header('event-id'),
        'event-type' => request->header('event-type'),
        'event-time' => request->header('event-time'),
        'event-namespace' => request->header('event-namespace'),
    );

    my $thr = threads->create('thread_wrap', \%event, \%function_context);
    my $sleep_interval = 250000;
    my $timeout_usecs = $timeout * 1000000;
    for (my $i = 0; $i < $timeout_usecs ; $i = $i + $sleep_interval) {
       if ($thr->is_running()) {
          usleep($sleep_interval);
       } else {
          last;
       }
    }

    my $result;
    if ($thr->is_joinable()) {
        $result = $thr->join();
    } else {
        $thr->kill('KILL')->detach();
        $result = "**** Error: Task killed since it exceeded " .
                  $timeout . " seconds ... ***\n";
        response->status(500);
    }

    return "$result";
};

# Main entry point for the Dancer web application framework.
dance;
