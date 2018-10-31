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

use strict;

sub print_hash {
    my $href = shift;
    my $result = "{ ";
    my @key_values = ();
    while( my( $key, $val ) = each %{$href} ) {
        push @key_values, "'$key' => '$val'";
    }
    $result = "{ " . (join ',', @key_values) . " }";
    print($result . "\n");
    return $result;
}

sub hello_data {
   my $event = shift;
   my $context = shift;
   return("Data : " . print_hash($event->{data}) . " \n");
}

1
