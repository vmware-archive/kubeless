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

build5.26:
	docker build -t kubeless/perl:5.26$$RUNTIME_TAG_MODIFIER -f Dockerfile.5.26 .

push5.26:
	docker push kubeless/perl:5.26$$RUNTIME_TAG_MODIFIER

# Mandatory jobs
build-all: build5.26
push-all: push5.26
