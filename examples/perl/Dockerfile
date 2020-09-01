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
# Create a custom image with a perl function
FROM stanguturi/perl-kubeless:5.26
ENV FUNC_HANDLER=hello \
    MOD_NAME=helloget
ADD helloget.pm /
RUN mkdir -p /kubeless/
RUN chown 1000:1000 /kubeless
ENTRYPOINT [ "bash", "-c", "cp /helloget.pm /kubeless/ && perl /kubeless.pl"]
